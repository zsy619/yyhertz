package context

import (
	"fmt"
	"io"
	"mime"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// ============= 文件服务增强功能 =============

// SendFile 高效地发送文件到客户端，支持Range请求和缓存控制
// 这个方法提供了完整的静态文件服务功能，包括：
// - Range请求支持（断点续传）
// - 自动MIME类型检测
// - 缓存头设置
// - 文件不存在处理
func (ctx *Context) SendFile(filePath string) error {
	if !ctx.ensureRequest() {
		return ErrRequestNotFound
	}

	// 检查文件是否存在
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			ctx.Status(StatusNotFound)
			return &ContextError{
				Code:    "FILE_NOT_FOUND",
				Message: "File not found: " + filePath,
				Cause:   err,
			}
		}
		return &ContextError{
			Code:    "FILE_ACCESS_ERROR",
			Message: "Cannot access file: " + filePath,
			Cause:   err,
		}
	}

	// 检查是否为目录
	if fileInfo.IsDir() {
		ctx.Status(StatusForbidden)
		return &ContextError{
			Code:    "DIRECTORY_ACCESS",
			Message: "Cannot serve directory as file: " + filePath,
		}
	}

	// 打开文件
	file, err := os.Open(filePath)
	if err != nil {
		return &ContextError{
			Code:    "FILE_OPEN_ERROR",
			Message: "Cannot open file: " + filePath,
			Cause:   err,
		}
	}
	defer file.Close()

	// 设置内容类型
	contentType := mime.TypeByExtension(filepath.Ext(filePath))
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	ctx.SetContentType(contentType)

	// 设置缓存相关头部
	ctx.SetHeader("Last-Modified", fileInfo.ModTime().UTC().Format(time.RFC1123))
	ctx.SetHeader("Content-Length", strconv.FormatInt(fileInfo.Size(), 10))

	// 生成ETag
	etag := fmt.Sprintf("\"%d-%d\"", fileInfo.Size(), fileInfo.ModTime().Unix())
	ctx.SetHeader("ETag", etag)

	// 检查条件请求
	if ctx.checkNotModified(fileInfo.ModTime(), etag) {
		ctx.Status(StatusNotModified)
		return nil
	}

	// 处理Range请求
	rangeHeader := ctx.Header("Range")
	if rangeHeader != "" {
		return ctx.sendFileWithRange(file, fileInfo.Size(), rangeHeader)
	}

	// 发送完整文件
	ctx.Status(StatusOK)
	_, err = ctx.Copy(file)
	return err
}

// SendStream 发送流数据到客户端，适合大文件或实时生成的内容
func (ctx *Context) SendStream(reader io.Reader, contentType string, size int64) error {
	if !ctx.ensureRequest() {
		return ErrRequestNotFound
	}

	// 设置响应头
	ctx.SetContentType(contentType)
	if size > 0 {
		ctx.SetHeader("Content-Length", strconv.FormatInt(size, 10))
	}

	// 使用流式复制
	_, err := ctx.StreamCopy(reader)
	return err
}

// Attachment 设置文件下载响应头，使浏览器提示下载而不是预览
func (ctx *Context) Attachment(filename string, contentType ...string) {
	if !ctx.ensureRequest() {
		return
	}

	// 设置Content-Disposition头
	disposition := "attachment"
	if filename != "" {
		// 处理文件名编码，支持中文文件名
		encodedFilename := encodeRFC5987(filename)
		disposition += fmt.Sprintf("; filename=\"%s\"; filename*=UTF-8''%s", filename, encodedFilename)
	}
	ctx.SetHeader("Content-Disposition", disposition)

	// 设置内容类型
	if len(contentType) > 0 && contentType[0] != "" {
		ctx.SetContentType(contentType[0])
	} else if filename != "" {
		// 从文件扩展名推断MIME类型
		mimeType := mime.TypeByExtension(filepath.Ext(filename))
		if mimeType != "" {
			ctx.SetContentType(mimeType)
		}
	}
}

