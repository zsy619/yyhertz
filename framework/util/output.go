package util

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"strings"
	"sync"
)

// Output buffer management
var (
	outputBuffers   [][]byte
	outputBufferMux sync.Mutex
	outputLevel     int
	outputHandlers  []func(string, int) string
)

// Output control constants
const (
	PHP_OUTPUT_HANDLER_START = 1
	PHP_OUTPUT_HANDLER_WRITE = 0
	PHP_OUTPUT_HANDLER_FLUSH = 4
	PHP_OUTPUT_HANDLER_CLEAN = 2
	PHP_OUTPUT_HANDLER_FINAL = 8
	PHP_OUTPUT_HANDLER_CONT  = 0
	PHP_OUTPUT_HANDLER_END   = 8

	PHP_OUTPUT_HANDLER_CLEANABLE = 16
	PHP_OUTPUT_HANDLER_FLUSHABLE = 32
	PHP_OUTPUT_HANDLER_REMOVABLE = 64
	PHP_OUTPUT_HANDLER_STDFLAGS  = 112
)

// ObStart turns on output buffering
func ObStart(outputCallback ...func(string, int) string) bool {
	outputBufferMux.Lock()
	defer outputBufferMux.Unlock()

	// Add a new buffer level
	outputBuffers = append(outputBuffers, []byte{})
	outputLevel++

	// Add callback if provided
	if len(outputCallback) > 0 && outputCallback[0] != nil {
		outputHandlers = append(outputHandlers, outputCallback[0])
	} else {
		outputHandlers = append(outputHandlers, nil)
	}

	return true
}

// ObGetContents returns the contents of the output buffer
func ObGetContents() string {
	outputBufferMux.Lock()
	defer outputBufferMux.Unlock()

	if outputLevel == 0 {
		return ""
	}

	return string(outputBuffers[outputLevel-1])
}

// ObGetLength returns the length of the output buffer
func ObGetLength() int {
	outputBufferMux.Lock()
	defer outputBufferMux.Unlock()

	if outputLevel == 0 {
		return 0
	}

	return len(outputBuffers[outputLevel-1])
}

// ObClean cleans the output buffer
func ObClean() bool {
	outputBufferMux.Lock()
	defer outputBufferMux.Unlock()

	if outputLevel == 0 {
		return false
	}

	outputBuffers[outputLevel-1] = []byte{}
	return true
}

// ObFlush flushes the output buffer
func ObFlush() bool {
	outputBufferMux.Lock()
	defer outputBufferMux.Unlock()

	if outputLevel == 0 {
		return false
	}

	content := string(outputBuffers[outputLevel-1])

	// Apply handler if exists
	if outputHandlers[outputLevel-1] != nil {
		content = outputHandlers[outputLevel-1](content, PHP_OUTPUT_HANDLER_FLUSH)
	}

	// In a real implementation, this would output to the response
	fmt.Print(content)

	// Clear the buffer
	outputBuffers[outputLevel-1] = []byte{}

	return true
}

// ObEndFlush flushes and turns off output buffering
func ObEndFlush() bool {
	outputBufferMux.Lock()
	defer outputBufferMux.Unlock()

	if outputLevel == 0 {
		return false
	}

	content := string(outputBuffers[outputLevel-1])

	// Apply handler if exists
	if outputHandlers[outputLevel-1] != nil {
		content = outputHandlers[outputLevel-1](content, PHP_OUTPUT_HANDLER_FINAL)
	}

	// In a real implementation, this would output to the response
	fmt.Print(content)

	// Remove the buffer level
	outputBuffers = outputBuffers[:outputLevel-1]
	outputHandlers = outputHandlers[:outputLevel-1]
	outputLevel--

	return true
}

// ObEndClean cleans and turns off output buffering
func ObEndClean() bool {
	outputBufferMux.Lock()
	defer outputBufferMux.Unlock()

	if outputLevel == 0 {
		return false
	}

	// Remove the buffer level without outputting
	outputBuffers = outputBuffers[:outputLevel-1]
	outputHandlers = outputHandlers[:outputLevel-1]
	outputLevel--

	return true
}

// ObGetClean gets the current buffer contents and cleans the buffer
func ObGetClean() string {
	content := ObGetContents()
	ObClean()
	return content
}

// ObGetFlush gets the current buffer contents and flushes the buffer
func ObGetFlush() string {
	content := ObGetContents()
	ObFlush()
	return content
}

// ObGetLevel returns the nesting level of output buffering
func ObGetLevel() int {
	outputBufferMux.Lock()
	defer outputBufferMux.Unlock()

	return outputLevel
}

