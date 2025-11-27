// Package mybatis 慢查询告警通道实现
//
// 实现多种告警通道：
// 1. 日志告警通道
// 2. 邮件告警通道  
// 3. Webhook告警通道
// 4. 控制台告警通道
package mybatis

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/smtp"
	"os"
	"strings"
	"sync"
	"time"
)

// LogAlertChannel 日志告警通道
type LogAlertChannel struct {
	logger    *log.Logger
	enabled   bool
	logLevel  string
	formatter AlertFormatter
}

// EmailAlertChannel 邮件告警通道
type EmailAlertChannel struct {
	config    *EmailConfig
	enabled   bool
	template  *template.Template
	client    SMTPClient
	mutex     sync.Mutex
}

// WebhookAlertChannel Webhook告警通道
type WebhookAlertChannel struct {
	config     *WebhookConfig
	enabled    bool
	httpClient *http.Client
	retryCount int
	mutex      sync.Mutex
}

// ConsoleAlertChannel 控制台告警通道
type ConsoleAlertChannel struct {
	enabled   bool
	colorize  bool
	formatter AlertFormatter
}

// EmailConfig 邮件配置
type EmailConfig struct {
	SMTPHost     string   `yaml:"smtp_host" json:"smtp_host"`
	SMTPPort     int      `yaml:"smtp_port" json:"smtp_port"`
	Username     string   `yaml:"username" json:"username"`
	Password     string   `yaml:"password" json:"password"`
	From         string   `yaml:"from" json:"from"`
	To           []string `yaml:"to" json:"to"`
	Subject      string   `yaml:"subject" json:"subject"`
	TemplatePath string   `yaml:"template_path" json:"template_path"`
	TLS          bool     `yaml:"tls" json:"tls"`
}

// WebhookConfig Webhook配置
type WebhookConfig struct {
	URL         string            `yaml:"url" json:"url"`
	Method      string            `yaml:"method" json:"method"`
	Headers     map[string]string `yaml:"headers" json:"headers"`
	Timeout     time.Duration     `yaml:"timeout" json:"timeout"`
	RetryCount  int               `yaml:"retry_count" json:"retry_count"`
	RetryDelay  time.Duration     `yaml:"retry_delay" json:"retry_delay"`
	SecretToken string            `yaml:"secret_token" json:"secret_token"`
}

// AlertFormatter 告警格式化器接口
type AlertFormatter interface {
	Format(alert *AlertRecord) string
	GetFormat() string
}

// SMTPClient SMTP客户端接口
type SMTPClient interface {
	SendMail(addr string, a smtp.Auth, from string, to []string, msg []byte) error
}

// DefaultSMTPClient 默认SMTP客户端
type DefaultSMTPClient struct{}

// WebhookPayload Webhook负载
type WebhookPayload struct {
	Alert     *AlertRecord `json:"alert"`
	Timestamp string       `json:"timestamp"`
	Source    string       `json:"source"`
	Signature string       `json:"signature,omitempty"`
}

// DefaultAlertFormatter 默认告警格式化器
type DefaultAlertFormatter struct {
	format string
}

// JSONAlertFormatter JSON告警格式化器
type JSONAlertFormatter struct {
	indent bool
}

// TemplateAlertFormatter 模板告警格式化器
type TemplateAlertFormatter struct {
	template *template.Template
}

// 颜色常量
const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorPurple = "\033[35m"
	ColorCyan   = "\033[36m"
	ColorWhite  = "\033[37m"
)

// NewLogAlertChannel 创建日志告警通道
func NewLogAlertChannel(logFile string, logLevel string) *LogAlertChannel {
	var logger *log.Logger
	
	if logFile == "" || logFile == "stdout" {
		logger = log.New(os.Stdout, "[SLOW_QUERY_ALERT] ", log.LstdFlags|log.Lshortfile)
	} else {
		file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			log.Printf("Failed to open log file %s: %v", logFile, err)
			logger = log.New(os.Stdout, "[SLOW_QUERY_ALERT] ", log.LstdFlags|log.Lshortfile)
		} else {
			logger = log.New(file, "[SLOW_QUERY_ALERT] ", log.LstdFlags|log.Lshortfile)
		}
	}
	
	return &LogAlertChannel{
		logger:    logger,
		enabled:   true,
		logLevel:  strings.ToUpper(logLevel),
		formatter: &DefaultAlertFormatter{format: "text"},
	}
}

