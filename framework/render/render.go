// Package render 渲染系统
// 借鉴Gin框架的渲染设计，支持多种响应格式
package render

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"html/template"
	"net/http"

	"github.com/cloudwego/hertz/pkg/app"
	"gopkg.in/yaml.v2"
)

// Render 渲染接口
type Render interface {
	Render(c *app.RequestContext) error
	WriteContentType(c *app.RequestContext)
}

// JSON JSON渲染器
type JSON struct {
	Data any
}

// IndentedJSON 带缩进的JSON渲染器
type IndentedJSON struct {
	Data any
}

// SecureJSON 安全的JSON渲染器（防止JSON劫持）
type SecureJSON struct {
	Prefix string
	Data   any
}

// JsonpJSON JSONP渲染器
type JsonpJSON struct {
	Callback string
	Data     any
}

// XML XML渲染器
type XML struct {
	Data any
}

// YAML YAML渲染器
type YAML struct {
	Data any
}

// String 字符串渲染器
type String struct {
	Format string
	Data   []any
}

// HTML HTML渲染器
type HTML struct {
	Template *template.Template
	Name     string
	Data     any
}

// HTMLDebug HTML调试渲染器
type HTMLDebug struct {
	Files  []string
	Glob   string
	Delims Delims
	Name   string
	Data   any
}

// Delims 模板分隔符
type Delims struct {
	Left  string
	Right string
}

// Redirect 重定向渲染器
type Redirect struct {
	Code     int
	Location string
}

// Data 原始数据渲染器
type Data struct {
	ContentType string
	Data        []byte
}

// Reader 流渲染器
type Reader struct {
	Headers       map[string]string
	ContentType   string
	ContentLength int64
	Reader        func(c *app.RequestContext)
}

// JSON渲染实现
func (r JSON) Render(c *app.RequestContext) error {
	r.WriteContentType(c)
	jsonBytes, err := json.Marshal(r.Data)
	if err != nil {
		return err
	}
	c.Write(jsonBytes)
	return nil
}

func (r JSON) WriteContentType(c *app.RequestContext) {
	writeContentType(c, []string{"application/json; charset=utf-8"})
}

// IndentedJSON渲染实现
func (r IndentedJSON) Render(c *app.RequestContext) error {
	r.WriteContentType(c)
	jsonBytes, err := json.MarshalIndent(r.Data, "", "    ")
	if err != nil {
		return err
	}
	c.Write(jsonBytes)
	return nil
}

func (r IndentedJSON) WriteContentType(c *app.RequestContext) {
	writeContentType(c, []string{"application/json; charset=utf-8"})
}

// SecureJSON渲染实现
func (r SecureJSON) Render(c *app.RequestContext) error {
	r.WriteContentType(c)
	jsonBytes, err := json.Marshal(r.Data)
	if err != nil {
		return err
	}
	if string(jsonBytes) == "null" && r.Prefix != "" {
		jsonBytes = []byte(r.Prefix + "null")
	}
	c.Write(jsonBytes)
	return nil
}

func (r SecureJSON) WriteContentType(c *app.RequestContext) {
	writeContentType(c, []string{"application/json; charset=utf-8"})
}

// JsonpJSON渲染实现
func (r JsonpJSON) Render(c *app.RequestContext) error {
	r.WriteContentType(c)
	jsonBytes, err := json.Marshal(r.Data)
	if err != nil {
		return err
	}

	if r.Callback == "" {
		c.Write(jsonBytes)
		return nil
	}

	c.WriteString(r.Callback)
	c.WriteString("(")
	c.Write(jsonBytes)
	c.WriteString(");")
	return nil
}

func (r JsonpJSON) WriteContentType(c *app.RequestContext) {
	writeContentType(c, []string{"application/javascript; charset=utf-8"})
}

// XML渲染实现
func (r XML) Render(c *app.RequestContext) error {
	r.WriteContentType(c)
	xmlBytes, err := xml.Marshal(r.Data)
	if err != nil {
		return err
	}
	c.Write(xmlBytes)
	return nil
}

func (r XML) WriteContentType(c *app.RequestContext) {
	writeContentType(c, []string{"application/xml; charset=utf-8"})
}

// YAML渲染实现
func (r YAML) Render(c *app.RequestContext) error {
	r.WriteContentType(c)
	yamlBytes, err := yaml.Marshal(r.Data)
	if err != nil {
		return err
	}
	c.Write(yamlBytes)
	return nil
}

func (r YAML) WriteContentType(c *app.RequestContext) {
	writeContentType(c, []string{"application/x-yaml; charset=utf-8"})
}