// ObGetStatus gets the status of output buffers
func ObGetStatus(fullStatus ...bool) any {
	outputBufferMux.Lock()
	defer outputBufferMux.Unlock()

	if len(fullStatus) > 0 && fullStatus[0] {
		// Return full status array
		var status []map[string]any
		for i := 0; i < outputLevel; i++ {
			bufferStatus := map[string]any{
				"level":      i,
				"type":       0,
				"status":     0,
				"name":       "default output handler",
				"del":        true,
				"chunk_size": 0,
			}
			status = append(status, bufferStatus)
		}
		return status
	}

	// Return simple status
	if outputLevel == 0 {
		return map[string]any{
			"level":      0,
			"type":       0,
			"status":     0,
			"name":       "default output handler",
			"del":        true,
			"chunk_size": 0,
		}
	}

	return map[string]any{
		"level":      outputLevel - 1,
		"type":       0,
		"status":     0,
		"name":       "default output handler",
		"del":        true,
		"chunk_size": 0,
	}
}

// ObListHandlers lists all output handlers
func ObListHandlers() []string {
	outputBufferMux.Lock()
	defer outputBufferMux.Unlock()

	var handlers []string
	for i := 0; i < outputLevel; i++ {
		if outputHandlers[i] != nil {
			handlers = append(handlers, "callback function")
		} else {
			handlers = append(handlers, "default output handler")
		}
	}

	return handlers
}

// Output adds content to the current output buffer
func Output(content string) {
	outputBufferMux.Lock()
	defer outputBufferMux.Unlock()

	if outputLevel > 0 {
		outputBuffers[outputLevel-1] = append(outputBuffers[outputLevel-1], []byte(content)...)
	} else {
		// Direct output if no buffering
		fmt.Print(content)
	}
}

// Echo outputs one or more strings
func Echo(values ...any) {
	var content strings.Builder
	for _, v := range values {
		content.WriteString(fmt.Sprintf("%v", v))
	}
	Output(content.String())
}

// Print outputs a string
func Print(value any) int {
	str := fmt.Sprintf("%v", value)
	Output(str)
	return 1
}

// Printf outputs a formatted string
func Printf(format string, args ...any) int {
	str := fmt.Sprintf(format, args...)
	Output(str)
	return len(str)
}

// Vprintf outputs a formatted string using an array of arguments
func Vprintf(format string, args []any) int {
	str := fmt.Sprintf(format, args...)
	Output(str)
	return len(str)
}

// Flush flushes the output
func Flush() {
	if outputLevel > 0 {
		ObFlush()
	}
	// In a real implementation, this would flush the HTTP response
}

// Built-in output handlers

// ObGzhandler compresses output with gzip
func ObGzhandler(buffer string, mode int) string {
	if mode&PHP_OUTPUT_HANDLER_FINAL == 0 {
		return buffer
	}

	var compressed bytes.Buffer
	gz := gzip.NewWriter(&compressed)

	_, err := gz.Write([]byte(buffer))
	if err != nil {
		return buffer
	}

	err = gz.Close()
	if err != nil {
		return buffer
	}

	return compressed.String()
}

// ObDeflateHandler compresses output with deflate (simplified version)
func ObDeflateHandler(buffer string, mode int) string {
	// Simplified implementation - in practice would use deflate compression
	return buffer
}

// ObTidyHandler tidies/cleans HTML output (simplified version)
func ObTidyHandler(buffer string, mode int) string {
	// Simplified HTML cleaning
	cleaned := strings.ReplaceAll(buffer, "\n\n", "\n")
	cleaned = strings.TrimSpace(cleaned)
	return cleaned
}

// ReadFile reads and outputs a file
func ReadFile(filename string, useIncludePath ...bool) (int, error) {
	content, err := FileGetContents(filename)
	if err != nil {
		return 0, err
	}

	Output(content)
	return len(content), nil
}

// OutputBuffering utility functions

// ObImplicitFlush turns implicit flushing on/off
func ObImplicitFlush(flag ...bool) {
	// In a real implementation, this would control automatic flushing
	// This is a placeholder
}

// ObGetHandler returns the current output handler (simplified)
func ObGetHandler() any {
	outputBufferMux.Lock()
	defer outputBufferMux.Unlock()

	if outputLevel == 0 {
		return nil
	}

	return outputHandlers[outputLevel-1]
}

// Advanced output control functions