// SendAlert 发送日志告警
func (lac *LogAlertChannel) SendAlert(alert *AlertRecord) error {
	if !lac.enabled {
		return nil
	}
	
	// 根据严重程度过滤
	if !lac.shouldLog(alert.Severity) {
		return nil
	}
	
	message := lac.formatter.Format(alert)
	
	switch alert.Severity {
	case "CRITICAL":
		lac.logger.Printf("[CRITICAL] %s", message)
	case "WARNING", "VERY_SLOW":
		lac.logger.Printf("[WARNING] %s", message)
	default:
		lac.logger.Printf("[INFO] %s", message)
	}
	
	return nil
}

// GetChannelType 获取通道类型
func (lac *LogAlertChannel) GetChannelType() string {
	return "LOG"
}

// IsEnabled 是否启用
func (lac *LogAlertChannel) IsEnabled() bool {
	return lac.enabled
}

// SetEnabled 设置启用状态
func (lac *LogAlertChannel) SetEnabled(enabled bool) {
	lac.enabled = enabled
}

// shouldLog 判断是否应该记录日志
func (lac *LogAlertChannel) shouldLog(severity string) bool {
	severityLevels := map[string]int{
		"NORMAL":    0,
		"SLOW":      1,
		"WARNING":   2,
		"VERY_SLOW": 3,
		"CRITICAL":  4,
	}
	
	configLevel := severityLevels[lac.logLevel]
	alertLevel := severityLevels[severity]
	
	return alertLevel >= configLevel
}

// NewEmailAlertChannel 创建邮件告警通道
func NewEmailAlertChannel(config *EmailConfig) *EmailAlertChannel {
	channel := &EmailAlertChannel{
		config:  config,
		enabled: true,
		client:  &DefaultSMTPClient{},
	}
	
	// 加载邮件模板
	if config.TemplatePath != "" {
		tmpl, err := template.ParseFiles(config.TemplatePath)
		if err != nil {
			log.Printf("Failed to parse email template: %v", err)
			channel.template = channel.getDefaultTemplate()
		} else {
			channel.template = tmpl
		}
	} else {
		channel.template = channel.getDefaultTemplate()
	}
	
	return channel
}

// SendAlert 发送邮件告警
func (eac *EmailAlertChannel) SendAlert(alert *AlertRecord) error {
	if !eac.enabled {
		return nil
	}
	
	eac.mutex.Lock()
	defer eac.mutex.Unlock()
	
	// 构建邮件内容
	subject := eac.buildSubject(alert)
	body, err := eac.buildBody(alert)
	if err != nil {
		return fmt.Errorf("failed to build email body: %w", err)
	}
	
	// 构建邮件消息
	message := eac.buildMessage(subject, body)
	
	// 构建SMTP认证
	auth := smtp.PlainAuth("", eac.config.Username, eac.config.Password, eac.config.SMTPHost)
	
	// 发送邮件
	addr := fmt.Sprintf("%s:%d", eac.config.SMTPHost, eac.config.SMTPPort)
	err = eac.client.SendMail(addr, auth, eac.config.From, eac.config.To, []byte(message))
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}
	
	log.Printf("[EmailAlert] Sent alert email: %s", subject)
	return nil
}

// GetChannelType 获取通道类型
func (eac *EmailAlertChannel) GetChannelType() string {
	return "EMAIL"
}

// IsEnabled 是否启用
func (eac *EmailAlertChannel) IsEnabled() bool {
	return eac.enabled
}

// SetEnabled 设置启用状态
func (eac *EmailAlertChannel) SetEnabled(enabled bool) {
	eac.enabled = enabled
}

