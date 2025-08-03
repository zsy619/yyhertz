package config

import (
	"fmt"
	"io"
	"log/syslog"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
	lSyslog "github.com/sirupsen/logrus/hooks/syslog"
)

// OutputWriter 输出写入器接口
type OutputWriter interface {
	io.Writer
	Close() error
}

// ConsoleWriter 控制台输出写入器
type ConsoleWriter struct {
	writer io.Writer
}

// NewConsoleWriter 创建控制台写入器
func NewConsoleWriter() *ConsoleWriter {
	return &ConsoleWriter{
		writer: os.Stdout,
	}
}

func (w *ConsoleWriter) Write(p []byte) (n int, err error) {
	return w.writer.Write(p)
}

func (w *ConsoleWriter) Close() error {
	return nil // 控制台不需要关闭
}

// FileWriter 文件输出写入器
type FileWriter struct {
	file *os.File
	path string
}

// NewFileWriter 创建文件写入器
func NewFileWriter(path string) (*FileWriter, error) {
	// 确保目录存在
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, fmt.Errorf("创建日志目录失败: %w", err)
	}
	
	// 打开文件
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, fmt.Errorf("打开日志文件失败: %w", err)
	}
	
	return &FileWriter{
		file: file,
		path: path,
	}, nil
}

func (w *FileWriter) Write(p []byte) (n int, err error) {
	return w.file.Write(p)
}

func (w *FileWriter) Close() error {
	if w.file != nil {
		return w.file.Close()
	}
	return nil
}

// SyslogWriter Syslog输出写入器
type SyslogWriter struct {
	writer *syslog.Writer
	config SyslogConfig
}

// NewSyslogWriter 创建Syslog写入器
func NewSyslogWriter(config SyslogConfig) (*SyslogWriter, error) {
	writer, err := syslog.Dial(config.Network, config.Address, syslog.Priority(config.Priority), config.Tag)
	if err != nil {
		return nil, fmt.Errorf("连接Syslog失败: %w", err)
	}
	
	return &SyslogWriter{
		writer: writer,
		config: config,
	}, nil
}

func (w *SyslogWriter) Write(p []byte) (n int, err error) {
	// Syslog写入器会自动添加时间戳等信息，这里直接写入消息内容
	return w.writer.Write(p)
}

func (w *SyslogWriter) Close() error {
	if w.writer != nil {
		return w.writer.Close()
	}
	return nil
}

// FluentdWriter Fluentd输出写入器
type FluentdWriter struct {
	config FluentdConfig
	// 这里可以集成Fluentd客户端库
	// 暂时使用模拟实现
	buffer []byte
}

// NewFluentdWriter 创建Fluentd写入器
func NewFluentdWriter(config FluentdConfig) (*FluentdWriter, error) {
	// 验证配置
	if config.Host == "" {
		return nil, fmt.Errorf("Fluentd主机不能为空")
	}
	if config.Port <= 0 {
		return nil, fmt.Errorf("Fluentd端口必须大于0")
	}
	
	return &FluentdWriter{
		config: config,
		buffer: make([]byte, 0),
	}, nil
}

func (w *FluentdWriter) Write(p []byte) (n int, err error) {
	// 实际实现中应该连接到Fluentd服务器
	// 这里暂时模拟输出到标准输出
	fmt.Printf("[FLUENTD %s:%d] %s", w.config.Host, w.config.Port, string(p))
	return len(p), nil
}

func (w *FluentdWriter) Close() error {
	return nil
}

// CloudWatchWriter AWS CloudWatch输出写入器
type CloudWatchWriter struct {
	config CloudWatchConfig
	// 这里可以集成AWS SDK
	// 暂时使用模拟实现
}

// NewCloudWatchWriter 创建CloudWatch写入器
func NewCloudWatchWriter(config CloudWatchConfig) (*CloudWatchWriter, error) {
	// 验证配置
	if config.Region == "" {
		return nil, fmt.Errorf("AWS区域不能为空")
	}
	if config.LogGroupName == "" {
		return nil, fmt.Errorf("日志组名称不能为空")
	}
	
	return &CloudWatchWriter{
		config: config,
	}, nil
}

func (w *CloudWatchWriter) Write(p []byte) (n int, err error) {
	// 实际实现中应该调用AWS CloudWatch Logs API
	// 这里暂时模拟输出
	fmt.Printf("[CLOUDWATCH %s/%s] %s", w.config.LogGroupName, w.config.LogStreamName, string(p))
	return len(p), nil
}

