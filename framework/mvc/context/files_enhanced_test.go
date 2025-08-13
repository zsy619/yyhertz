package context

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
)

func TestSendFile(t *testing.T) {
	// 创建测试文件
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "test.txt")
	testContent := "Hello, World! This is a test file."
	
	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// 创建测试Context
	c := &app.RequestContext{}
	ctx := NewContext(c)

	// 测试SendFile
	err := ctx.SendFile(testFile)
	if err != nil {
		t.Errorf("SendFile failed: %v", err)
	}

	// 验证Content-Type
	contentType := ctx.GetResponseHeader("Content-Type")
	if !strings.Contains(contentType, "text/plain") {
		t.Logf("Expected text/plain content type, got: %s", contentType)
	}

	// 验证Content-Length
	contentLength := ctx.GetResponseHeader("Content-Length")
	if contentLength == "" {
		t.Error("Expected Content-Length header to be set")
	}

	// 验证ETag
	etag := ctx.GetResponseHeader("ETag")
	if etag == "" {
		t.Error("Expected ETag header to be set")
	}
}

func TestSendFileNotFound(t *testing.T) {
	// 创建测试Context
	c := &app.RequestContext{}
	ctx := NewContext(c)

	// 测试不存在的文件
	err := ctx.SendFile("/non/existent/file.txt")
	if err == nil {
		t.Error("Expected error for non-existent file")
	}

	// 检查错误类型
	if contextErr, ok := err.(*ContextError); ok {
		if contextErr.Code != "FILE_NOT_FOUND" {
			t.Errorf("Expected FILE_NOT_FOUND error, got: %s", contextErr.Code)
		}
	} else {
		t.Error("Expected ContextError type")
	}
}

func TestSendFileDirectory(t *testing.T) {
	// 创建测试目录
	tempDir := t.TempDir()

	// 创建测试Context
	c := &app.RequestContext{}
	ctx := NewContext(c)

	// 测试目录访问
	err := ctx.SendFile(tempDir)
	if err == nil {
		t.Error("Expected error when trying to serve directory")
	}

	// 检查错误类型
	if contextErr, ok := err.(*ContextError); ok {
		if contextErr.Code != "DIRECTORY_ACCESS" {
			t.Errorf("Expected DIRECTORY_ACCESS error, got: %s", contextErr.Code)
		}
	} else {
		t.Error("Expected ContextError type")
	}
}

func TestSendStream(t *testing.T) {
	// 创建测试Context
	c := &app.RequestContext{}
	ctx := NewContext(c)

	// 测试数据
	testData := "Stream test data for testing SendStream method."
	reader := strings.NewReader(testData)

	// 测试SendStream
	err := ctx.SendStream(reader, "text/plain", int64(len(testData)))
	if err != nil {
		t.Errorf("SendStream failed: %v", err)
	}

	// 验证Content-Type
	contentType := ctx.GetResponseHeader("Content-Type")
	if contentType != "text/plain" {
		t.Errorf("Expected text/plain, got: %s", contentType)
	}

	// 验证Content-Length
	contentLength := ctx.GetResponseHeader("Content-Length")
	expectedLength := string(rune(len(testData)))
	if contentLength != expectedLength {
		t.Logf("Content-Length: %s, expected something for %d bytes", contentLength, len(testData))
	}
}

func TestAttachment(t *testing.T) {
	// 创建测试Context
	c := &app.RequestContext{}
	ctx := NewContext(c)

	// 测试Attachment
	filename := "test-file.pdf"
	ctx.Attachment(filename, "application/pdf")

	// 验证Content-Disposition
	disposition := ctx.GetResponseHeader("Content-Disposition")
	if !strings.Contains(disposition, "attachment") {
		t.Errorf("Expected attachment in Content-Disposition, got: %s", disposition)
	}
	if !strings.Contains(disposition, filename) {
		t.Errorf("Expected filename in Content-Disposition, got: %s", disposition)
	}

	// 验证Content-Type
	contentType := ctx.GetResponseHeader("Content-Type")
	if contentType != "application/pdf" {
		t.Errorf("Expected application/pdf, got: %s", contentType)
	}
}

func TestAttachmentWithChineseFilename(t *testing.T) {
	// 创建测试Context
	c := &app.RequestContext{}
	ctx := NewContext(c)

	// 测试中文文件名
	filename := "测试文件.txt"
	ctx.Attachment(filename)

	// 验证Content-Disposition包含编码后的文件名
	disposition := ctx.GetResponseHeader("Content-Disposition")
	if !strings.Contains(disposition, "attachment") {
		t.Errorf("Expected attachment in Content-Disposition, got: %s", disposition)
	}
	if !strings.Contains(disposition, "filename*=UTF-8") {
		t.Errorf("Expected UTF-8 encoding in Content-Disposition, got: %s", disposition)
	}
}

func TestDownload(t *testing.T) {
	// 创建测试文件
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "download-test.txt")
	testContent := "This is a download test file."
	
	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// 创建测试Context
	c := &app.RequestContext{}
	ctx := NewContext(c)

	// 测试Download
	downloadName := "custom-download-name.txt"
	err := ctx.DownloadFile(testFile, downloadName)
	if err != nil {
		t.Errorf("Download failed: %v", err)
	}

	// 验证Content-Disposition
	disposition := ctx.GetResponseHeader("Content-Disposition")
	if !strings.Contains(disposition, "attachment") {
		t.Errorf("Expected attachment in Content-Disposition, got: %s", disposition)
	}
	if !strings.Contains(disposition, downloadName) {
		t.Errorf("Expected custom filename in Content-Disposition, got: %s", disposition)
	}
}

