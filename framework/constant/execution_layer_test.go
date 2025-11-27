package constant

import (
	"testing"
)

func TestExecutionLayerConstants(t *testing.T) {
	// 测试统一常量的基本功能
	tests := []struct {
		name     string
		layer    ExecutionLayer
		expected string
	}{
		{"BeforeStatic", LayerBeforeStatic, "BeforeStatic"},
		{"Global", LayerGlobal, "Global"},
		{"Group", LayerGroup, "Group"},
		{"Route", LayerRoute, "Route"},
		{"Controller", LayerController, "Controller"},
		{"FinishRouter", LayerFinishRouter, "FinishRouter"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if name := GetExecutionLayerName(tt.layer); name != tt.expected {
				t.Errorf("GetExecutionLayerName(%v) = %v, want %v", tt.layer, name, tt.expected)
			}
		})
	}
}

func TestFilterPositionCompatibility(t *testing.T) {
	// 测试过滤器位置常量的向后兼容性
	tests := []struct {
		name     string
		position int
		expected string
	}{
		{"BeforeStatic", BeforeStatic, "BeforeStatic"},
		{"BeforeRouter", BeforeRouter, "BeforeRouter"},
		{"BeforeExec", BeforeExec, "BeforeExec"},
		{"AfterExec", AfterExec, "AfterExec"},
		{"FinishRouter", FinishRouter, "FinishRouter"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if name := GetFilterPositionName(tt.position); name != tt.expected {
				t.Errorf("GetFilterPositionName(%v) = %v, want %v", tt.position, name, tt.expected)
			}
		})
	}
}

func TestConversionFunctions(t *testing.T) {
	// 测试转换函数
	conversionTests := []struct {
		name     string
		position int
		layer    ExecutionLayer
	}{
		{"BeforeStatic", BeforeStatic, LayerBeforeStatic},
		{"BeforeRouter", BeforeRouter, LayerGlobal},
		{"BeforeExec", BeforeExec, LayerRoute},
		{"AfterExec", AfterExec, LayerController},
		{"FinishRouter", FinishRouter, LayerFinishRouter},
	}

	for _, tt := range conversionTests {
		t.Run(tt.name, func(t *testing.T) {
			// 测试位置到层级的转换
			if layer := FilterPositionToLayer(tt.position); layer != tt.layer {
				t.Errorf("FilterPositionToLayer(%v) = %v, want %v", tt.position, layer, tt.layer)
			}

			// 测试层级到位置的转换
			if position := LayerToFilterPosition(tt.layer); position != tt.position {
				t.Errorf("LayerToFilterPosition(%v) = %v, want %v", tt.layer, position, tt.position)
			}
		})
	}
}

func TestValidationFunctions(t *testing.T) {
	// 测试验证函数
	validLayers := []ExecutionLayer{
		LayerBeforeStatic, LayerGlobal, LayerGroup, LayerRoute, LayerController, LayerFinishRouter,
	}
	
	for _, layer := range validLayers {
		if !IsValidExecutionLayer(layer) {
			t.Errorf("IsValidExecutionLayer(%v) should be true", layer)
		}
	}

	validPositions := []int{BeforeStatic, BeforeRouter, BeforeExec, AfterExec, FinishRouter}
	for _, position := range validPositions {
		if !IsValidFilterPosition(position) {
			t.Errorf("IsValidFilterPosition(%v) should be true", position)
		}
	}

	// 测试无效值
	if IsValidExecutionLayer(ExecutionLayer(999)) {
		t.Error("IsValidExecutionLayer(999) should be false")
	}

	if IsValidFilterPosition(999) {
		t.Error("IsValidFilterPosition(999) should be false")
	}
}

func TestConstantMapping(t *testing.T) {
	// 验证常量映射的一致性
	expectedMappings := map[int]ExecutionLayer{
		BeforeStatic: LayerBeforeStatic,
		BeforeRouter: LayerGlobal,
		BeforeExec:   LayerRoute,
		AfterExec:    LayerController,
		FinishRouter: LayerFinishRouter,
	}

	for position, expectedLayer := range expectedMappings {
		if actualLayer := FilterPositionToLayer(position); actualLayer != expectedLayer {
			t.Errorf("Position %d should map to layer %v, got %v", position, expectedLayer, actualLayer)
		}
		
		if actualPosition := LayerToFilterPosition(expectedLayer); actualPosition != position {
			t.Errorf("Layer %v should map to position %d, got %d", expectedLayer, position, actualPosition)
		}
	}
}