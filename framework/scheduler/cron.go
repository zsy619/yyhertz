package scheduler

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/zsy619/yyhertz/framework/config"
)

// CronExpression Cron表达式
type CronExpression struct {
	Second     []int // 0-59
	Minute     []int // 0-59
	Hour       []int // 0-23
	DayOfMonth []int // 1-31
	Month      []int // 1-12
	DayOfWeek  []int // 0-6 (Sunday = 0)
	Year       []int // 1970-3000 (可选)
}

// CronParser Cron表达式解析器
type CronParser struct {
	allowSeconds bool
	allowYears   bool
}

// NewCronParser 创建Cron解析器
func NewCronParser() *CronParser {
	return &CronParser{
		allowSeconds: true,
		allowYears:   true,
	}
}

// Parse 解析Cron表达式
func (cp *CronParser) Parse(cronExpr string) (*CronExpression, error) {
	fields := strings.Fields(cronExpr)

	// 支持的格式：
	// 5字段: * * * * *        (分 时 日 月 周)
	// 6字段: * * * * * *      (秒 分 时 日 月 周)
	// 7字段: * * * * * * *    (秒 分 时 日 月 周 年)

	var second, minute, hour, dayOfMonth, month, dayOfWeek, year []int
	var err error

	switch len(fields) {
	case 5:
		// 分 时 日 月 周
		minute, err = cp.parseField(fields[0], 0, 59)
		if err != nil {
			return nil, fmt.Errorf("invalid minute field: %w", err)
		}
		hour, err = cp.parseField(fields[1], 0, 23)
		if err != nil {
			return nil, fmt.Errorf("invalid hour field: %w", err)
		}
		dayOfMonth, err = cp.parseField(fields[2], 1, 31)
		if err != nil {
			return nil, fmt.Errorf("invalid day of month field: %w", err)
		}
		month, err = cp.parseField(fields[3], 1, 12)
		if err != nil {
			return nil, fmt.Errorf("invalid month field: %w", err)
		}
		dayOfWeek, err = cp.parseField(fields[4], 0, 6)
		if err != nil {
			return nil, fmt.Errorf("invalid day of week field: %w", err)
		}
		second = []int{0} // 默认为0秒
		year = []int{}    // 不限制年份

	case 6:
		// 秒 分 时 日 月 周
		second, err = cp.parseField(fields[0], 0, 59)
		if err != nil {
			return nil, fmt.Errorf("invalid second field: %w", err)
		}
		minute, err = cp.parseField(fields[1], 0, 59)
		if err != nil {
			return nil, fmt.Errorf("invalid minute field: %w", err)
		}
		hour, err = cp.parseField(fields[2], 0, 23)
		if err != nil {
			return nil, fmt.Errorf("invalid hour field: %w", err)
		}
		dayOfMonth, err = cp.parseField(fields[3], 1, 31)
		if err != nil {
			return nil, fmt.Errorf("invalid day of month field: %w", err)
		}
		month, err = cp.parseField(fields[4], 1, 12)
		if err != nil {
			return nil, fmt.Errorf("invalid month field: %w", err)
		}
		dayOfWeek, err = cp.parseField(fields[5], 0, 6)
		if err != nil {
			return nil, fmt.Errorf("invalid day of week field: %w", err)
		}
		year = []int{} // 不限制年份

	case 7:
		// 秒 分 时 日 月 周 年
		second, err = cp.parseField(fields[0], 0, 59)
		if err != nil {
			return nil, fmt.Errorf("invalid second field: %w", err)
		}
		minute, err = cp.parseField(fields[1], 0, 59)
		if err != nil {
			return nil, fmt.Errorf("invalid minute field: %w", err)
		}
		hour, err = cp.parseField(fields[2], 0, 23)
		if err != nil {
			return nil, fmt.Errorf("invalid hour field: %w", err)
		}
		dayOfMonth, err = cp.parseField(fields[3], 1, 31)
		if err != nil {
			return nil, fmt.Errorf("invalid day of month field: %w", err)
		}
		month, err = cp.parseField(fields[4], 1, 12)
		if err != nil {
			return nil, fmt.Errorf("invalid month field: %w", err)
		}
		dayOfWeek, err = cp.parseField(fields[5], 0, 6)
		if err != nil {
			return nil, fmt.Errorf("invalid day of week field: %w", err)
		}
		year, err = cp.parseField(fields[6], 1970, 3000)
		if err != nil {
			return nil, fmt.Errorf("invalid year field: %w", err)
		}

	default:
		return nil, fmt.Errorf("invalid cron expression: expected 5, 6 or 7 fields, got %d", len(fields))
	}

	return &CronExpression{
		Second:     second,
		Minute:     minute,
		Hour:       hour,
		DayOfMonth: dayOfMonth,
		Month:      month,
		DayOfWeek:  dayOfWeek,
		Year:       year,
	}, nil
}