// DownloadFile 便捷的文件下载方法，结合SendFile和Attachment
func (ctx *Context) DownloadFile(filePath string, filename ...string) error {
	if !ctx.ensureRequest() {
		return ErrRequestNotFound
	}

	// 确定下载文件名
	downloadName := ""
	if len(filename) > 0 && filename[0] != "" {
		downloadName = filename[0]
	} else {
		downloadName = filepath.Base(filePath)
	}

	// 设置下载响应头
	ctx.Attachment(downloadName)

	// 发送文件
	return ctx.SendFile(filePath)
}

// SaveFormFile 便捷的上传文件保存方法
func (ctx *Context) SaveFormFile(formName, destPath string) error {
	if !ctx.ensureRequest() {
		return ErrRequestNotFound
	}

	// 获取上传的文件
	fileHeader, err := ctx.request.FormFile(formName)
	if err != nil {
		return &ContextError{
			Code:    "FILE_UPLOAD_ERROR",
			Message: "Cannot get uploaded file: " + formName,
			Cause:   err,
		}
	}

	// 打开上传的文件
	src, err := fileHeader.Open()
	if err != nil {
		return &ContextError{
			Code:    "UPLOADED_FILE_OPEN_ERROR",
			Message: "Cannot open uploaded file",
			Cause:   err,
		}
	}
	defer src.Close()

	// 创建目标目录（如果不存在）
	destDir := filepath.Dir(destPath)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return &ContextError{
			Code:    "DEST_DIR_CREATE_ERROR",
			Message: "Cannot create destination directory: " + destDir,
			Cause:   err,
		}
	}

	// 创建目标文件
	dst, err := os.Create(destPath)
	if err != nil {
		return &ContextError{
			Code:    "DEST_FILE_CREATE_ERROR",
			Message: "Cannot create destination file: " + destPath,
			Cause:   err,
		}
	}
	defer dst.Close()

	// 复制文件内容
	_, err = io.Copy(dst, src)
	if err != nil {
		// 如果复制失败，删除不完整的文件
		os.Remove(destPath)
		return &ContextError{
			Code:    "FILE_COPY_ERROR",
			Message: "Failed to copy uploaded file to destination",
			Cause:   err,
		}
	}

	return nil
}

// ============= 静态文件服务相关方法 =============

// ServeStatic 服务单个静态文件（SendFile的别名，为了兼容性）
func (ctx *Context) ServeStatic(filePath string) error {
	return ctx.SendFile(filePath)
}

// FileExists 检查文件是否存在且可读
func (ctx *Context) FileExists(filePath string) bool {
	info, err := os.Stat(filePath)
	return err == nil && !info.IsDir()
}

// FileInfo 获取文件信息
func (ctx *Context) FileInfo(filePath string) (os.FileInfo, error) {
	return os.Stat(filePath)
}

// ============= Range请求支持 =============

// sendFileWithRange 处理Range请求（部分内容请求）
func (ctx *Context) sendFileWithRange(file *os.File, fileSize int64, rangeHeader string) error {
	// 解析Range头
	ranges, err := parseRange(rangeHeader, fileSize)
	if err != nil {
		ctx.Status(StatusRequestedRangeNotSatisfiable)
		ctx.SetHeader("Content-Range", fmt.Sprintf("bytes */%d", fileSize))
		return err
	}

	if len(ranges) == 0 {
		// 无效的Range请求，返回完整文件
		ctx.Status(StatusOK)
		_, err = ctx.Copy(file)
		return err
	}

	// 目前只支持单个range，多range支持需要multipart response
	if len(ranges) == 1 {
		r := ranges[0]
		
		// 设置响应头
		ctx.Status(StatusPartialContent)
		ctx.SetHeader("Content-Range", fmt.Sprintf("bytes %d-%d/%d", r.start, r.end, fileSize))
		ctx.SetHeader("Content-Length", strconv.FormatInt(r.end-r.start+1, 10))

		// 移动到起始位置
		_, err = file.Seek(r.start, io.SeekStart)
		if err != nil {
			return err
		}

		// 创建限制读取器，只读取指定范围
		limitReader := io.LimitReader(file, r.end-r.start+1)
		_, err = ctx.Copy(limitReader)
		return err
	}

	// 多range请求（暂不支持，返回完整文件）
	ctx.Status(StatusOK)
	_, err = ctx.Copy(file)
	return err
}