// String渲染实现
func (r String) Render(c *app.RequestContext) error {
	r.WriteContentType(c)
	if len(r.Data) > 0 {
		c.WriteString(fmt.Sprintf(r.Format, r.Data...))
	} else {
		c.WriteString(r.Format)
	}
	return nil
}

func (r String) WriteContentType(c *app.RequestContext) {
	writeContentType(c, []string{"text/plain; charset=utf-8"})
}

// HTML渲染实现
func (r HTML) Render(c *app.RequestContext) error {
	r.WriteContentType(c)
	if r.Name == "" {
		return r.Template.Execute(c, r.Data)
	}
	return r.Template.ExecuteTemplate(c, r.Name, r.Data)
}

func (r HTML) WriteContentType(c *app.RequestContext) {
	writeContentType(c, []string{"text/html; charset=utf-8"})
}

// HTMLDebug渲染实现
func (r HTMLDebug) Render(c *app.RequestContext) error {
	r.WriteContentType(c)

	var tmpl *template.Template
	var err error

	if len(r.Files) > 0 {
		tmpl, err = template.ParseFiles(r.Files...)
	} else if r.Glob != "" {
		tmpl, err = template.ParseGlob(r.Glob)
	} else {
		return fmt.Errorf("no template files or glob pattern specified")
	}

	if err != nil {
		return err
	}

	if r.Delims.Left != "" || r.Delims.Right != "" {
		tmpl.Delims(r.Delims.Left, r.Delims.Right)
	}

	if r.Name == "" {
		return tmpl.Execute(c, r.Data)
	}
	return tmpl.ExecuteTemplate(c, r.Name, r.Data)
}

func (r HTMLDebug) WriteContentType(c *app.RequestContext) {
	writeContentType(c, []string{"text/html; charset=utf-8"})
}

// Redirect渲染实现
func (r Redirect) Render(c *app.RequestContext) error {
	if (r.Code < http.StatusMultipleChoices || r.Code > http.StatusPermanentRedirect) && r.Code != http.StatusCreated {
		panic(fmt.Sprintf("Cannot redirect with status code %d", r.Code))
	}
	c.Redirect(r.Code, []byte(r.Location))
	return nil
}

func (r Redirect) WriteContentType(c *app.RequestContext) {}

// Data渲染实现
func (r Data) Render(c *app.RequestContext) error {
	r.WriteContentType(c)
	c.Write(r.Data)
	return nil
}

func (r Data) WriteContentType(c *app.RequestContext) {
	writeContentType(c, []string{r.ContentType})
}

// Reader渲染实现
func (r Reader) Render(c *app.RequestContext) error {
	r.WriteContentType(c)
	if r.ContentLength >= 0 {
		if r.Headers == nil {
			r.Headers = map[string]string{}
		}
		r.Headers["Content-Length"] = fmt.Sprintf("%d", r.ContentLength)
	}
	r.writeHeaders(c, r.Headers)
	if r.Reader != nil {
		r.Reader(c)
	}
	return nil
}

func (r Reader) WriteContentType(c *app.RequestContext) {
	writeContentType(c, []string{r.ContentType})
}

func (r Reader) writeHeaders(c *app.RequestContext, headers map[string]string) {
	for k, v := range headers {
		if len(c.Response.Header.Peek(k)) == 0 {
			c.Header(k, v)
		}
	}
}

// 辅助函数
func writeContentType(c *app.RequestContext, value []string) {
	header := c.Response.Header.Peek("Content-Type")
	if len(header) == 0 {
		c.Header("Content-Type", value[0])
	}
}

// 便捷函数
func WriteJSON(c *app.RequestContext, obj any) error {
	return JSON{Data: obj}.Render(c)
}

func WriteIndentedJSON(c *app.RequestContext, obj any) error {
	return IndentedJSON{Data: obj}.Render(c)
}

func WriteSecureJSON(c *app.RequestContext, prefix string, obj any) error {
	return SecureJSON{Prefix: prefix, Data: obj}.Render(c)
}

func WriteJsonpJSON(c *app.RequestContext, callback string, obj any) error {
	return JsonpJSON{Callback: callback, Data: obj}.Render(c)
}

func WriteXML(c *app.RequestContext, obj any) error {
	return XML{Data: obj}.Render(c)
}

func WriteYAML(c *app.RequestContext, obj any) error {
	return YAML{Data: obj}.Render(c)
}

func WriteString(c *app.RequestContext, format string, values ...any) error {
	return String{Format: format, Data: values}.Render(c)
}

func WriteHTML(c *app.RequestContext, tmpl *template.Template, name string, data any) error {
	return HTML{Template: tmpl, Name: name, Data: data}.Render(c)
}

func WriteData(c *app.RequestContext, contentType string, data []byte) error {
	return Data{ContentType: contentType, Data: data}.Render(c)
}