// parseField 解析单个字段
func (cp *CronParser) parseField(field string, min, max int) ([]int, error) {
	if field == "*" {
		// 通配符，返回所有可能的值
		values := make([]int, max-min+1)
		for i := 0; i < len(values); i++ {
			values[i] = min + i
		}
		return values, nil
	}

	if field == "?" {
		// 问号，用于日期和星期字段的互斥
		return []int{}, nil
	}

	var values []int

	// 处理逗号分隔的值
	parts := strings.Split(field, ",")
	for _, part := range parts {
		if strings.Contains(part, "/") {
			// 处理步长 (例如: 0/5, */2, 1-59/2)
			stepParts := strings.Split(part, "/")
			if len(stepParts) != 2 {
				return nil, fmt.Errorf("invalid step format: %s", part)
			}

			step, err := strconv.Atoi(stepParts[1])
			if err != nil {
				return nil, fmt.Errorf("invalid step value: %s", stepParts[1])
			}

			if step <= 0 {
				return nil, fmt.Errorf("step must be positive: %d", step)
			}

			var start, end int
			if stepParts[0] == "*" {
				start = min
				end = max
			} else if strings.Contains(stepParts[0], "-") {
				// 范围步长 (例如: 1-59/2)
				rangeValues, err := cp.parseRange(stepParts[0], min, max)
				if err != nil {
					return nil, err
				}
				start = rangeValues[0]
				end = rangeValues[len(rangeValues)-1]
			} else {
				// 单个值步长 (例如: 0/5)
				start, err = strconv.Atoi(stepParts[0])
				if err != nil {
					return nil, fmt.Errorf("invalid start value: %s", stepParts[0])
				}
				end = max
			}

			for i := start; i <= end; i += step {
				if i >= min && i <= max {
					values = append(values, i)
				}
			}

		} else if strings.Contains(part, "-") {
			// 处理范围 (例如: 1-5)
			rangeValues, err := cp.parseRange(part, min, max)
			if err != nil {
				return nil, err
			}
			values = append(values, rangeValues...)

		} else {
			// 处理单个值
			value, err := cp.parseSingleValue(part, min, max)
			if err != nil {
				return nil, err
			}
			values = append(values, value)
		}
	}

	// 去重并排序
	uniqueValues := cp.removeDuplicates(values)
	return uniqueValues, nil
}

// parseRange 解析范围
func (cp *CronParser) parseRange(rangeStr string, min, max int) ([]int, error) {
	parts := strings.Split(rangeStr, "-")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid range format: %s", rangeStr)
	}

	start, err := cp.parseSingleValue(parts[0], min, max)
	if err != nil {
		return nil, fmt.Errorf("invalid range start: %w", err)
	}

	end, err := cp.parseSingleValue(parts[1], min, max)
	if err != nil {
		return nil, fmt.Errorf("invalid range end: %w", err)
	}

	if start > end {
		return nil, fmt.Errorf("range start (%d) cannot be greater than end (%d)", start, end)
	}

	values := make([]int, end-start+1)
	for i := 0; i < len(values); i++ {
		values[i] = start + i
	}

	return values, nil
}

// parseSingleValue 解析单个值
func (cp *CronParser) parseSingleValue(valueStr string, min, max int) (int, error) {
	// 处理特殊别名
	switch strings.ToUpper(valueStr) {
	case "SUN", "SUNDAY":
		return 0, nil
	case "MON", "MONDAY":
		return 1, nil
	case "TUE", "TUESDAY":
		return 2, nil
	case "WED", "WEDNESDAY":
		return 3, nil
	case "THU", "THURSDAY":
		return 4, nil
	case "FRI", "FRIDAY":
		return 5, nil
	case "SAT", "SATURDAY":
		return 6, nil
	case "JAN", "JANUARY":
		return 1, nil
	case "FEB", "FEBRUARY":
		return 2, nil
	case "MAR", "MARCH":
		return 3, nil
	case "APR", "APRIL":
		return 4, nil
	case "MAY":
		return 5, nil
	case "JUN", "JUNE":
		return 6, nil
	case "JUL", "JULY":
		return 7, nil
	case "AUG", "AUGUST":
		return 8, nil
	case "SEP", "SEPTEMBER":
		return 9, nil
	case "OCT", "OCTOBER":
		return 10, nil
	case "NOV", "NOVEMBER":
		return 11, nil
	case "DEC", "DECEMBER":
		return 12, nil
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return 0, fmt.Errorf("invalid value: %s", valueStr)
	}

	if value < min || value > max {
		return 0, fmt.Errorf("value %d out of range [%d, %d]", value, min, max)
	}

	return value, nil
}

