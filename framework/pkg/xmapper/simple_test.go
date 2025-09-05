package xmapper

import (
	"testing"
)

// 简单的测试结构体
type SimpleUser struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Age  int    `json:"age"`
}

type SimpleUserDTO struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func TestBasicMapping(t *testing.T) {
	mapper := NewMapper()

	src := &SimpleUser{
		ID:   1,
		Name: "Test User",
		Age:  25,
	}

	var dst SimpleUserDTO

	err := mapper.Map(src, &dst)
	if err != nil {
		t.Fatalf("映射失败: %v", err)
	}

	if dst.ID != src.ID {
		t.Errorf("ID 映射错误: 期望 %d, 得到 %d", src.ID, dst.ID)
	}

	if dst.Name != src.Name {
		t.Errorf("Name 映射错误: 期望 %s, 得到 %s", src.Name, dst.Name)
	}

	if dst.Age != src.Age {
		t.Errorf("Age 映射错误: 期望 %d, 得到 %d", src.Age, dst.Age)
	}
}

func TestJSONStrategy(t *testing.T) {
	mapper := NewMapper(WithStrategy(StrategyJSON))

	src := &SimpleUser{
		ID:   2,
		Name: "JSON Test",
		Age:  30,
	}

	var dst SimpleUserDTO

	err := mapper.Map(src, &dst)
	if err != nil {
		t.Fatalf("JSON策略映射失败: %v", err)
	}

	if dst.ID != src.ID || dst.Name != src.Name || dst.Age != src.Age {
		t.Errorf("JSON策略映射结果不正确")
	}
}

func TestReflectionStrategy(t *testing.T) {
	mapper := NewMapper(WithStrategy(StrategyReflection))

	src := &SimpleUser{
		ID:   3,
		Name: "Reflection Test",
		Age:  35,
	}

	var dst SimpleUserDTO

	err := mapper.Map(src, &dst)
	if err != nil {
		t.Fatalf("反射策略映射失败: %v", err)
	}

	if dst.ID != src.ID || dst.Name != src.Name || dst.Age != src.Age {
		t.Errorf("反射策略映射结果不正确")
	}
}
