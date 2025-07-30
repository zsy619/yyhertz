package handler

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestGenericHandler 测试泛型处理器基础功能
func TestGenericHandler(t *testing.T) {
	// 定义测试用的输入输出类型
	type TestInput struct {
		Value string `json:"value"`
		Count int    `json:"count"`
	}

	type TestOutput struct {
		Result    string `json:"result"`
		Processed bool   `json:"processed"`
	}

	t.Run("成功处理", func(t *testing.T) {
		handler := NewBaseGenericHandler[TestInput, TestOutput]("TestHandler").
			WithValidator(func(input TestInput) error {
				if input.Value == "" {
					return errors.New("value cannot be empty")
				}
				return nil
			}).
			WithProcessor(func(ctx context.Context, input TestInput) (TestOutput, error) {
				return TestOutput{
					Result:    "processed_" + input.Value,
					Processed: true,
				}, nil
			})

		input := TestInput{Value: "test", Count: 5}
		result, err := handler.Handle(context.Background(), input)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.True(t, result.Success)
		assert.Equal(t, "processed_test", result.Data.Result)
		assert.True(t, result.Data.Processed)
	})

	t.Run("验证失败", func(t *testing.T) {
		handler := NewBaseGenericHandler[TestInput, TestOutput]("TestHandler").
			WithValidator(func(input TestInput) error {
				if input.Value == "" {
					return errors.New("value cannot be empty")
				}
				return nil
			}).
			WithProcessor(func(ctx context.Context, input TestInput) (TestOutput, error) {
				return TestOutput{}, nil
			})

		input := TestInput{Value: "", Count: 5} // 空值应该验证失败
		result, err := handler.Handle(context.Background(), input)

		assert.Error(t, err)
		assert.NotNil(t, result)
		assert.False(t, result.Success)
		assert.Equal(t, 400, result.Code)
	})

	t.Run("处理器未配置", func(t *testing.T) {
		handler := NewBaseGenericHandler[TestInput, TestOutput]("TestHandler")
		// 不设置processor

		input := TestInput{Value: "test", Count: 5}
		result, err := handler.Handle(context.Background(), input)

		assert.Error(t, err)
		assert.NotNil(t, result)
		assert.False(t, result.Success)
		assert.Contains(t, err.Error(), "processor not configured")
	})

	t.Run("预处理和后处理", func(t *testing.T) {
		handler := NewBaseGenericHandler[TestInput, TestOutput]("TestHandler").
			WithPreProcessor(func(ctx context.Context, input TestInput) (TestInput, error) {
				input.Value = "pre_" + input.Value
				input.Count *= 2
				return input, nil
			}).
			WithProcessor(func(ctx context.Context, input TestInput) (TestOutput, error) {
				return TestOutput{
					Result:    input.Value,
					Processed: input.Count > 0,
				}, nil
			}).
			WithPostProcessor(func(ctx context.Context, output TestOutput) (TestOutput, error) {
				output.Result = output.Result + "_post"
				return output, nil
			})

		input := TestInput{Value: "test", Count: 5}
		result, err := handler.Handle(context.Background(), input)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.True(t, result.Success)
		assert.Equal(t, "pre_test_post", result.Data.Result)
		assert.True(t, result.Data.Processed)
	})
}

// TestGenericChain 测试泛型处理链
func TestGenericChain(t *testing.T) {
	type ChainData struct {
		Value string
		Count int
	}

	t.Run("成功执行链", func(t *testing.T) {
		chain := NewGenericChain[ChainData]("TestChain").
			Add(func(ctx context.Context, data ChainData) (ChainData, error) {
				data.Value = "step1_" + data.Value
				data.Count += 1
				return data, nil
			}).
			Add(func(ctx context.Context, data ChainData) (ChainData, error) {
				data.Value = "step2_" + data.Value
				data.Count += 2
				return data, nil
			}).
			Add(func(ctx context.Context, data ChainData) (ChainData, error) {
				data.Value = "step3_" + data.Value
				data.Count += 3
				return data, nil
			})

		input := ChainData{Value: "test", Count: 0}
		result, err := chain.Execute(context.Background(), input)

		assert.NoError(t, err)
		assert.Equal(t, "step3_step2_step1_test", result.Value)
		assert.Equal(t, 6, result.Count) // 1+2+3
	})

	t.Run("链中步骤失败", func(t *testing.T) {
		chain := NewGenericChain[ChainData]("TestChain").
			Add(func(ctx context.Context, data ChainData) (ChainData, error) {
				data.Value = "step1_" + data.Value
				return data, nil
			}).
			Add(func(ctx context.Context, data ChainData) (ChainData, error) {
				return data, errors.New("step2 failed")
			}).
			Add(func(ctx context.Context, data ChainData) (ChainData, error) {
				data.Value = "step3_" + data.Value
				return data, nil
			})

		input := ChainData{Value: "test", Count: 0}
		result, err := chain.Execute(context.Background(), input)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "step2 failed")
		// 结果应该是step1处理后的结果
		assert.Equal(t, "step1_test", result.Value)
	})

	t.Run("空链", func(t *testing.T) {
		chain := NewGenericChain[ChainData]("EmptyChain")

		input := ChainData{Value: "test", Count: 0}
		result, err := chain.Execute(context.Background(), input)

		assert.NoError(t, err)
		assert.Equal(t, input, result) // 应该返回原始输入
	})
}

