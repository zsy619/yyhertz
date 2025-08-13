package context

// 这个文件提供了使用Copy功能的示例

import (
	"bytes"
	"io"
	"os"
	"strings"
)

// ExampleCopyUsage 演示Copy方法的基本用法
func ExampleCopyUsage(ctx *Context) {
	// 示例1: 从字符串复制数据到响应
	data := "Hello, World! 这是一个Copy方法的示例。"
	reader := strings.NewReader(data)

	// 使用Copy方法将数据写入响应
	bytesWritten, err := ctx.Copy(reader)
	if err != nil {
		// 处理错误
		ctx.JSON(500, map[string]any{
			"error":   "复制数据失败",
			"message": err.Error(),
		})
		return
	}

	// 可以记录写入的字节数用于监控
	ctx.Set("bytes_written", bytesWritten)
}

// ExampleCopyWithContentType 演示带Content-Type的复制
func ExampleCopyWithContentType(ctx *Context) {
	// JSON数据示例
	jsonData := `{
		"status": "success",
		"message": "这是通过Copy方法传输的JSON数据",
		"data": {
			"timestamp": "2023-01-01T00:00:00Z",
			"version": "1.0"
		}
	}`

	reader := strings.NewReader(jsonData)

	// 使用CopyWithContentType一次性设置Content-Type并复制数据
	bytesWritten, err := ctx.CopyWithContentType(reader, "application/json; charset=utf-8")
	if err != nil {
		ctx.JSON(500, map[string]any{
			"error":   "复制JSON数据失败",
			"details": err.Error(),
		})
		return
	}

	// 记录传输统计
	ctx.Set("json_bytes_transmitted", bytesWritten)
}

// ExampleCopyBuffer 演示自定义缓冲区的Copy
func ExampleCopyBuffer(ctx *Context) {
	// 假设我们有大量数据需要传输
	largeData := bytes.Repeat([]byte("ABCDEFGHIJK"), 10000) // 约110KB数据
	reader := bytes.NewReader(largeData)

	// 使用64KB的自定义缓冲区进行高效传输
	customBuffer := make([]byte, 64*1024)

	bytesWritten, err := ctx.CopyBuffer(reader, customBuffer)
	if err != nil {
		ctx.JSON(500, map[string]any{
			"error": "大数据传输失败",
			"size":  len(largeData),
		})
		return
	}

	// 设置适当的Content-Type
	ctx.SetContentType("application/octet-stream")

	// 记录传输信息
	ctx.Set("large_data_transmitted", bytesWritten)
}

// ExampleStreamCopy 演示流式传输
func ExampleStreamCopy(ctx *Context) {
	// 模拟流数据源（比如从数据库或API获取数据）
	streamData := "实时数据流: "
	for i := 0; i < 100; i++ {
		streamData += "数据块" + string(rune('A'+i%26)) + " "
	}

	reader := strings.NewReader(streamData)

	// 设置流式传输的头部
	ctx.SetHeader("Transfer-Encoding", "chunked")
	ctx.SetContentType("text/plain; charset=utf-8")

	// 使用StreamCopy进行流式传输，每次写入后自动刷新
	bytesWritten, err := ctx.StreamCopy(reader)
	if err != nil {
		// 注意：在流式传输中，错误处理需要特别小心
		// 因为响应可能已经开始发送
		return
	}

	ctx.Set("stream_bytes_sent", bytesWritten)
}

// ExampleFileTransfer 演示文件传输场景
func ExampleFileTransfer(ctx *Context) {
	// 假设我们要传输一个文件
	filePath := "/path/to/your/file.txt"

	// 打开文件
	file, err := os.Open(filePath)
	if err != nil {
		ctx.JSON(404, map[string]any{
			"error": "文件未找到",
			"path":  filePath,
		})
		return
	}
	defer file.Close()

	// 设置文件下载相关的头部
	ctx.SetHeader("Content-Disposition", "attachment; filename=\"file.txt\"")
	ctx.SetHeader("Content-Type", "application/octet-stream")

	// 使用Copy传输文件内容
	bytesWritten, err := ctx.Copy(file)
	if err != nil {
		// 文件传输错误
		// 注意：如果已经开始传输，就不能再发送JSON错误响应了
		return
	}

	// 记录文件传输统计
	ctx.Set("file_bytes_transmitted", bytesWritten)
	ctx.Set("file_path", filePath)
}