func (w *CloudWatchWriter) Close() error {
	return nil
}

// AzureInsightsWriter Azure Application Insights输出写入器
type AzureInsightsWriter struct {
	config AzureInsightsConfig
	// 这里可以集成Azure SDK
}

// NewAzureInsightsWriter 创建Azure Insights写入器
func NewAzureInsightsWriter(config AzureInsightsConfig) (*AzureInsightsWriter, error) {
	// 验证配置
	if config.InstrumentationKey == "" {
		return nil, fmt.Errorf("Instrumentation Key不能为空")
	}
	
	return &AzureInsightsWriter{
		config: config,
	}, nil
}

func (w *AzureInsightsWriter) Write(p []byte) (n int, err error) {
	// 实际实现中应该调用Azure Application Insights API
	// 这里暂时模拟输出
	fmt.Printf("[AZURE_INSIGHTS %s] %s", w.config.InstrumentationKey[:8]+"...", string(p))
	return len(p), nil
}

func (w *AzureInsightsWriter) Close() error {
	return nil
}

// ElasticsearchWriter Elasticsearch输出写入器
type ElasticsearchWriter struct {
	config ElasticsearchConfig
	// 这里可以集成Elasticsearch客户端
}

// NewElasticsearchWriter 创建Elasticsearch写入器
func NewElasticsearchWriter(config ElasticsearchConfig) (*ElasticsearchWriter, error) {
	// 验证配置
	if len(config.URLs) == 0 {
		return nil, fmt.Errorf("Elasticsearch URLs不能为空")
	}
	if config.Index == "" {
		return nil, fmt.Errorf("索引名称不能为空")
	}
	
	return &ElasticsearchWriter{
		config: config,
	}, nil
}

func (w *ElasticsearchWriter) Write(p []byte) (n int, err error) {
	// 实际实现中应该调用Elasticsearch API
	// 这里暂时模拟输出
	fmt.Printf("[ELASTICSEARCH %s/%s] %s", w.config.URLs[0], w.config.Index, string(p))
	return len(p), nil
}

func (w *ElasticsearchWriter) Close() error {
	return nil
}

// KafkaWriter Kafka输出写入器
type KafkaWriter struct {
	config KafkaConfig
	// 这里可以集成Kafka客户端
}

// NewKafkaWriter 创建Kafka写入器
func NewKafkaWriter(config KafkaConfig) (*KafkaWriter, error) {
	// 验证配置
	if len(config.Brokers) == 0 {
		return nil, fmt.Errorf("Kafka brokers不能为空")
	}
	if config.Topic == "" {
		return nil, fmt.Errorf("Kafka topic不能为空")
	}
	
	return &KafkaWriter{
		config: config,
	}, nil
}

func (w *KafkaWriter) Write(p []byte) (n int, err error) {
	// 实际实现中应该发送消息到Kafka
	// 这里暂时模拟输出
	fmt.Printf("[KAFKA %s/%s] %s", w.config.Brokers[0], w.config.Topic, string(p))
	return len(p), nil
}

func (w *KafkaWriter) Close() error {
	return nil
}

// MultiWriter 多输出写入器
type MultiWriter struct {
	writers []OutputWriter
}

// NewMultiWriter 创建多输出写入器
func NewMultiWriter(writers ...OutputWriter) *MultiWriter {
	return &MultiWriter{
		writers: writers,
	}
}

func (w *MultiWriter) Write(p []byte) (n int, err error) {
	for _, writer := range w.writers {
		if writer != nil {
			_, writeErr := writer.Write(p)
			if writeErr != nil {
				// 记录错误但继续写入其他输出
				fmt.Printf("写入输出失败: %v\n", writeErr)
			}
		}
	}
	return len(p), nil
}

func (w *MultiWriter) Close() error {
	var lastErr error
	for _, writer := range w.writers {
		if writer != nil {
			if err := writer.Close(); err != nil {
				lastErr = err
			}
		}
	}
	return lastErr
}

// AddWriter 添加写入器
func (w *MultiWriter) AddWriter(writer OutputWriter) {
	w.writers = append(w.writers, writer)
}

// RemoveWriter 移除写入器
func (w *MultiWriter) RemoveWriter(index int) {
	if index >= 0 && index < len(w.writers) {
		w.writers = append(w.writers[:index], w.writers[index+1:]...)
	}
}