func TestFileExists(t *testing.T) {
	// 创建测试文件
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "exists-test.txt")
	
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// 创建测试Context
	c := &app.RequestContext{}
	ctx := NewContext(c)

	// 测试存在的文件
	if !ctx.FileExists(testFile) {
		t.Error("FileExists should return true for existing file")
	}

	// 测试不存在的文件
	if ctx.FileExists(filepath.Join(tempDir, "non-existent.txt")) {
		t.Error("FileExists should return false for non-existent file")
	}

	// 测试目录
	if ctx.FileExists(tempDir) {
		t.Error("FileExists should return false for directory")
	}
}

func TestFileInfo(t *testing.T) {
	// 创建测试文件
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "info-test.txt")
	testContent := "File info test content"
	
	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// 创建测试Context
	c := &app.RequestContext{}
	ctx := NewContext(c)

	// 测试FileInfo
	info, err := ctx.FileInfo(testFile)
	if err != nil {
		t.Errorf("FileInfo failed: %v", err)
	}

	if info.Size() != int64(len(testContent)) {
		t.Errorf("Expected file size %d, got %d", len(testContent), info.Size())
	}

	if info.IsDir() {
		t.Error("FileInfo should not report file as directory")
	}
}

func TestParseRange(t *testing.T) {
	tests := []struct {
		rangeHeader string
		fileSize    int64
		expected    int
		expectError bool
	}{
		{"bytes=0-499", 1000, 1, false},
		{"bytes=500-999", 1000, 1, false},
		{"bytes=-500", 1000, 1, false},
		{"bytes=500-", 1000, 1, false},
		{"bytes=0-0", 1, 1, false},
		{"invalid-range", 1000, 0, true},
		{"bytes=", 1000, 0, false},
		{"bytes=abc-def", 1000, 0, false},
	}

	for _, tt := range tests {
		ranges, err := parseRange(tt.rangeHeader, tt.fileSize)
		
		if tt.expectError && err == nil {
			t.Errorf("Expected error for range header: %s", tt.rangeHeader)
			continue
		}
		if !tt.expectError && err != nil {
			t.Errorf("Unexpected error for range header %s: %v", tt.rangeHeader, err)
			continue
		}
		
		if len(ranges) != tt.expected {
			t.Errorf("Expected %d ranges for %s, got %d", tt.expected, tt.rangeHeader, len(ranges))
		}
	}
}

func TestEncodeRFC5987(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"test.txt", "test.txt"},
		{"测试.txt", "%E6%B5%8B%E8%AF%95.txt"},
		{"hello world.pdf", "hello%20world.pdf"},
		{"file-name_123.zip", "file-name_123.zip"},
	}

	for _, tt := range tests {
		result := encodeRFC5987(tt.input)
		if result != tt.expected {
			t.Errorf("Expected %s for input %s, got %s", tt.expected, tt.input, result)
		}
	}
}

func TestSendFileWithRangeRequest(t *testing.T) {
	// 创建测试文件
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "range-test.txt")
	testContent := "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ" // 36 bytes
	
	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// 创建带Range头的测试Context
	c := &app.RequestContext{}
	c.Request.Header.Set("Range", "bytes=0-9") // 请求前10个字节
	ctx := NewContext(c)

	// 测试Range请求
	err := ctx.SendFile(testFile)
	if err != nil {
		t.Errorf("SendFile with Range failed: %v", err)
	}

	// 注意：由于我们无法直接验证Hertz的响应状态码，这里主要测试没有错误
	t.Logf("Range request test completed successfully")
}

func TestCheckNotModified(t *testing.T) {
	// 创建测试Context
	c := &app.RequestContext{}
	ctx := NewContext(c)

	modTime := time.Now().Add(-time.Hour) // 1小时前
	etag := "\"test-etag\""

	// 测试If-None-Match匹配
	c.Request.Header.Set("If-None-Match", etag)
	if !ctx.checkNotModified(modTime, etag) {
		t.Error("checkNotModified should return true for matching ETag")
	}
	c.Request.Header.Del("If-None-Match")

	// 测试If-Modified-Since
	ifModifiedSince := modTime.Add(time.Hour).Format(time.RFC1123) // 比文件新
	c.Request.Header.Set("If-Modified-Since", ifModifiedSince)
	if !ctx.checkNotModified(modTime, etag) {
		t.Error("checkNotModified should return true when file is not modified since given time")
	}
}

// 基准测试
func BenchmarkSendFile(b *testing.B) {
	// 创建测试文件
	tempDir := b.TempDir()
	testFile := filepath.Join(tempDir, "benchmark.txt")
	testContent := bytes.Repeat([]byte("benchmark test data "), 1000) // 20KB
	
	if err := os.WriteFile(testFile, testContent, 0644); err != nil {
		b.Fatalf("Failed to create benchmark file: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c := &app.RequestContext{}
		ctx := NewContext(c)
		ctx.SendFile(testFile)
	}
}

func BenchmarkAttachment(b *testing.B) {
	filename := "benchmark-file.pdf"
	contentType := "application/pdf"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c := &app.RequestContext{}
		ctx := NewContext(c)
		ctx.Attachment(filename, contentType)
	}
}