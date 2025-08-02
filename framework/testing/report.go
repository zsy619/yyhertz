package testing

import (
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/zsy619/yyhertz/framework/config"
)

// TestResult 测试结果
type TestResult struct {
	Name       string         `json:"name"`
	Package    string         `json:"package"`
	Status     TestStatus     `json:"status"`
	Duration   time.Duration  `json:"duration"`
	StartTime  time.Time      `json:"start_time"`
	EndTime    time.Time      `json:"end_time"`
	Error      string         `json:"error,omitempty"`
	Output     string         `json:"output,omitempty"`
	Assertions int            `json:"assertions"`
	Failures   []TestFailure  `json:"failures,omitempty"`
	Tags       []string       `json:"tags,omitempty"`
	Metadata   map[string]any `json:"metadata,omitempty"`
}

// TestStatus 测试状态
type TestStatus string

const (
	TestStatusPass    TestStatus = "PASS"
	TestStatusFail    TestStatus = "FAIL"
	TestStatusSkip    TestStatus = "SKIP"
	TestStatusPending TestStatus = "PENDING"
)

// TestFailure 测试失败信息
type TestFailure struct {
	Message   string `json:"message"`
	File      string `json:"file"`
	Line      int    `json:"line"`
	Expected  string `json:"expected,omitempty"`
	Actual    string `json:"actual,omitempty"`
	Assertion string `json:"assertion"`
}

// TestSuiteResult 测试套件结果
type TestSuiteResult struct {
	Name         string        `json:"name"`
	Package      string        `json:"package"`
	Tests        []*TestResult `json:"tests"`
	Duration     time.Duration `json:"duration"`
	StartTime    time.Time     `json:"start_time"`
	EndTime      time.Time     `json:"end_time"`
	PassedCount  int           `json:"passed_count"`
	FailedCount  int           `json:"failed_count"`
	SkippedCount int           `json:"skipped_count"`
	TotalCount   int           `json:"total_count"`
}

// TestReport 测试报告
type TestReport struct {
	Suites       []*TestSuiteResult `json:"suites"`
	Duration     time.Duration      `json:"duration"`
	StartTime    time.Time          `json:"start_time"`
	EndTime      time.Time          `json:"end_time"`
	PassedCount  int                `json:"passed_count"`
	FailedCount  int                `json:"failed_count"`
	SkippedCount int                `json:"skipped_count"`
	TotalCount   int                `json:"total_count"`
	Coverage     *Coverage          `json:"coverage,omitempty"`
	Environment  *Environment       `json:"environment"`
	Config       *ReportConfig      `json:"config"`
}

// Coverage 代码覆盖率信息
type Coverage struct {
	TotalLines      int                         `json:"total_lines"`
	CoveredLines    int                         `json:"covered_lines"`
	Percentage      float64                     `json:"percentage"`
	PackageCoverage map[string]*PackageCoverage `json:"package_coverage,omitempty"`
}

// PackageCoverage 包覆盖率
type PackageCoverage struct {
	Package      string                   `json:"package"`
	TotalLines   int                      `json:"total_lines"`
	CoveredLines int                      `json:"covered_lines"`
	Percentage   float64                  `json:"percentage"`
	Files        map[string]*FileCoverage `json:"files,omitempty"`
}

// FileCoverage 文件覆盖率
type FileCoverage struct {
	File         string  `json:"file"`
	TotalLines   int     `json:"total_lines"`
	CoveredLines int     `json:"covered_lines"`
	Percentage   float64 `json:"percentage"`
	Lines        []int   `json:"covered_lines_detail,omitempty"`
}

// Environment 环境信息
type Environment struct {
	GoVersion    string            `json:"go_version"`
	OS           string            `json:"os"`
	Architecture string            `json:"architecture"`
	Hostname     string            `json:"hostname"`
	Timestamp    time.Time         `json:"timestamp"`
	Environment  map[string]string `json:"environment,omitempty"`
}

// ReportConfig 报告配置
type ReportConfig struct {
	OutputDir       string   `json:"output_dir"`
	Formats         []string `json:"formats"`
	IncludeCoverage bool     `json:"include_coverage"`
	IncludeOutput   bool     `json:"include_output"`
	Verbose         bool     `json:"verbose"`
	Title           string   `json:"title"`
	Description     string   `json:"description"`
}

