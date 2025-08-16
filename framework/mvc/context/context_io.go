package context

import "io"

// ============= Writer Copy功能 =============

// Copy 从io.Reader复制数据到Context的Writer，类似io.Copy功能
// 这个方法提供了类似io.Copy(ctx.Writer(), reader)的功能
// 返回写入的字节数和可能的错误
func (ctx *Context) Copy(reader io.Reader) (int64, error) {
	if !ctx.ensureRequest() {
		return 0, ErrRequestNotFound
	}

	if reader == nil {
		return 0, &ContextError{
			Code:    "NIL_READER",
			Message: "Reader cannot be nil",
		}
	}

	// 使用缓冲区进行高效复制
	buf := make([]byte, 32*1024) // 32KB缓冲区，平衡性能和内存使用
	var totalWritten int64

	for {
		// 从reader读取数据
		bytesRead, readErr := reader.Read(buf)
		if bytesRead > 0 {
			// 写入到response
			if _, writeErr := ctx.writer.Write(buf[:bytesRead]); writeErr != nil {
				return totalWritten, &ContextError{
					Code:    "WRITE_ERROR",
					Message: "Failed to write to response: " + writeErr.Error(),
					Cause:   writeErr,
				}
			}
			totalWritten += int64(bytesRead)
		}

		// 检查读取错误
		if readErr != nil {
			if readErr == io.EOF {
				// 正常结束
				break
			}
			return totalWritten, &ContextError{
				Code:    "READ_ERROR",
				Message: "Failed to read from source: " + readErr.Error(),
				Cause:   readErr,
			}
		}
	}

	return totalWritten, nil
}

// CopyBuffer 使用指定缓冲区从io.Reader复制数据到Context的Writer
// 允许用户自定义缓冲区大小以优化性能
func (ctx *Context) CopyBuffer(reader io.Reader, buf []byte) (int64, error) {
	if !ctx.ensureRequest() {
		return 0, ErrRequestNotFound
	}

	if reader == nil {
		return 0, &ContextError{
			Code:    "NIL_READER",
			Message: "Reader cannot be nil",
		}
	}

	if nil == buf || len(buf) == 0 {
		// 如果没有提供缓冲区，使用默认的Copy方法
		return ctx.Copy(reader)
	}

	var totalWritten int64

	for {
		// 从reader读取数据到指定缓冲区
		bytesRead, readErr := reader.Read(buf)
		if bytesRead > 0 {
			// 写入到response
			if _, writeErr := ctx.writer.Write(buf[:bytesRead]); writeErr != nil {
				return totalWritten, &ContextError{
					Code:    "WRITE_ERROR",
					Message: "Failed to write to response: " + writeErr.Error(),
					Cause:   writeErr,
				}
			}
			totalWritten += int64(bytesRead)
		}

		// 检查读取错误
		if readErr != nil {
			if readErr == io.EOF {
				// 正常结束
				break
			}
			return totalWritten, &ContextError{
				Code:    "READ_ERROR",
				Message: "Failed to read from source: " + readErr.Error(),
				Cause:   readErr,
			}
		}
	}

	return totalWritten, nil
}

// CopyWithContentType 复制数据并设置Content-Type
// 这是一个便捷方法，将Copy和设置Content-Type结合
func (ctx *Context) CopyWithContentType(reader io.Reader, contentType string) (int64, error) {
	// 设置Content-Type
	ctx.SetContentType(contentType)

	// 执行复制
	return ctx.Copy(reader)
}

// StreamCopy 流式复制，支持实时写入（适合大文件或流数据）
// 每次写入后会刷新缓冲区，适合需要实时传输的场景
func (ctx *Context) StreamCopy(reader io.Reader) (int64, error) {
	if !ctx.ensureRequest() {
		return 0, ErrRequestNotFound
	}

	if reader == nil {
		return 0, &ContextError{
			Code:    "NIL_READER",
			Message: "Reader cannot be nil",
		}
	}

	// 对于流式传输，使用较小的缓冲区以减少延迟
	buf := make([]byte, 8*1024) // 8KB缓冲区
	var totalWritten int64

	for {
		// 从reader读取数据
		bytesRead, readErr := reader.Read(buf)
		if bytesRead > 0 {
			// 写入到response
			if _, writeErr := ctx.writer.Write(buf[:bytesRead]); writeErr != nil {
				return totalWritten, &ContextError{
					Code:    "WRITE_ERROR",
					Message: "Failed to write to response: " + writeErr.Error(),
					Cause:   writeErr,
				}
			}
			totalWritten += int64(bytesRead)

			// 流式传输时，每次写入后尝试刷新（如果writer支持）
			if flusher, ok := ctx.writer.(interface{ Flush() error }); ok {
				flusher.Flush()
			}
		}

		// 检查读取错误
		if readErr != nil {
			if readErr == io.EOF {
				// 正常结束
				break
			}
			return totalWritten, &ContextError{
				Code:    "READ_ERROR",
				Message: "Failed to read from source: " + readErr.Error(),
				Cause:   readErr,
			}
		}
	}

	return totalWritten, nil
}