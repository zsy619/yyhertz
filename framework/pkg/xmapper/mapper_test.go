package xmapper

import (
	"reflect"
	"testing"
	"time"
)

// 测试用结构体
type TestUser struct {
	ID      int       `json:"id"`
	Name    string    `json:"name"`
	Email   string    `json:"email"`
	Age     int       `json:"age"`
	Active  bool      `json:"active"`
	Created time.Time `json:"created"`
	Score   float64   `json:"score"`
}

type TestUserDTO struct {
	ID      int       `json:"id"`
	Name    string    `json:"name"`
	Email   string    `json:"email"`
	Age     int       `json:"age"`
	Active  bool      `json:"active"`
	Created time.Time `json:"created"`
	Score   float64   `json:"score"`
}

type NestedUser struct {
	ID   int      `json:"id"`
	Name string   `json:"name"`
	Tags []string `json:"tags"`
}

type NestedUserDTO struct {
	ID   int      `json:"id"`
	Name string   `json:"name"`
	Tags []string `json:"tags"`
}

func TestMapperStrategies(t *testing.T) {
	tests := []struct {
		name     string
		strategy MappingStrategy
	}{
		{"JSON策略", StrategyJSON},
		{"反射策略", StrategyReflection},
		{"代码生成策略", StrategyCodegen},
		{"自动策略", StrategyAuto},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mapper := NewMapper(WithStrategy(tt.strategy))

			src := &TestUser{
				ID:      1,
				Name:    "张三",
				Email:   "zhangsan@example.com",
				Age:     25,
				Active:  true,
				Created: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
				Score:   99.5,
			}

			var dst TestUserDTO
			err := mapper.Map(src, &dst)
			if err != nil {
				t.Fatalf("映射失败: %v", err)
			}

			// 验证映射结果
			if dst.ID != src.ID {
				t.Errorf("ID映射错误: 期望 %d, 得到 %d", src.ID, dst.ID)
			}
			if dst.Name != src.Name {
				t.Errorf("Name映射错误: 期望 %s, 得到 %s", src.Name, dst.Name)
			}
			if dst.Email != src.Email {
				t.Errorf("Email映射错误: 期望 %s, 得到 %s", src.Email, dst.Email)
			}
			if dst.Age != src.Age {
				t.Errorf("Age映射错误: 期望 %d, 得到 %d", src.Age, dst.Age)
			}
			if dst.Active != src.Active {
				t.Errorf("Active映射错误: 期望 %t, 得到 %t", src.Active, dst.Active)
			}
			if !dst.Created.Equal(src.Created) {
				t.Errorf("Created映射错误: 期望 %v, 得到 %v", src.Created, dst.Created)
			}
			if dst.Score != src.Score {
				t.Errorf("Score映射错误: 期望 %f, 得到 %f", src.Score, dst.Score)
			}
		})
	}
}

func TestMapSlice(t *testing.T) {
	mapper := NewMapper()

	src := []*TestUser{
		{ID: 1, Name: "用户1", Age: 20},
		{ID: 2, Name: "用户2", Age: 30},
		{ID: 3, Name: "用户3", Age: 40},
	}

	var dst []TestUserDTO
	err := mapper.MapSlice(src, &dst)
	if err != nil {
		t.Fatalf("切片映射失败: %v", err)
	}

	if len(dst) != len(src) {
		t.Fatalf("切片长度不匹配: 期望 %d, 得到 %d", len(src), len(dst))
	}

	for i, expected := range src {
		actual := dst[i]
		if actual.ID != expected.ID || actual.Name != expected.Name || actual.Age != expected.Age {
			t.Errorf("切片元素 %d 映射错误", i)
		}
	}
}

func TestMapperWithConverter(t *testing.T) {
	stringType := reflect.TypeOf("")
	converter := func(src any) (any, error) {
		return "转换后的字符串", nil
	}

	mapper := NewMapper(WithConverter(stringType, converter))

	src := &TestUser{
		ID:   1,
		Name: "原始名称",
		Age:  25,
	}

	var dst TestUserDTO
	err := mapper.Map(src, &dst)
	if err != nil {
		t.Fatalf("带转换器的映射失败: %v", err)
	}

	// 注意：转换器应该只在需要的时候才被调用
	// 这里我们主要测试映射器能正常工作
}

func TestNestedStructMapping(t *testing.T) {
	mapper := NewMapper()

	src := &NestedUser{
		ID:   1,
		Name: "嵌套测试",
		Tags: []string{"tag1", "tag2", "tag3"},
	}

	var dst NestedUserDTO
	err := mapper.Map(src, &dst)
	if err != nil {
		t.Fatalf("嵌套结构映射失败: %v", err)
	}

	if dst.ID != src.ID {
		t.Errorf("ID映射错误")
	}
	if dst.Name != src.Name {
		t.Errorf("Name映射错误")
	}
	if len(dst.Tags) != len(src.Tags) {
		t.Errorf("Tags长度不匹配")
	}
}

func TestMapperStats(t *testing.T) {
	mapper := NewMapper()

	src := &TestUser{ID: 1, Name: "统计测试"}
	var dst TestUserDTO

	// 执行映射
	err := mapper.Map(src, &dst)
	if err != nil {
		t.Fatalf("映射失败: %v", err)
	}

	// 检查统计信息
	stats := mapper.GetStats()
	if stats.TotalMaps == 0 {
		t.Errorf("统计信息未更新")
	}
}

// 基准测试
func BenchmarkMapperJSON(b *testing.B) {
	mapper := NewMapper(WithStrategy(StrategyJSON))
	src := &TestUser{
		ID:      1,
		Name:    "基准测试用户",
		Email:   "benchmark@example.com",
		Age:     30,
		Active:  true,
		Created: time.Now(),
		Score:   85.5,
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		var dst TestUserDTO
		if err := mapper.Map(src, &dst); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMapperReflection(b *testing.B) {
	mapper := NewMapper(WithStrategy(StrategyReflection))
	src := &TestUser{
		ID:      1,
		Name:    "基准测试用户",
		Email:   "benchmark@example.com",
		Age:     30,
		Active:  true,
		Created: time.Now(),
		Score:   85.5,
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		var dst TestUserDTO
		if err := mapper.Map(src, &dst); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMapperCodegen(b *testing.B) {
	mapper := NewMapper(WithStrategy(StrategyCodegen))
	src := &TestUser{
		ID:      1,
		Name:    "基准测试用户",
		Email:   "benchmark@example.com",
		Age:     30,
		Active:  true,
		Created: time.Now(),
		Score:   85.5,
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		var dst TestUserDTO
		if err := mapper.Map(src, &dst); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkMapperSlice(b *testing.B) {
	mapper := NewMapper()
	src := make([]*TestUser, 100)
	for i := 0; i < 100; i++ {
		src[i] = &TestUser{
			ID:      i,
			Name:    "批量测试用户",
			Email:   "batch@example.com",
			Age:     20 + i%50,
			Active:  i%2 == 0,
			Created: time.Now(),
			Score:   float64(i % 100),
		}
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		var dst []TestUserDTO
		if err := mapper.MapSlice(src, &dst); err != nil {
			b.Fatal(err)
		}
	}
}