// removeDuplicates 去重
func (cp *CronParser) removeDuplicates(values []int) []int {
	seen := make(map[int]bool)
	result := []int{}

	for _, value := range values {
		if !seen[value] {
			seen[value] = true
			result = append(result, value)
		}
	}

	// 简单排序
	for i := 0; i < len(result)-1; i++ {
		for j := i + 1; j < len(result); j++ {
			if result[i] > result[j] {
				result[i], result[j] = result[j], result[i]
			}
		}
	}

	return result
}

// NextTime 计算下次执行时间
func (ce *CronExpression) NextTime(from time.Time) time.Time {
	// 从下一秒开始计算
	from = from.Add(time.Second).Truncate(time.Second)

	// 最多向前搜索4年
	end := from.AddDate(4, 0, 0)

	for current := from; current.Before(end); current = current.Add(time.Second) {
		if ce.matches(current) {
			return current
		}
	}

	// 如果找不到，返回零值
	return time.Time{}
}

// matches 检查时间是否匹配Cron表达式
func (ce *CronExpression) matches(t time.Time) bool {
	// 检查年份
	if len(ce.Year) > 0 && !ce.contains(ce.Year, t.Year()) {
		return false
	}

	// 检查月份
	if !ce.contains(ce.Month, int(t.Month())) {
		return false
	}

	// 检查小时
	if !ce.contains(ce.Hour, t.Hour()) {
		return false
	}

	// 检查分钟
	if !ce.contains(ce.Minute, t.Minute()) {
		return false
	}

	// 检查秒
	if !ce.contains(ce.Second, t.Second()) {
		return false
	}

	// 检查日期和星期（OR关系）
	dayOfMonthMatch := len(ce.DayOfMonth) == 0 || ce.contains(ce.DayOfMonth, t.Day())
	dayOfWeekMatch := len(ce.DayOfWeek) == 0 || ce.contains(ce.DayOfWeek, int(t.Weekday()))

	// 如果两个都有值，则是OR关系；如果只有一个有值，则必须匹配
	if len(ce.DayOfMonth) > 0 && len(ce.DayOfWeek) > 0 {
		return dayOfMonthMatch || dayOfWeekMatch
	} else {
		return dayOfMonthMatch && dayOfWeekMatch
	}
}

// contains 检查数组是否包含值
func (ce *CronExpression) contains(arr []int, value int) bool {
	for _, v := range arr {
		if v == value {
			return true
		}
	}
	return false
}

// String 返回Cron表达式的字符串表示
func (ce *CronExpression) String() string {
	return fmt.Sprintf("Cron{sec:%v, min:%v, hour:%v, dom:%v, month:%v, dow:%v, year:%v}",
		ce.Second, ce.Minute, ce.Hour, ce.DayOfMonth, ce.Month, ce.DayOfWeek, ce.Year)
}

// ============= 预定义的Cron表达式 =============

var (
	// 常用的Cron表达式
	CronEveryMinute = "0 * * * * *"
	CronEveryHour   = "0 0 * * * *"
	CronEveryDay    = "0 0 0 * * *"
	CronEveryWeek   = "0 0 0 * * 0"
	CronEveryMonth  = "0 0 0 1 * *"
	CronEveryYear   = "0 0 0 1 1 *"

	// 工作日相关
	CronWeekdays = "0 0 9 * * 1-5"  // 工作日上午9点
	CronWeekends = "0 0 10 * * 0,6" // 周末上午10点

	// 特定时间
	CronMidnight = "0 0 0 * * *"  // 每天午夜
	CronNoon     = "0 0 12 * * *" // 每天中午

	// 高频率
	CronEvery30Sec = "*/30 * * * * *" // 每30秒
	CronEvery5Min  = "0 */5 * * * *"  // 每5分钟
	CronEvery15Min = "0 */15 * * * *" // 每15分钟
	CronEvery30Min = "0 */30 * * * *" // 每30分钟
)

// ParseCronExpression 解析Cron表达式的便捷函数
func ParseCronExpression(cronExpr string) (*CronExpression, error) {
	parser := NewCronParser()
	return parser.Parse(cronExpr)
}