// GetWriters 获取所有写入器
func (w *MultiWriter) GetWriters() []OutputWriter {
	return w.writers
}

// CreateOutputWriters 根据配置创建输出写入器
func CreateOutputWriters(config *LogConfig) ([]OutputWriter, error) {
	var writers []OutputWriter
	
	// 处理传统的控制台和文件输出
	if config.EnableConsole {
		writers = append(writers, NewConsoleWriter())
	}
	
	if config.EnableFile && config.FilePath != "" {
		fileWriter, err := NewFileWriter(config.FilePath)
		if err != nil {
			return nil, fmt.Errorf("创建文件写入器失败: %w", err)
		}
		writers = append(writers, fileWriter)
	}
	
	// 处理扩展输出
	for _, output := range config.Outputs {
		switch output {
		case "console":
			// 已在上面处理
			continue
		case "file":
			// 已在上面处理
			continue
		case "syslog":
			if outputConfig, exists := config.GetOutputConfig("syslog"); exists {
				if syslogConfig, ok := outputConfig.(SyslogConfig); ok {
					writer, err := NewSyslogWriter(syslogConfig)
					if err != nil {
						fmt.Printf("创建Syslog写入器失败: %v\n", err)
						continue
					}
					writers = append(writers, writer)
				}
			}
		case "fluentd":
			if outputConfig, exists := config.GetOutputConfig("fluentd"); exists {
				if fluentdConfig, ok := outputConfig.(FluentdConfig); ok {
					writer, err := NewFluentdWriter(fluentdConfig)
					if err != nil {
						fmt.Printf("创建Fluentd写入器失败: %v\n", err)
						continue
					}
					writers = append(writers, writer)
				}
			}
		case "cloudwatch":
			if outputConfig, exists := config.GetOutputConfig("cloudwatch"); exists {
				if cwConfig, ok := outputConfig.(CloudWatchConfig); ok {
					writer, err := NewCloudWatchWriter(cwConfig)
					if err != nil {
						fmt.Printf("创建CloudWatch写入器失败: %v\n", err)
						continue
					}
					writers = append(writers, writer)
				}
			}
		case "azure_insights":
			if outputConfig, exists := config.GetOutputConfig("azure_insights"); exists {
				if aiConfig, ok := outputConfig.(AzureInsightsConfig); ok {
					writer, err := NewAzureInsightsWriter(aiConfig)
					if err != nil {
						fmt.Printf("创建Azure Insights写入器失败: %v\n", err)
						continue
					}
					writers = append(writers, writer)
				}
			}
		case "elasticsearch":
			if outputConfig, exists := config.GetOutputConfig("elasticsearch"); exists {
				if esConfig, ok := outputConfig.(ElasticsearchConfig); ok {
					writer, err := NewElasticsearchWriter(esConfig)
					if err != nil {
						fmt.Printf("创建Elasticsearch写入器失败: %v\n", err)
						continue
					}
					writers = append(writers, writer)
				}
			}
		case "kafka":
			if outputConfig, exists := config.GetOutputConfig("kafka"); exists {
				if kafkaConfig, ok := outputConfig.(KafkaConfig); ok {
					writer, err := NewKafkaWriter(kafkaConfig)
					if err != nil {
						fmt.Printf("创建Kafka写入器失败: %v\n", err)
						continue
					}
					writers = append(writers, writer)
				}
			}
		}
	}
	
	return writers, nil
}

// SetupLoggerHooks 设置日志钩子
func SetupLoggerHooks(logger *logrus.Logger, config *LogConfig) error {
	// 设置Syslog钩子
	if config.HasOutput("syslog") {
		if outputConfig, exists := config.GetOutputConfig("syslog"); exists {
			if syslogConfig, ok := outputConfig.(SyslogConfig); ok {
				hook, err := lSyslog.NewSyslogHook(syslogConfig.Network, syslogConfig.Address, 
					syslog.Priority(syslogConfig.Priority), syslogConfig.Tag)
				if err != nil {
					fmt.Printf("创建Syslog钩子失败: %v\n", err)
				} else {
					logger.AddHook(hook)
				}
			}
		}
	}
	
	// 可以添加更多钩子
	// 如：Fluentd钩子、CloudWatch钩子等
	
	return nil
}