// TestGenericPipeline 测试泛型管道
func TestGenericPipeline(t *testing.T) {
	type PipelineInput struct {
		Numbers []int
	}

	type PipelineOutput struct {
		Sum     int
		Average float64
		Count   int
	}

	t.Run("成功执行管道", func(t *testing.T) {
		pipeline := NewGenericPipeline[PipelineInput, PipelineOutput]("TestPipeline").
			AddStage(func(ctx context.Context, data any) (any, error) {
				// 验证阶段
				input := data.(PipelineInput)
				if len(input.Numbers) == 0 {
					return nil, errors.New("numbers cannot be empty")
				}
				return input, nil
			}).
			AddStage(func(ctx context.Context, data any) (any, error) {
				// 计算阶段
				input := data.(PipelineInput)
				sum := 0
				for _, num := range input.Numbers {
					sum += num
				}
				return map[string]any{
					"sum":   sum,
					"count": len(input.Numbers),
				}, nil
			}).
			AddStage(func(ctx context.Context, data any) (any, error) {
				// 构建结果阶段
				intermediate := data.(map[string]any)
				sum := intermediate["sum"].(int)
				count := intermediate["count"].(int)

				return PipelineOutput{
					Sum:     sum,
					Average: float64(sum) / float64(count),
					Count:   count,
				}, nil
			})

		input := PipelineInput{Numbers: []int{1, 2, 3, 4, 5}}
		result, err := pipeline.Execute(context.Background(), input)

		assert.NoError(t, err)
		assert.Equal(t, 15, result.Sum)
		assert.Equal(t, 3.0, result.Average)
		assert.Equal(t, 5, result.Count)
	})

	t.Run("管道阶段失败", func(t *testing.T) {
		pipeline := NewGenericPipeline[PipelineInput, PipelineOutput]("TestPipeline").
			AddStage(func(ctx context.Context, data any) (any, error) {
				return nil, errors.New("first stage failed")
			})

		input := PipelineInput{Numbers: []int{1, 2, 3}}
		result, err := pipeline.Execute(context.Background(), input)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "first stage failed")
		// 结果应该是零值
		assert.Equal(t, PipelineOutput{}, result)
	})
}

// TestGenericResult 测试泛型结果结构
func TestGenericResult(t *testing.T) {
	t.Run("创建成功结果", func(t *testing.T) {
		data := map[string]string{"key": "value"}
		result := NewSuccessResult(data, "操作成功")

		assert.True(t, result.Success)
		assert.Equal(t, 200, result.Code)
		assert.Equal(t, "操作成功", result.Message)
		assert.Equal(t, data, result.Data)
		assert.NotZero(t, result.Timestamp)
		assert.NotNil(t, result.Meta)
	})

	t.Run("创建错误结果", func(t *testing.T) {
		result := NewErrorResult[string]("操作失败", 400)

		assert.False(t, result.Success)
		assert.Equal(t, 400, result.Code)
		assert.Equal(t, "操作失败", result.Message)
		assert.Equal(t, "", result.Data) // 字符串零值
		assert.NotZero(t, result.Timestamp)
	})

	t.Run("自定义结果", func(t *testing.T) {
		data := 42
		result := NewGenericResult(data, true, "自定义消息", 201)

		assert.True(t, result.Success)
		assert.Equal(t, 201, result.Code)
		assert.Equal(t, "自定义消息", result.Message)
		assert.Equal(t, 42, result.Data)
	})
}

// BenchmarkGenericHandler 基准测试
func BenchmarkGenericHandler(b *testing.B) {
	type TestData struct {
		Value string
		Count int
	}

	handler := NewBaseGenericHandler[TestData, TestData]("BenchHandler").
		WithProcessor(func(ctx context.Context, input TestData) (TestData, error) {
			return TestData{
				Value: "processed_" + input.Value,
				Count: input.Count * 2,
			}, nil
		})

	input := TestData{Value: "test", Count: 10}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = handler.Handle(context.Background(), input)
	}
}

// BenchmarkGenericChain 链处理基准测试
func BenchmarkGenericChain(b *testing.B) {
	type ChainData struct {
		Value string
		Count int
	}

	chain := NewGenericChain[ChainData]("BenchChain").
		Add(func(ctx context.Context, data ChainData) (ChainData, error) {
			data.Value = "step1_" + data.Value
			data.Count++
			return data, nil
		}).
		Add(func(ctx context.Context, data ChainData) (ChainData, error) {
			data.Value = "step2_" + data.Value
			data.Count++
			return data, nil
		}).
		Add(func(ctx context.Context, data ChainData) (ChainData, error) {
			data.Value = "step3_" + data.Value
			data.Count++
			return data, nil
		})

	input := ChainData{Value: "test", Count: 0}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = chain.Execute(context.Background(), input)
	}
}