// ReportGenerator 报告生成器
type ReportGenerator struct {
	config  *ReportConfig
	report  *TestReport
	mutex   sync.RWMutex
	writers map[string]ReportWriter
}

// ReportWriter 报告写入器接口
type ReportWriter interface {
	Write(report *TestReport, outputPath string) error
	Extension() string
	ContentType() string
}

// NewReportGenerator 创建报告生成器
func NewReportGenerator(config *ReportConfig) *ReportGenerator {
	if config == nil {
		config = DefaultReportConfig()
	}

	rg := &ReportGenerator{
		config:  config,
		writers: make(map[string]ReportWriter),
	}

	// 注册默认的报告写入器
	rg.RegisterWriter("json", &JSONReportWriter{})
	rg.RegisterWriter("html", &HTMLReportWriter{})
	rg.RegisterWriter("xml", &XMLReportWriter{})
	rg.RegisterWriter("text", &TextReportWriter{})

	return rg
}

// DefaultReportConfig 默认报告配置
func DefaultReportConfig() *ReportConfig {
	return &ReportConfig{
		OutputDir:       "./test-reports",
		Formats:         []string{"html", "json"},
		IncludeCoverage: true,
		IncludeOutput:   false,
		Verbose:         false,
		Title:           "Test Report",
		Description:     "Generated by YYHertz Testing Framework",
	}
}

// RegisterWriter 注册报告写入器
func (rg *ReportGenerator) RegisterWriter(format string, writer ReportWriter) {
	rg.mutex.Lock()
	defer rg.mutex.Unlock()
	rg.writers[format] = writer
}

// StartReport 开始报告
func (rg *ReportGenerator) StartReport() {
	rg.mutex.Lock()
	defer rg.mutex.Unlock()

	rg.report = &TestReport{
		Suites:      make([]*TestSuiteResult, 0),
		StartTime:   time.Now(),
		Environment: rg.collectEnvironmentInfo(),
		Config:      rg.config,
	}
}

// AddSuite 添加测试套件
func (rg *ReportGenerator) AddSuite(suite *TestSuiteResult) {
	rg.mutex.Lock()
	defer rg.mutex.Unlock()

	if rg.report == nil {
		rg.StartReport()
	}

	rg.report.Suites = append(rg.report.Suites, suite)
}

// AddTest 添加测试结果
func (rg *ReportGenerator) AddTest(suiteName string, test *TestResult) {
	rg.mutex.Lock()
	defer rg.mutex.Unlock()

	if rg.report == nil {
		rg.StartReport()
	}

	// 查找或创建测试套件
	var suite *TestSuiteResult
	for _, s := range rg.report.Suites {
		if s.Name == suiteName {
			suite = s
			break
		}
	}

	if suite == nil {
		suite = &TestSuiteResult{
			Name:      suiteName,
			Package:   test.Package,
			Tests:     make([]*TestResult, 0),
			StartTime: time.Now(),
		}
		rg.report.Suites = append(rg.report.Suites, suite)
	}

	suite.Tests = append(suite.Tests, test)
}

// FinishReport 完成报告
func (rg *ReportGenerator) FinishReport() *TestReport {
	rg.mutex.Lock()
	defer rg.mutex.Unlock()

	if rg.report == nil {
		return nil
	}

	rg.report.EndTime = time.Now()
	rg.report.Duration = rg.report.EndTime.Sub(rg.report.StartTime)

	// 计算统计信息
	rg.calculateStatistics()

	return rg.report
}

