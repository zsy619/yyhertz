package context

import (
	"bytes"
	"strings"
	"testing"

	"github.com/cloudwego/hertz/pkg/app"
)

func TestCopy(t *testing.T) {
	// 创建测试Context
	c := &app.RequestContext{}
	ctx := NewContext(c)

	// 测试数据
	testData := "Hello, World! This is a test for the Copy functionality."
	reader := strings.NewReader(testData)

	// 执行Copy
	written, err := ctx.Copy(reader)
	if err != nil {
		t.Errorf("Copy failed: %v", err)
	}

	// 验证写入的字节数
	if written != int64(len(testData)) {
		t.Errorf("Expected %d bytes written, got %d", len(testData), written)
	}

	// 验证响应数据 (这里我们需要检查response body，但Hertz的测试方式可能不同)
	// 注意：实际的验证可能需要根据Hertz框架的测试方式进行调整
}

func TestCopyBuffer(t *testing.T) {
	// 创建测试Context
	c := &app.RequestContext{}
	ctx := NewContext(c)

	// 测试数据
	testData := "Test data for CopyBuffer method with custom buffer size."
	reader := strings.NewReader(testData)

	// 自定义缓冲区
	customBuffer := make([]byte, 16) // 16字节缓冲区

	// 执行CopyBuffer
	written, err := ctx.CopyBuffer(reader, customBuffer)
	if err != nil {
		t.Errorf("CopyBuffer failed: %v", err)
	}

	// 验证写入的字节数
	if written != int64(len(testData)) {
		t.Errorf("Expected %d bytes written, got %d", len(testData), written)
	}
}

func TestCopyWithContentType(t *testing.T) {
	// 创建测试Context
	c := &app.RequestContext{}
	ctx := NewContext(c)

	// 测试数据
	testData := `{"message": "Hello, World!"}`
	reader := strings.NewReader(testData)
	contentType := "application/json"

	// 执行CopyWithContentType
	written, err := ctx.CopyWithContentType(reader, contentType)
	if err != nil {
		t.Errorf("CopyWithContentType failed: %v", err)
	}

	// 验证写入的字节数
	if written != int64(len(testData)) {
		t.Errorf("Expected %d bytes written, got %d", len(testData), written)
	}

	// 验证Content-Type是否已设置
	responseContentType := ctx.GetResponseHeader("Content-Type")
	if responseContentType != contentType {
		t.Errorf("Expected Content-Type %s, got %s", contentType, responseContentType)
	}
}

func TestStreamCopy(t *testing.T) {
	// 创建测试Context
	c := &app.RequestContext{}
	ctx := NewContext(c)

	// 测试数据
	testData := "Stream data for testing StreamCopy method which flushes after each write."
	reader := strings.NewReader(testData)

	// 执行StreamCopy
	written, err := ctx.StreamCopy(reader)
	if err != nil {
		t.Errorf("StreamCopy failed: %v", err)
	}

	// 验证写入的字节数
	if written != int64(len(testData)) {
		t.Errorf("Expected %d bytes written, got %d", len(testData), written)
	}
}

func TestCopyNilReader(t *testing.T) {
	// 创建测试Context
	c := &app.RequestContext{}
	ctx := NewContext(c)

	// 测试nil reader
	_, err := ctx.Copy(nil)
	if err == nil {
		t.Error("Expected error for nil reader, got nil")
	}

	// 检查错误类型
	if contextErr, ok := err.(*ContextError); ok {
		if contextErr.Code != "NIL_READER" {
			t.Errorf("Expected error code 'NIL_READER', got '%s'", contextErr.Code)
		}
	} else {
		t.Error("Expected ContextError, got different error type")
	}
}

func TestCopyEmptyReader(t *testing.T) {
	// 创建测试Context
	c := &app.RequestContext{}
	ctx := NewContext(c)

	// 测试空reader
	reader := strings.NewReader("")

	// 执行Copy
	written, err := ctx.Copy(reader)
	if err != nil {
		t.Errorf("Copy with empty reader failed: %v", err)
	}

	// 验证写入的字节数应为0
	if written != 0 {
		t.Errorf("Expected 0 bytes written for empty reader, got %d", written)
	}
}

func TestCopyLargeData(t *testing.T) {
	// 创建测试Context
	c := &app.RequestContext{}
	ctx := NewContext(c)

	// 创建大数据 (64KB)
	largeData := bytes.Repeat([]byte("0123456789abcdef"), 4096) // 16 * 4096 = 64KB
	reader := bytes.NewReader(largeData)

	// 执行Copy
	written, err := ctx.Copy(reader)
	if err != nil {
		t.Errorf("Copy with large data failed: %v", err)
	}

	// 验证写入的字节数
	expectedSize := int64(len(largeData))
	if written != expectedSize {
		t.Errorf("Expected %d bytes written, got %d", expectedSize, written)
	}
}

func TestCopyBufferWithNilBuffer(t *testing.T) {
	// 创建测试Context
	c := &app.RequestContext{}
	ctx := NewContext(c)

	// 测试数据
	testData := "Test data for CopyBuffer with nil buffer."
	reader := strings.NewReader(testData)

	// 使用nil缓冲区，应该回退到默认Copy方法
	written, err := ctx.CopyBuffer(reader, nil)
	if err != nil {
		t.Errorf("CopyBuffer with nil buffer failed: %v", err)
	}

	// 验证写入的字节数
	if written != int64(len(testData)) {
		t.Errorf("Expected %d bytes written, got %d", len(testData), written)
	}
}

func BenchmarkCopy(b *testing.B) {
	// 创建测试Context
	c := &app.RequestContext{}
	ctx := NewContext(c)

	// 测试数据
	testData := strings.Repeat("benchmark test data ", 1000) // 约20KB数据

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reader := strings.NewReader(testData)
		_, err := ctx.Copy(reader)
		if err != nil {
			b.Errorf("Benchmark Copy failed: %v", err)
		}
	}
}

func BenchmarkCopyBuffer(b *testing.B) {
	// 创建测试Context
	c := &app.RequestContext{}
	ctx := NewContext(c)

	// 测试数据
	testData := strings.Repeat("benchmark test data ", 1000) // 约20KB数据
	customBuffer := make([]byte, 64*1024)                   // 64KB缓冲区

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reader := strings.NewReader(testData)
		_, err := ctx.CopyBuffer(reader, customBuffer)
		if err != nil {
			b.Errorf("Benchmark CopyBuffer failed: %v", err)
		}
	}
}