// OutputAddRewriteVar adds URL rewriter values
func OutputAddRewriteVar(name, value string) bool {
	// In a real implementation, this would add variables for URL rewriting
	// This is a placeholder
	return true
}

// OutputResetRewriteVars resets URL rewriter values
func OutputResetRewriteVars() bool {
	// In a real implementation, this would reset URL rewriting variables
	// This is a placeholder
	return true
}

// Content processing utilities

// ProcessOutputBuffer processes the output buffer with a callback
func ProcessOutputBuffer(processor func(string) string) {
	content := ObGetContents()
	if content != "" {
		processed := processor(content)
		ObClean()
		Output(processed)
	}
}

// MinifyHTML removes unnecessary whitespace from HTML
func MinifyHTML(html string) string {
	// Simple HTML minification
	lines := strings.Split(html, "\n")
	var minified []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			minified = append(minified, trimmed)
		}
	}

	return strings.Join(minified, "")
}

// CompressOutput compresses output using gzip
func CompressOutput() {
	if outputLevel > 0 {
		ObStart(ObGzhandler)
	}
}

// CacheOutput caches output to a file
func CacheOutput(filename string) {
	content := ObGetContents()
	if content != "" {
		FilePutContents(filename, content)
	}
}

// StreamOutput streams output in chunks
func StreamOutput(content string, chunkSize int) {
	for len(content) > 0 {
		if len(content) <= chunkSize {
			Output(content)
			break
		}

		Output(content[:chunkSize])
		content = content[chunkSize:]
		Flush()

		// In a real implementation, might add delay here
	}
}

// Template output functions

// IncludeTemplate includes and processes a template file
func IncludeTemplate(filename string, variables map[string]any) error {
	template, err := FileGetContents(filename)
	if err != nil {
		return err
	}

	// Simple template processing - replace {{variable}} with values
	for key, value := range variables {
		placeholder := "{{" + key + "}}"
		replacement := fmt.Sprintf("%v", value)
		template = strings.ReplaceAll(template, placeholder, replacement)
	}

	Output(template)
	return nil
}

// RenderTemplate renders a template with given data
func RenderTemplate(template string, data map[string]any) string {
	result := template

	for key, value := range data {
		placeholder := "{{" + key + "}}"
		replacement := fmt.Sprintf("%v", value)
		result = strings.ReplaceAll(result, placeholder, replacement)
	}

	return result
}

// Content type and encoding functions

// SetContentType sets the content type header (placeholder)
func SetContentType(contentType string) {
	// In a real implementation, this would set HTTP headers
	fmt.Printf("Content-Type: %s\n", contentType)
}

// SetCharset sets the character set (placeholder)
func SetCharset(charset string) {
	// In a real implementation, this would set encoding
	fmt.Printf("Charset: %s\n", charset)
}

// Error output functions

// ErrorOutput outputs an error message
func ErrorOutput(message string) {
	// In a real implementation, this might go to error log
	Output("Error: " + message + "\n")
}

// DebugOutput outputs debug information
func DebugOutput(value any) {
	debug := fmt.Sprintf("DEBUG: %+v\n", value)
	Output(debug)
}

// LogOutput logs output to a file
func LogOutput(message string, filename string) error {
	timestamp := Date("Y-m-d H:i:s")
	logEntry := fmt.Sprintf("[%s] %s\n", timestamp, message)

	return FilePutContents(filename, logEntry, 8) // FILE_APPEND
}

// Buffering utilities

// GetBufferedOutput gets all buffered output at all levels
func GetBufferedOutput() []string {
	outputBufferMux.Lock()
	defer outputBufferMux.Unlock()

	var buffers []string
	for i := 0; i < outputLevel; i++ {
		buffers = append(buffers, string(outputBuffers[i]))
	}

	return buffers
}

// ClearAllBuffers clears all output buffers
func ClearAllBuffers() {
	for outputLevel > 0 {
		ObEndClean()
	}
}

// FlushAllBuffers flushes all output buffers
func FlushAllBuffers() {
	for outputLevel > 0 {
		ObEndFlush()
	}
}

// Capture output from a function
func CaptureOutput(fn func()) string {
	ObStart()
	fn()
	return ObGetClean()
}

// OutputWriter implements io.Writer interface for output buffering
type OutputWriter struct{}

func (ow *OutputWriter) Write(p []byte) (n int, err error) {
	Output(string(p))
	return len(p), nil
}

// GetOutputWriter returns an io.Writer that writes to the output buffer
func GetOutputWriter() io.Writer {
	return &OutputWriter{}
}