// ValidateCronExpression 验证Cron表达式
func ValidateCronExpression(cronExpr string) error {
	_, err := ParseCronExpression(cronExpr)
	return err
}

// GetNextCronTime 获取下次Cron执行时间
func GetNextCronTime(cronExpr string, from time.Time) (time.Time, error) {
	cron, err := ParseCronExpression(cronExpr)
	if err != nil {
		return time.Time{}, err
	}

	nextTime := cron.NextTime(from)
	if nextTime.IsZero() {
		return time.Time{}, fmt.Errorf("no next execution time found")
	}

	return nextTime, nil
}

// ============= Cron调度器增强 =============

// 扩展原有的parseSchedule方法以支持Cron表达式
func (s *Scheduler) parseScheduleWithCron(schedule string) (time.Time, error) {
	// 首先尝试解析为Cron表达式
	if err := ValidateCronExpression(schedule); err == nil {
		return GetNextCronTime(schedule, time.Now())
	}

	// 如果不是Cron表达式，使用原有的解析逻辑
	return s.parseSchedule(schedule)
}

// CronTask Cron任务的便捷包装
type CronTask struct {
	*Task
	cronExpr *CronExpression
}

// NewCronTask 创建Cron任务
func NewCronTask(id, name, description, cronExpr string, job Job) (*CronTask, error) {
	// 验证Cron表达式
	expr, err := ParseCronExpression(cronExpr)
	if err != nil {
		return nil, fmt.Errorf("invalid cron expression: %w", err)
	}

	task := NewTask(id, name, description, cronExpr, job)

	return &CronTask{
		Task:     task,
		cronExpr: expr,
	}, nil
}

// GetNextRunTime 获取下次运行时间
func (ct *CronTask) GetNextRunTime() time.Time {
	return ct.cronExpr.NextTime(time.Now())
}

// UpdateNextRunTime 更新下次运行时间
func (ct *CronTask) UpdateNextRunTime() {
	nextTime := ct.GetNextRunTime()
	if !nextTime.IsZero() {
		ct.SetNextRunTime(nextTime)
	}
}

// ============= 便捷函数 =============

// ScheduleCronJob 调度Cron任务的便捷函数
func ScheduleCronJob(id, name, description, cronExpr string, jobFunc func(ctx context.Context) error) error {
	job := NewJobFunc(name, description, jobFunc)

	cronTask, err := NewCronTask(id, name, description, cronExpr, job)
	if err != nil {
		return err
	}

	return GetGlobalScheduler().AddTask(cronTask.Task)
}

// ScheduleAt 在指定时间调度任务
func ScheduleAt(id, name, description string, scheduleTime time.Time, jobFunc func(ctx context.Context) error) error {
	timeStr := scheduleTime.Format("2006-01-02 15:04:05")
	job := NewJobFunc(name, description, jobFunc)
	task := NewTask(id, name, description, timeStr, job)

	return GetGlobalScheduler().AddTask(task)
}

// ScheduleAfter 在指定延迟后调度任务
func ScheduleAfter(id, name, description string, delay time.Duration, jobFunc func(ctx context.Context) error) error {
	job := NewJobFunc(name, description, jobFunc)
	task := NewTask(id, name, description, delay.String(), job)

	return GetGlobalScheduler().AddTask(task)
}

// ScheduleEvery 按间隔调度任务
func ScheduleEvery(id, name, description string, interval time.Duration, jobFunc func(ctx context.Context) error) error {
	job := NewJobFunc(name, description, jobFunc)
	task := NewTask(id, name, description, "@every_"+interval.String(), job)

	return GetGlobalScheduler().AddTask(task)
}

// ExampleCronUsage 使用示例
func ExampleCronUsage() {
	config.Info("=== Cron Expression Examples ===")

	examples := []string{
		"0 0 * * * *",    // 每小时
		"0 */15 * * * *", // 每15分钟
		"0 0 9 * * 1-5",  // 工作日上午9点
		"0 0 0 1 * *",    // 每月1号午夜
		"0 30 8 * * MON", // 每周一上午8:30
		"*/30 * * * * *", // 每30秒
	}

	for _, expr := range examples {
		if cron, err := ParseCronExpression(expr); err == nil {
			nextTime := cron.NextTime(time.Now())
			config.Infof("Expression: %s -> Next: %s", expr, nextTime.Format("2006-01-02 15:04:05"))
		} else {
			config.Errorf("Invalid expression: %s -> %v", expr, err)
		}
	}
}