// buildSubject 构建邮件主题
func (eac *EmailAlertChannel) buildSubject(alert *AlertRecord) string {
	if eac.config.Subject != "" {
		return fmt.Sprintf("%s - %s Alert", eac.config.Subject, alert.Severity)
	}
	return fmt.Sprintf("Slow Query Alert - %s", alert.Severity)
}

// buildBody 构建邮件正文
func (eac *EmailAlertChannel) buildBody(alert *AlertRecord) (string, error) {
	var buf bytes.Buffer
	
	data := map[string]any{
		"Alert":     alert,
		"Timestamp": alert.Timestamp.Format("2006-01-02 15:04:05"),
		"Details":   alert.Details,
	}
	
	err := eac.template.Execute(&buf, data)
	if err != nil {
		return "", err
	}
	
	return buf.String(), nil
}

// buildMessage 构建邮件消息
func (eac *EmailAlertChannel) buildMessage(subject, body string) string {
	var buf bytes.Buffer
	
	buf.WriteString(fmt.Sprintf("From: %s\r\n", eac.config.From))
	buf.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(eac.config.To, ",")))
	buf.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	buf.WriteString("MIME-Version: 1.0\r\n")
	buf.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
	buf.WriteString("\r\n")
	buf.WriteString(body)
	
	return buf.String()
}

// getDefaultTemplate 获取默认邮件模板
func (eac *EmailAlertChannel) getDefaultTemplate() *template.Template {
	const defaultTemplate = `
<!DOCTYPE html>
<html>
<head>
    <title>Slow Query Alert</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        .alert { padding: 15px; border-radius: 5px; margin-bottom: 20px; }
        .critical { background-color: #ffebee; border-left: 5px solid #f44336; }
        .warning { background-color: #fff3e0; border-left: 5px solid #ff9800; }
        .info { background-color: #e3f2fd; border-left: 5px solid #2196f3; }
        .details { background-color: #f5f5f5; padding: 10px; border-radius: 3px; }
        pre { background-color: #f0f0f0; padding: 10px; overflow-x: auto; }
    </style>
</head>
<body>
    <div class="alert {{if eq .Alert.Severity "CRITICAL"}}critical{{else if eq .Alert.Severity "WARNING"}}warning{{else}}info{{end}}">
        <h2>{{.Alert.Severity}} Slow Query Alert</h2>
        <p><strong>Time:</strong> {{.Timestamp}}</p>
        <p><strong>Alert Type:</strong> {{.Alert.Type}}</p>
        <p><strong>Message:</strong> {{.Alert.Message}}</p>
    </div>
    
    <div class="details">
        <h3>Query Details</h3>
        {{if .Details.sql}}<p><strong>SQL:</strong></p>
        <pre>{{.Details.sql}}</pre>{{end}}
        
        {{if .Details.duration}}<p><strong>Duration:</strong> {{.Details.duration}}</p>{{end}}
        {{if .Details.memory_used}}<p><strong>Memory Used:</strong> {{.Details.memory_used}} bytes</p>{{end}}
        {{if .Details.table_names}}<p><strong>Tables:</strong> {{range .Details.table_names}}{{.}} {{end}}</p>{{end}}
    </div>
    
    <div class="details">
        <h3>Alert Information</h3>
        <p><strong>Alert ID:</strong> {{.Alert.ID}}</p>
        <p><strong>Slow Query ID:</strong> {{.Alert.SlowQueryID}}</p>
    </div>
</body>
</html>
`
	
	tmpl, err := template.New("default").Parse(defaultTemplate)
	if err != nil {
		log.Printf("Failed to parse default email template: %v", err)
		return template.Must(template.New("fallback").Parse("<p>{{.Alert.Message}}</p>"))
	}
	
	return tmpl
}

// SendMail 默认SMTP客户端发送邮件
func (dsc *DefaultSMTPClient) SendMail(addr string, a smtp.Auth, from string, to []string, msg []byte) error {
	return smtp.SendMail(addr, a, from, to, msg)
}