// ExampleErrorHandling 演示错误处理
func ExampleErrorHandling(ctx *Context) {
	// 模拟一个可能出错的reader
	errorReader := &ErrorReader{shouldFail: true}

	bytesWritten, err := ctx.Copy(errorReader)
	if err != nil {
		// 检查错误类型
		if contextErr, ok := err.(*ContextError); ok {
			switch contextErr.Code {
			case "READ_ERROR":
				ctx.JSON(500, map[string]any{
					"error":              "数据读取失败",
					"code":               contextErr.Code,
					"bytes_before_error": bytesWritten,
				})
			case "WRITE_ERROR":
				ctx.JSON(500, map[string]any{
					"error":              "数据写入失败",
					"code":               contextErr.Code,
					"bytes_before_error": bytesWritten,
				})
			case "NIL_READER":
				ctx.JSON(400, map[string]any{
					"error": "无效的数据源",
					"code":  contextErr.Code,
				})
			default:
				ctx.JSON(500, map[string]any{
					"error":   "未知错误",
					"code":    contextErr.Code,
					"message": contextErr.Message,
				})
			}
		} else {
			// 其他类型的错误
			ctx.JSON(500, map[string]any{
				"error":   "系统错误",
				"message": err.Error(),
			})
		}
		return
	}

	// 成功处理
	ctx.JSON(200, map[string]any{
		"success":           true,
		"bytes_transmitted": bytesWritten,
	})
}

// ExamplePerformanceMonitoring 演示性能监控
func ExamplePerformanceMonitoring(ctx *Context) {
	data := strings.Repeat("性能测试数据 ", 1000)
	reader := strings.NewReader(data)

	// 记录开始时间（在实际应用中可以使用更精确的时间测量）
	startTime := getCurrentTime()

	bytesWritten, err := ctx.Copy(reader)
	if err != nil {
		ctx.JSON(500, map[string]any{
			"error": "数据传输失败",
		})
		return
	}

	// 计算传输时间和速度
	endTime := getCurrentTime()
	duration := endTime - startTime
	throughput := float64(bytesWritten) / float64(duration) * 1000 // bytes per second

	// 返回性能统计
	ctx.JSON(200, map[string]any{
		"success": true,
		"statistics": map[string]any{
			"bytes_transmitted": bytesWritten,
			"duration_ms":       duration,
			"throughput_bps":    throughput,
			"data_size_kb":      float64(bytesWritten) / 1024,
		},
	})
}

// ErrorReader 用于测试错误处理的模拟Reader
type ErrorReader struct {
	shouldFail bool
	count      int
}

func (er *ErrorReader) Read(p []byte) (n int, err error) {
	if er.shouldFail && er.count > 0 {
		return 0, io.ErrUnexpectedEOF
	}

	if er.count >= 3 {
		return 0, io.EOF
	}

	// 写入一些测试数据
	testData := "test data chunk " + string(rune('A'+er.count))
	copy(p, testData)
	er.count++

	return len(testData), nil
}

/*
使用示例总结:

1. 基本Copy使用:
   bytesWritten, err := ctx.Copy(reader)

2. 带Content-Type的Copy:
   bytesWritten, err := ctx.CopyWithContentType(reader, "application/json")

3. 自定义缓冲区Copy:
   buf := make([]byte, 64*1024)
   bytesWritten, err := ctx.CopyBuffer(reader, buf)

4. 流式Copy:
   bytesWritten, err := ctx.StreamCopy(reader)

5. 错误处理:
   if contextErr, ok := err.(*ContextError); ok {
       // 处理特定的上下文错误
   }

性能建议:
- 对于大文件使用CopyBuffer并提供适当大小的缓冲区
- 对于实时数据使用StreamCopy
- 对于小数据使用默认的Copy方法
- 总是检查错误并适当处理
- 考虑设置适当的Content-Type和其他HTTP头部
*/