// GenerateReport 生成报告
func (rg *ReportGenerator) GenerateReport() error {
	report := rg.FinishReport()
	if report == nil {
		return fmt.Errorf("no report data available")
	}

	// 创建输出目录
	if err := os.MkdirAll(rg.config.OutputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// 生成各种格式的报告
	for _, format := range rg.config.Formats {
		if writer, exists := rg.writers[format]; exists {
			filename := fmt.Sprintf("test-report.%s", writer.Extension())
			outputPath := filepath.Join(rg.config.OutputDir, filename)

			if err := writer.Write(report, outputPath); err != nil {
				config.Errorf("Failed to write %s report: %v", format, err)
				continue
			}

			config.Infof("Generated %s report: %s", format, outputPath)
		} else {
			config.Warnf("Unknown report format: %s", format)
		}
	}

	return nil
}

// calculateStatistics 计算统计信息
func (rg *ReportGenerator) calculateStatistics() {
	for _, suite := range rg.report.Suites {
		suite.TotalCount = len(suite.Tests)
		suite.PassedCount = 0
		suite.FailedCount = 0
		suite.SkippedCount = 0

		var minStart time.Time
		var maxEnd time.Time

		for i, test := range suite.Tests {
			switch test.Status {
			case TestStatusPass:
				suite.PassedCount++
			case TestStatusFail:
				suite.FailedCount++
			case TestStatusSkip:
				suite.SkippedCount++
			}

			if i == 0 || test.StartTime.Before(minStart) {
				minStart = test.StartTime
			}
			if i == 0 || test.EndTime.After(maxEnd) {
				maxEnd = test.EndTime
			}
		}

		if suite.TotalCount > 0 {
			suite.StartTime = minStart
			suite.EndTime = maxEnd
			suite.Duration = maxEnd.Sub(minStart)
		}

		// 累计到总报告
		rg.report.TotalCount += suite.TotalCount
		rg.report.PassedCount += suite.PassedCount
		rg.report.FailedCount += suite.FailedCount
		rg.report.SkippedCount += suite.SkippedCount
	}
}

// collectEnvironmentInfo 收集环境信息
func (rg *ReportGenerator) collectEnvironmentInfo() *Environment {
	return &Environment{
		GoVersion:    "go1.21", // 简化实现
		OS:           "linux",
		Architecture: "amd64",
		Hostname:     "localhost",
		Timestamp:    time.Now(),
		Environment:  make(map[string]string),
	}
}

// ============= 报告写入器实现 =============

// JSONReportWriter JSON报告写入器
type JSONReportWriter struct{}

func (jw *JSONReportWriter) Write(report *TestReport, outputPath string) error {
	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(report)
}

func (jw *JSONReportWriter) Extension() string {
	return "json"
}

func (jw *JSONReportWriter) ContentType() string {
	return "application/json"
}

// HTMLReportWriter HTML报告写入器
type HTMLReportWriter struct{}

func (hw *HTMLReportWriter) Write(report *TestReport, outputPath string) error {
	tmpl := `<!DOCTYPE html>
<html>
<head>
    <title>{{.Config.Title}}</title>
    <meta charset="utf-8">
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; }
        .header { background: #f5f5f5; padding: 20px; border-radius: 5px; }
        .summary { display: flex; gap: 20px; margin: 20px 0; }
        .stat { background: #e9e9e9; padding: 15px; border-radius: 5px; text-align: center; }
        .stat.passed { background: #d4edda; }
        .stat.failed { background: #f8d7da; }
        .stat.skipped { background: #fff3cd; }
        .suite { margin: 20px 0; border: 1px solid #ddd; border-radius: 5px; }
        .suite-header { background: #f8f9fa; padding: 15px; font-weight: bold; }
        .test { padding: 10px 15px; border-bottom: 1px solid #eee; }
        .test:last-child { border-bottom: none; }
        .test.passed { background: #f8fff8; }
        .test.failed { background: #fff8f8; }
        .test.skipped { background: #fffef8; }
        .duration { color: #666; font-size: 0.9em; }
        .error { color: #d32f2f; margin-top: 10px; font-family: monospace; }
    </style>
</head>
<body>
    <div class="header">
        <h1>{{.Config.Title}}</h1>
        <p>{{.Config.Description}}</p>
        <p><strong>Generated:</strong> {{.Environment.Timestamp.Format "2006-01-02 15:04:05"}}</p>
        <p><strong>Duration:</strong> {{.Duration}}</p>
    </div>
    
    <div class="summary">
        <div class="stat passed">
            <div style="font-size: 2em;">{{.PassedCount}}</div>
            <div>Passed</div>
        </div>
        <div class="stat failed">
            <div style="font-size: 2em;">{{.FailedCount}}</div>
            <div>Failed</div>
        </div>
        <div class="stat skipped">
            <div style="font-size: 2em;">{{.SkippedCount}}</div>
            <div>Skipped</div>
        </div>
        <div class="stat">
            <div style="font-size: 2em;">{{.TotalCount}}</div>
            <div>Total</div>
        </div>
    </div>
    
    {{range .Suites}}
    <div class="suite">
        <div class="suite-header">
            {{.Name}} ({{.Package}})
            <span class="duration">{{.Duration}}</span>
        </div>
        {{range .Tests}}
        <div class="test {{.Status | lower}}">
            <strong>{{.Name}}</strong>
            <span class="duration">{{.Duration}}</span>
            {{if eq .Status "FAIL"}}
                {{range .Failures}}
                <div class="error">{{.Message}}</div>
                {{end}}
            {{end}}
        </div>
        {{end}}
    </div>
    {{end}}
</body>
</html>`

	funcMap := template.FuncMap{
		"lower": strings.ToLower,
	}

	t, err := template.New("report").Funcs(funcMap).Parse(tmpl)
	if err != nil {
		return err
	}

	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	return t.Execute(file, report)
}

func (hw *HTMLReportWriter) Extension() string {
	return "html"
}

func (hw *HTMLReportWriter) ContentType() string {
	return "text/html"
}

// XMLReportWriter XML报告写入器 (JUnit格式)
type XMLReportWriter struct{}

func (xw *XMLReportWriter) Write(report *TestReport, outputPath string) error {
	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	// 简化的JUnit XML格式
	file.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")
	file.WriteString(fmt.Sprintf(`<testsuites tests="%d" failures="%d" time="%.3f">`+"\n",
		report.TotalCount, report.FailedCount, report.Duration.Seconds()))

	for _, suite := range report.Suites {
		file.WriteString(fmt.Sprintf(`  <testsuite name="%s" tests="%d" failures="%d" time="%.3f">`+"\n",
			suite.Name, suite.TotalCount, suite.FailedCount, suite.Duration.Seconds()))

		for _, test := range suite.Tests {
			file.WriteString(fmt.Sprintf(`    <testcase name="%s" classname="%s" time="%.3f"`,
				test.Name, test.Package, test.Duration.Seconds()))

			if test.Status == TestStatusFail {
				file.WriteString(">\n")
				for _, failure := range test.Failures {
					file.WriteString(fmt.Sprintf(`      <failure message="%s">%s</failure>`+"\n",
						failure.Message, failure.Message))
				}
				file.WriteString("    </testcase>\n")
			} else if test.Status == TestStatusSkip {
				file.WriteString(">\n")
				file.WriteString("      <skipped/>\n")
				file.WriteString("    </testcase>\n")
			} else {
				file.WriteString("/>\n")
			}
		}

		file.WriteString("  </testsuite>\n")
	}

	file.WriteString("</testsuites>\n")
	return nil
}

func (xw *XMLReportWriter) Extension() string {
	return "xml"
}

func (xw *XMLReportWriter) ContentType() string {
	return "application/xml"
}

// TextReportWriter 文本报告写入器
type TextReportWriter struct{}

func (tw *TextReportWriter) Write(report *TestReport, outputPath string) error {
	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	// 报告头部
	file.WriteString(fmt.Sprintf("=== %s ===\n", report.Config.Title))
	file.WriteString(fmt.Sprintf("%s\n", report.Config.Description))
	file.WriteString(fmt.Sprintf("Generated: %s\n", report.Environment.Timestamp.Format("2006-01-02 15:04:05")))
	file.WriteString(fmt.Sprintf("Duration: %v\n\n", report.Duration))

	// 统计摘要
	file.WriteString("=== SUMMARY ===\n")
	file.WriteString(fmt.Sprintf("Total Tests: %d\n", report.TotalCount))
	file.WriteString(fmt.Sprintf("Passed: %d\n", report.PassedCount))
	file.WriteString(fmt.Sprintf("Failed: %d\n", report.FailedCount))
	file.WriteString(fmt.Sprintf("Skipped: %d\n", report.SkippedCount))
	file.WriteString("\n")

	// 详细结果
	for _, suite := range report.Suites {
		file.WriteString(fmt.Sprintf("=== %s (%s) ===\n", suite.Name, suite.Package))
		file.WriteString(fmt.Sprintf("Duration: %v\n", suite.Duration))
		file.WriteString(fmt.Sprintf("Tests: %d, Passed: %d, Failed: %d, Skipped: %d\n\n",
			suite.TotalCount, suite.PassedCount, suite.FailedCount, suite.SkippedCount))

		// 排序测试结果（失败的在前）
		tests := make([]*TestResult, len(suite.Tests))
		copy(tests, suite.Tests)
		sort.Slice(tests, func(i, j int) bool {
			if tests[i].Status != tests[j].Status {
				return tests[i].Status == TestStatusFail
			}
			return tests[i].Name < tests[j].Name
		})

		for _, test := range tests {
			status := string(test.Status)
			file.WriteString(fmt.Sprintf("  [%s] %s (%v)\n", status, test.Name, test.Duration))

			if test.Status == TestStatusFail {
				for _, failure := range test.Failures {
					file.WriteString(fmt.Sprintf("    ERROR: %s\n", failure.Message))
				}
			}
		}

		file.WriteString("\n")
	}

	return nil
}

func (tw *TextReportWriter) Extension() string {
	return "txt"
}

func (tw *TextReportWriter) ContentType() string {
	return "text/plain"
}

// ============= 报告工具函数 =============

// NewTestResult 创建测试结果
func NewTestResult(name, pkg string) *TestResult {
	return &TestResult{
		Name:       name,
		Package:    pkg,
		Status:     TestStatusPending,
		StartTime:  time.Now(),
		Assertions: 0,
		Failures:   make([]TestFailure, 0),
		Tags:       make([]string, 0),
		Metadata:   make(map[string]any),
	}
}

// AddFailure 添加失败信息
func (tr *TestResult) AddFailure(message, file string, line int, assertion string) {
	failure := TestFailure{
		Message:   message,
		File:      file,
		Line:      line,
		Assertion: assertion,
	}
	tr.Failures = append(tr.Failures, failure)
}

// SetPassed 设置为通过
func (tr *TestResult) SetPassed() {
	tr.Status = TestStatusPass
	tr.EndTime = time.Now()
	tr.Duration = tr.EndTime.Sub(tr.StartTime)
}

// SetFailed 设置为失败
func (tr *TestResult) SetFailed(err error) {
	tr.Status = TestStatusFail
	tr.EndTime = time.Now()
	tr.Duration = tr.EndTime.Sub(tr.StartTime)
	if err != nil {
		tr.Error = err.Error()
	}
}

// SetSkipped 设置为跳过
func (tr *TestResult) SetSkipped() {
	tr.Status = TestStatusSkip
	tr.EndTime = time.Now()
	tr.Duration = tr.EndTime.Sub(tr.StartTime)
}

// AddTag 添加标签
func (tr *TestResult) AddTag(tag string) {
	tr.Tags = append(tr.Tags, tag)
}

// SetMetadata 设置元数据
func (tr *TestResult) SetMetadata(key string, value any) {
	tr.Metadata[key] = value
}

// ============= 全局报告生成器 =============

var (
	globalReportGenerator *ReportGenerator
	reportGeneratorOnce   sync.Once
)

// GetGlobalReportGenerator 获取全局报告生成器
func GetGlobalReportGenerator() *ReportGenerator {
	reportGeneratorOnce.Do(func() {
		globalReportGenerator = NewReportGenerator(DefaultReportConfig())
	})
	return globalReportGenerator
}

// StartGlobalReport 开始全局报告
func StartGlobalReport() {
	GetGlobalReportGenerator().StartReport()
}

// AddGlobalTest 添加全局测试
func AddGlobalTest(suiteName string, test *TestResult) {
	GetGlobalReportGenerator().AddTest(suiteName, test)
}

// GenerateGlobalReport 生成全局报告
func GenerateGlobalReport() error {
	return GetGlobalReportGenerator().GenerateReport()
}