// NewWebhookAlertChannel 创建Webhook告警通道
func NewWebhookAlertChannel(config *WebhookConfig) *WebhookAlertChannel {
	if config.Method == "" {
		config.Method = "POST"
	}
	if config.Timeout == 0 {
		config.Timeout = 10 * time.Second
	}
	if config.RetryCount == 0 {
		config.RetryCount = 3
	}
	if config.RetryDelay == 0 {
		config.RetryDelay = 1 * time.Second
	}
	
	return &WebhookAlertChannel{
		config: config,
		enabled: true,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
		retryCount: config.RetryCount,
	}
}

// SendAlert 发送Webhook告警
func (wac *WebhookAlertChannel) SendAlert(alert *AlertRecord) error {
	if !wac.enabled {
		return nil
	}
	
	wac.mutex.Lock()
	defer wac.mutex.Unlock()
	
	// 构建payload
	payload := &WebhookPayload{
		Alert:     alert,
		Timestamp: alert.Timestamp.Format(time.RFC3339),
		Source:    "mybatis-slow-query-monitor",
	}
	
	// 添加签名
	if wac.config.SecretToken != "" {
		payload.Signature = wac.generateSignature(payload)
	}
	
	// 序列化payload
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal webhook payload: %w", err)
	}
	
	// 发送请求（带重试）
	return wac.sendWithRetry(jsonData)
}

// GetChannelType 获取通道类型
func (wac *WebhookAlertChannel) GetChannelType() string {
	return "WEBHOOK"
}

// IsEnabled 是否启用
func (wac *WebhookAlertChannel) IsEnabled() bool {
	return wac.enabled
}

// SetEnabled 设置启用状态
func (wac *WebhookAlertChannel) SetEnabled(enabled bool) {
	wac.enabled = enabled
}

// sendWithRetry 带重试发送
func (wac *WebhookAlertChannel) sendWithRetry(jsonData []byte) error {
	var lastErr error
	
	for i := 0; i <= wac.retryCount; i++ {
		if i > 0 {
			time.Sleep(wac.config.RetryDelay * time.Duration(i))
		}
		
		err := wac.sendRequest(jsonData)
		if err == nil {
			return nil
		}
		
		lastErr = err
		log.Printf("[WebhookAlert] Attempt %d failed: %v", i+1, err)
	}
	
	return fmt.Errorf("webhook alert failed after %d attempts: %w", wac.retryCount+1, lastErr)
}

// sendRequest 发送HTTP请求
func (wac *WebhookAlertChannel) sendRequest(jsonData []byte) error {
	req, err := http.NewRequest(wac.config.Method, wac.config.URL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	
	// 设置默认头部
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "mybatis-slow-query-monitor/1.0")
	
	// 设置自定义头部
	for key, value := range wac.config.Headers {
		req.Header.Set(key, value)
	}
	
	// 添加签名头部
	if wac.config.SecretToken != "" {
		signature := wac.generateSignatureForData(jsonData)
		req.Header.Set("X-Hub-Signature-256", signature)
	}
	
	// 发送请求
	resp, err := wac.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()
	
	// 检查响应状态码
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook returned status %d", resp.StatusCode)
	}
	
	log.Printf("[WebhookAlert] Successfully sent alert to %s", wac.config.URL)
	return nil
}

// generateSignature 生成签名
func (wac *WebhookAlertChannel) generateSignature(payload *WebhookPayload) string {
	// 简化的签名生成，实际应该使用HMAC
	return fmt.Sprintf("sha256=%x", payload.Alert.ID)
}

// generateSignatureForData 为数据生成签名
func (wac *WebhookAlertChannel) generateSignatureForData(data []byte) string {
	// 简化的签名生成，实际应该使用HMAC-SHA256
	return fmt.Sprintf("sha256=%x", len(data))
}

