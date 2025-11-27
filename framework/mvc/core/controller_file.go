package core

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/cloudwego/hertz/pkg/protocol/consts"

	"github.com/zsy619/yyhertz/framework/config"
)

// ============= 文件响应方法 =============

// File 发送文件响应
func (c *BaseController) File(filename string) {
	if c.Ctx == nil {
		config.Error("Context is nil when trying to send file")
		return
	}

	// 检查文件是否存在
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		config.Errorf("File not found: %s", filename)
		c.Status(consts.StatusNotFound)
		return
	}

	c.Ctx.FileAttachment(filename, "")
}

// FileAttachment 发送文件作为附件下载
func (c *BaseController) FileAttachment(filepath, filename string) {
	if c.Ctx == nil {
		config.Error("Context is nil when trying to send file attachment")
		return
	}

	// 检查文件是否存在
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		config.Errorf("File not found: %s", filepath)
		c.Status(consts.StatusNotFound)
		return
	}

	c.Ctx.FileAttachment(filepath, filename)
}

// Inline 内联发送文件
func (c *BaseController) Inline(file, name string) {
	if c.Ctx == nil {
		config.Error("Context is nil when trying to send inline file")
		return
	}

	// 检查文件是否存在
	if _, err := os.Stat(file); os.IsNotExist(err) {
		config.Errorf("File not found: %s", file)
		c.Status(consts.StatusNotFound)
		return
	}

	// 设置内联头
	if name != "" {
		c.SetHeader("Content-Disposition", fmt.Sprintf("inline; filename=\"%s\"", name))
	} else {
		c.SetHeader("Content-Disposition", "inline")
	}

	c.File(file)
}

// ServeFile 服务文件 (别名)
func (c *BaseController) ServeFile(filename string) {
	c.File(filename)
}

// Download 下载文件 (别名)
func (c *BaseController) Download(filepath, filename string) {
	c.FileAttachment(filepath, filename)
}

// ============= 文件上传处理方法 =============

// SaveUploadedFile 保存上传文件到指定路径
func (c *BaseController) SaveUploadedFile(fileKey, dst string) error {
	if c.Ctx == nil {
		return fmt.Errorf("context is nil")
	}

	_, file, err := c.Ctx.FormFile(fileKey)
	if err != nil {
		return fmt.Errorf("failed to get uploaded file: %v", err)
	}

	return c.Ctx.SaveUploadedFile(file, dst)
}

// GetFile 获取上传文件
func (c *BaseController) GetFile(key string) (multipart.File, *multipart.FileHeader, error) {
	if c.Ctx == nil {
		return nil, nil, fmt.Errorf("context is nil")
	}

	return c.Ctx.FormFile(key)
}

// HasFile 检查是否有上传文件
func (c *BaseController) HasFile(key string) bool {
	_, _, err := c.GetFile(key)
	return err == nil
}

// GetFileSize 获取上传文件大小
func (c *BaseController) GetFileSize(key string) int64 {
	_, file, err := c.GetFile(key)
	if err != nil {
		return 0
	}
	return file.Size
}

// GetFileName 获取上传文件名
func (c *BaseController) GetFileName(key string) string {
	_, file, err := c.GetFile(key)
	if err != nil {
		return ""
	}
	return file.Filename
}

// GetFileHeader 获取上传文件的MIME类型
func (c *BaseController) GetFileHeader(key string) string {
	_, file, err := c.GetFile(key)
	if err != nil {
		return ""
	}

	// 尝试从文件头获取Content-Type
	if file.Header != nil {
		return file.Header.Get("Content-Type")
	}
	return ""
}

// ============= 批量文件处理方法 =============

// GetFiles 获取所有上传文件
func (c *BaseController) GetFiles(key string) ([]*multipart.FileHeader, error) {
	if c.Ctx == nil {
		return nil, fmt.Errorf("context is nil")
	}

	// 解析多部分表单
	form, err := c.Ctx.Request().MultipartForm()
	if err != nil {
		return nil, fmt.Errorf("failed to parse multipart form: %v", err)
	}

	files := form.File[key]
	if len(files) == 0 {
		return nil, fmt.Errorf("no files found for key: %s", key)
	}

	return files, nil
}