// ============= 缓存相关辅助方法 =============

// checkNotModified 检查条件请求（If-Modified-Since, If-None-Match）
func (ctx *Context) checkNotModified(modTime time.Time, etag string) bool {
	// 检查If-None-Match (ETag)
	ifNoneMatch := ctx.Header("If-None-Match")
	if ifNoneMatch != "" {
		if ifNoneMatch == "*" || ifNoneMatch == etag {
			return true
		}
	}

	// 检查If-Modified-Since
	ifModifiedSince := ctx.Header("If-Modified-Since")
	if ifModifiedSince != "" {
		if t, err := time.Parse(time.RFC1123, ifModifiedSince); err == nil {
			// 只精确到秒，忽略纳秒差异
			if !modTime.Truncate(time.Second).After(t) {
				return true
			}
		}
	}

	return false
}

// ============= 辅助类型和函数 =============

// httpRange 表示HTTP Range请求的范围
type httpRange struct {
	start, end int64
}

// parseRange 解析HTTP Range头
func parseRange(rangeHeader string, size int64) ([]httpRange, error) {
	if !strings.HasPrefix(rangeHeader, "bytes=") {
		return nil, fmt.Errorf("invalid range header")
	}

	rangeSpec := strings.TrimPrefix(rangeHeader, "bytes=")
	ranges := strings.Split(rangeSpec, ",")
	
	var result []httpRange
	for _, r := range ranges {
		r = strings.TrimSpace(r)
		if r == "" {
			continue
		}

		parts := strings.Split(r, "-")
		if len(parts) != 2 {
			continue
		}

		var start, end int64
		var err error

		if parts[0] == "" {
			// 后缀范围 "-500"
			if parts[1] == "" {
				continue
			}
			suffix, err := strconv.ParseInt(parts[1], 10, 64)
			if err != nil || suffix <= 0 {
				continue
			}
			start = size - suffix
			if start < 0 {
				start = 0
			}
			end = size - 1
		} else if parts[1] == "" {
			// 前缀范围 "500-"
			start, err = strconv.ParseInt(parts[0], 10, 64)
			if err != nil || start < 0 || start >= size {
				continue
			}
			end = size - 1
		} else {
			// 完整范围 "0-499"
			start, err = strconv.ParseInt(parts[0], 10, 64)
			if err != nil || start < 0 {
				continue
			}
			end, err = strconv.ParseInt(parts[1], 10, 64)
			if err != nil || end < start || end >= size {
				continue
			}
		}

		if start <= end && start < size {
			result = append(result, httpRange{start: start, end: end})
		}
	}

	return result, nil
}

// encodeRFC5987 按照RFC 5987编码文件名，支持Unicode字符
func encodeRFC5987(s string) string {
	var buf strings.Builder
	for _, r := range s {
		switch {
		case r >= 'A' && r <= 'Z' || r >= 'a' && r <= 'z' || r >= '0' && r <= '9':
			buf.WriteRune(r)
		case r == '-' || r == '.' || r == '_' || r == '~':
			buf.WriteRune(r)
		default:
			for _, b := range []byte(string(r)) {
				fmt.Fprintf(&buf, "%%%02X", b)
			}
		}
	}
	return buf.String()
}

// ============= 文件上传相关错误 =============

var (
	ErrFileNotFound     = &ContextError{Code: "FILE_NOT_FOUND", Message: "File not found"}
	ErrDirectoryAccess  = &ContextError{Code: "DIRECTORY_ACCESS", Message: "Directory access not allowed"}
	ErrFileUploadFailed = &ContextError{Code: "FILE_UPLOAD_FAILED", Message: "File upload failed"}
)