// NewConsoleAlertChannel 创建控制台告警通道
func NewConsoleAlertChannel(colorize bool) *ConsoleAlertChannel {
	return &ConsoleAlertChannel{
		enabled:   true,
		colorize:  colorize,
		formatter: &DefaultAlertFormatter{format: "console"},
	}
}

// SendAlert 发送控制台告警
func (cac *ConsoleAlertChannel) SendAlert(alert *AlertRecord) error {
	if !cac.enabled {
		return nil
	}
	
	message := cac.formatter.Format(alert)
	
	if cac.colorize {
		message = cac.colorizeMessage(alert.Severity, message)
	}
	
	fmt.Println(message)
	return nil
}

// GetChannelType 获取通道类型
func (cac *ConsoleAlertChannel) GetChannelType() string {
	return "CONSOLE"
}

// IsEnabled 是否启用
func (cac *ConsoleAlertChannel) IsEnabled() bool {
	return cac.enabled
}

// SetEnabled 设置启用状态
func (cac *ConsoleAlertChannel) SetEnabled(enabled bool) {
	cac.enabled = enabled
}

// colorizeMessage 给消息添加颜色
func (cac *ConsoleAlertChannel) colorizeMessage(severity, message string) string {
	var color string
	
	switch severity {
	case "CRITICAL":
		color = ColorRed
	case "WARNING", "VERY_SLOW":
		color = ColorYellow
	case "SLOW":
		color = ColorBlue
	default:
		color = ColorWhite
	}
	
	return fmt.Sprintf("%s%s%s", color, message, ColorReset)
}

// Format 默认格式化器
func (daf *DefaultAlertFormatter) Format(alert *AlertRecord) string {
	switch daf.format {
	case "json":
		data, _ := json.MarshalIndent(alert, "", "  ")
		return string(data)
	case "console":
		return fmt.Sprintf("[%s] %s Alert: %s (ID: %s, Time: %s)",
			alert.Severity,
			alert.Type,
			alert.Message,
			alert.ID,
			alert.Timestamp.Format("15:04:05"))
	default:
		return fmt.Sprintf("%s Alert: %s", alert.Severity, alert.Message)
	}
}

// GetFormat 获取格式
func (daf *DefaultAlertFormatter) GetFormat() string {
	return daf.format
}

// NewDefaultAlertFormatter 创建默认格式化器
func NewDefaultAlertFormatter(format string) *DefaultAlertFormatter {
	return &DefaultAlertFormatter{format: format}
}

// Format JSON格式化器
func (jaf *JSONAlertFormatter) Format(alert *AlertRecord) string {
	if jaf.indent {
		data, _ := json.MarshalIndent(alert, "", "  ")
		return string(data)
	}
	
	data, _ := json.Marshal(alert)
	return string(data)
}

// GetFormat 获取格式
func (jaf *JSONAlertFormatter) GetFormat() string {
	return "json"
}

// NewJSONAlertFormatter 创建JSON格式化器
func NewJSONAlertFormatter(indent bool) *JSONAlertFormatter {
	return &JSONAlertFormatter{indent: indent}
}

// Format 模板格式化器
func (taf *TemplateAlertFormatter) Format(alert *AlertRecord) string {
	if taf.template == nil {
		return alert.Message
	}
	
	var buf bytes.Buffer
	err := taf.template.Execute(&buf, map[string]any{
		"Alert": alert,
		"Now":   time.Now(),
	})
	
	if err != nil {
		log.Printf("Template format error: %v", err)
		return alert.Message
	}
	
	return buf.String()
}

// GetFormat 获取格式
func (taf *TemplateAlertFormatter) GetFormat() string {
	return "template"
}

// NewTemplateAlertFormatter 创建模板格式化器
func NewTemplateAlertFormatter(templateContent string) *TemplateAlertFormatter {
	tmpl, err := template.New("alert").Parse(templateContent)
	if err != nil {
		log.Printf("Failed to parse alert template: %v", err)
		return &TemplateAlertFormatter{template: nil}
	}
	
	return &TemplateAlertFormatter{template: tmpl}
}