// SaveMultipleFiles 保存多个上传文件
func (c *BaseController) SaveMultipleFiles(fileKey, dstDir string) ([]string, error) {
	files, err := c.GetFiles(fileKey)
	if err != nil {
		return nil, err
	}

	var savedFiles []string

	// 确保目标目录存在
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %v", err)
	}

	for i, file := range files {
		// 生成唯一文件名
		filename := file.Filename
		if filename == "" {
			filename = fmt.Sprintf("upload_%d", i)
		}

		dst := filepath.Join(dstDir, filename)

		// 如果文件已存在，添加序号
		if _, err := os.Stat(dst); err == nil {
			ext := filepath.Ext(filename)
			name := strings.TrimSuffix(filename, ext)
			for j := 1; ; j++ {
				newName := fmt.Sprintf("%s_%d%s", name, j, ext)
				dst = filepath.Join(dstDir, newName)
				if _, err := os.Stat(dst); os.IsNotExist(err) {
					break
				}
			}
		}

		if err := c.Ctx.SaveUploadedFile(file, dst); err != nil {
			return savedFiles, fmt.Errorf("failed to save file %s: %v", filename, err)
		}

		savedFiles = append(savedFiles, dst)
	}

	return savedFiles, nil
}

// ============= 文件信息和验证方法 =============

// ValidateFileSize 验证文件大小
func (c *BaseController) ValidateFileSize(key string, maxSize int64) error {
	size := c.GetFileSize(key)
	if size > maxSize {
		return fmt.Errorf("file size %d exceeds maximum allowed size %d", size, maxSize)
	}
	return nil
}

// ValidateFileExtension 验证文件扩展名
func (c *BaseController) ValidateFileExtension(key string, allowedExts []string) error {
	filename := c.GetFileName(key)
	if filename == "" {
		return fmt.Errorf("no file found for key: %s", key)
	}

	ext := strings.ToLower(filepath.Ext(filename))

	for _, allowedExt := range allowedExts {
		if ext == strings.ToLower(allowedExt) {
			return nil
		}
	}

	return fmt.Errorf("file extension %s is not allowed", ext)
}

// ValidateFileType 验证文件MIME类型
func (c *BaseController) ValidateFileType(key string, allowedTypes []string) error {
	fileType := c.GetFileHeader(key)
	if fileType == "" {
		return fmt.Errorf("unable to determine file type for key: %s", key)
	}

	for _, allowedType := range allowedTypes {
		if strings.Contains(fileType, allowedType) {
			return nil
		}
	}

	return fmt.Errorf("file type %s is not allowed", fileType)
}

// ============= 文件流处理方法 =============

// StreamFile 流式传输文件
func (c *BaseController) StreamFile(filename string) error {
	if c.Ctx == nil {
		return fmt.Errorf("context is nil")
	}

	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	// 获取文件信息
	info, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file info: %v", err)
	}

	// 设置响应头
	c.SetHeader("Content-Length", strconv.FormatInt(info.Size(), 10))
	c.SetHeader("Content-Type", "application/octet-stream")

	// 流式传输 - 使用Write方法替代Stream
	c.SetContentType("application/octet-stream")
	if _, err := c.Ctx.Copy(file); err != nil {
		return fmt.Errorf("failed to stream file: %v", err)
	}
	return nil
}

// StreamReader 流式传输Reader
func (c *BaseController) StreamReader(contentType string, reader io.Reader) {
	if c.Ctx == nil {
		config.Error("Context is nil when trying to stream reader")
		return
	}
	c.SetContentType(contentType)
	c.Ctx.Copy(reader)
}

// ============= 文件响应辅助方法 =============

// SetFileResponseHeaders 设置文件响应头
func (c *BaseController) SetFileResponseHeaders(filename string, inline bool) {
	if inline {
		c.SetHeader("Content-Disposition", fmt.Sprintf("inline; filename=\"%s\"", filename))
	} else {
		c.SetHeader("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	}

	// 根据文件扩展名设置Content-Type
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".pdf":
		c.SetContentType("application/pdf")
	case ".jpg", ".jpeg":
		c.SetContentType("image/jpeg")
	case ".png":
		c.SetContentType("image/png")
	case ".gif":
		c.SetContentType("image/gif")
	case ".txt":
		c.SetContentType("text/plain")
	case ".html", ".htm":
		c.SetContentType("text/html")
	case ".css":
		c.SetContentType("text/css")
	case ".js":
		c.SetContentType("application/javascript")
	case ".json":
		c.SetContentType("application/json")
	case ".xml":
		c.SetContentType("application/xml")
	default:
		c.SetContentType("application/octet-stream")
	}
}

// GetContentTypeByExtension 根据文件扩展名获取Content-Type
func (c *BaseController) GetContentTypeByExtension(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".pdf":
		return "application/pdf"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".txt":
		return "text/plain"
	case ".html", ".htm":
		return "text/html"
	case ".css":
		return "text/css"
	case ".js":
		return "application/javascript"
	case ".json":
		return "application/json"
	case ".xml":
		return "application/xml"
	default:
		return "application/octet-stream"
	}
}
