package context

import (
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
)

// ============= 文件上传处理方法 =============

// FormFile 获取上传的文件
func (ctx *Context) FormFile(name string) (multipart.File, *multipart.FileHeader, error) {
	if !ctx.ensureRequest() {
		return nil, nil, ErrRequestNotFound
	}

	hd, err := ctx.request.FormFile(name)
	if err != nil {
		return nil, nil, err
	}
	fl, err := hd.Open()
	return fl, hd, err
}

// SaveUploadedFile 保存上传文件到指定路径
func (ctx *Context) SaveUploadedFile(file *multipart.FileHeader, dst string) error {
	if file == nil {
		return ErrEmptyBody
	}

	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	// 创建目标目录
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	// 创建目标文件
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	// 复制文件内容
	_, err = io.Copy(out, src)
	return err
}

// ============= 文件下载响应方法 =============

// File 发送文件
func (ctx *Context) File(filepath string) {
	if ctx.ensureRequest() {
		ctx.request.File(filepath)
	}
}

// FileAttachment 发送文件作为附件下载
func (ctx *Context) FileAttachment(filepath, filename string) {
	if !ctx.ensureRequest() {
		return
	}

	if filename != "" {
		ctx.SetHeader(HeaderContentDisposition, "attachment; filename=\""+filename+"\"")
	}

	ctx.request.File(filepath)
}

// Inline 发送文件作为内联显示
func (ctx *Context) Inline(filepath, filename string) {
	if !ctx.ensureRequest() {
		return
	}

	if filename != "" {
		ctx.SetHeader(HeaderContentDisposition, "inline; filename=\""+filename+"\"")
	}

	ctx.request.File(filepath)
}

// Download 下载文件（FileAttachment的别名）
func (ctx *Context) Download(filepath, filename string) {
	ctx.FileAttachment(filepath, filename)
}

// ServeFile 服务文件（File的别名）
func (ctx *Context) ServeFile(filepath string) {
	ctx.File(filepath)
}

// ============= 文件信息获取方法 =============

// HasFile 检查是否有文件上传
func (ctx *Context) HasFile(name string) bool {
	_, file, err := ctx.FormFile(name)
	return err == nil && file != nil
}

// GetFileSize 获取上传文件大小
func (ctx *Context) GetFileSize(name string) int64 {
	_, file, err := ctx.FormFile(name)
	if err != nil || file == nil {
		return 0
	}
	return file.Size
}

// GetFileName 获取上传文件名
func (ctx *Context) GetFileName(name string) string {
	_, file, err := ctx.FormFile(name)
	if err != nil || file == nil {
		return ""
	}
	return file.Filename
}

// GetFileHeader 获取文件头信息
func (ctx *Context) GetFileHeader(name string) *multipart.FileHeader {
	_, file, _ := ctx.FormFile(name)
	return file
}

// ============= 批量文件处理方法 =============

// GetFiles 获取所有上传的文件
func (ctx *Context) GetFiles(name string) []*multipart.FileHeader {
	if !ctx.ensureRequest() {
		return nil
	}

	form, err := ctx.request.MultipartForm()
	if err != nil {
		return nil
	}

	if form.File == nil {
		return nil
	}

	return form.File[name]
}

// SaveMultipleFiles 保存多个文件
func (ctx *Context) SaveMultipleFiles(name, destDir string) error {
	files := ctx.GetFiles(name)
	if len(files) == 0 {
		return ErrNoFilesUploaded
	}

	for i, file := range files {
		filename := file.Filename
		if filename == "" {
			filename = "upload_" + string(rune('0'+i))
		}

		destPath := filepath.Join(destDir, filename)
		if err := ctx.SaveUploadedFile(file, destPath); err != nil {
			return err
		}
	}

	return nil
}

// ============= 文件验证方法 =============

// ValidateFileSize 验证文件大小
func (ctx *Context) ValidateFileSize(name string, maxSize int64) error {
	fileSize := ctx.GetFileSize(name)
	if fileSize == 0 {
		return ErrNoFileUploaded
	}
	if fileSize > maxSize {
		return ErrFileTooLarge
	}
	return nil
}

// ValidateFileExtension 验证文件扩展名
func (ctx *Context) ValidateFileExtension(name string, allowedExt []string) error {
	filename := ctx.GetFileName(name)
	if filename == "" {
		return ErrNoFileUploaded
	}

	ext := filepath.Ext(filename)
	for _, allowed := range allowedExt {
		if ext == allowed {
			return nil
		}
	}

	return ErrInvalidFileExtension
}

// ValidateFileType 验证文件MIME类型
func (ctx *Context) ValidateFileType(name string, allowedTypes []string) error {
	_, file, err := ctx.FormFile(name)
	if err != nil {
		return ErrNoFileUploaded
	}

	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	// 读取文件头来检测MIME类型
	buffer := make([]byte, 512)
	_, err = src.Read(buffer)
	if err != nil {
		return err
	}

	// 这里需要实现MIME类型检测逻辑
	// 为简化，暂时返回nil
	return nil
}

// ============= 文件流处理方法 =============

// StreamFile 流式发送文件
func (ctx *Context) StreamFile(filepath string, contentType string) error {
	if !ctx.ensureRequest() {
		return ErrRequestNotFound
	}

	file, err := os.Open(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	if contentType != "" {
		ctx.SetContentType(contentType)
	}

	_, err = io.Copy(ctx.request.Response.BodyWriter(), file)
	return err
}

// StreamReader 流式发送数据
func (ctx *Context) StreamReader(reader io.Reader, contentType string) error {
	if !ctx.ensureRequest() {
		return ErrRequestNotFound
	}

	if contentType != "" {
		ctx.SetContentType(contentType)
	}

	_, err := io.Copy(ctx.request.Response.BodyWriter(), reader)
	return err
}

// ============= 错误定义 =============

var (
	ErrNoFileUploaded       = &ContextError{Code: "NO_FILE_UPLOADED", Message: "No file uploaded"}
	ErrNoFilesUploaded      = &ContextError{Code: "NO_FILES_UPLOADED", Message: "No files uploaded"}
	ErrFileTooLarge         = &ContextError{Code: "FILE_TOO_LARGE", Message: "File size exceeds limit"}
	ErrInvalidFileExtension = &ContextError{Code: "INVALID_FILE_EXTENSION", Message: "Invalid file extension"}
	ErrInvalidFileType      = &ContextError{Code: "INVALID_FILE_TYPE", Message: "Invalid file type"}
)
