package yyhertz

import (
	"context"
	"fmt"

	"github.com/zsy619/yyhertz/framework/config"
	"github.com/zsy619/yyhertz/framework/util"
)

// GenericController 基于泛型的通用控制器
type GenericController struct {
	BaseController
}

// NewGenericController 创建泛型控制器实例
func NewGenericController() *GenericController {
	return &GenericController{
		BaseController: *NewBaseController(),
	}
}

// =============== 泛型处理器集成方法 ===============

// HandleWithGenericHandler 使用泛型处理器处理请求 (类型安全版本)
func HandleWithGenericHandler[T any, R any](c *GenericController, handler util.HTTPGenericHandler[T, R]) {
	if c.Ctx == nil {
		config.Error("GenericController: Context is nil")
		return
	}

	// 记录控制器动作
	c.LogControllerAction("HandleWithGeneric", map[string]any{
		"handler_type": fmt.Sprintf("%T", handler),
	})

	// 调用泛型处理器
	response, err := handler.HandleHTTP(c.Ctx)
	if err != nil {
		// 错误已经在处理器中处理
		return
	}

	// 设置响应
	statusCode := 200
	if !response.Success {
		statusCode = response.Code
	}

	c.JSONWithStatus(statusCode, response)
}

// ProcessWithChain 使用处理链处理数据
func ProcessWithChain[T any](c *GenericController, chain *util.GenericChain[T], input T) (T, error) {
	// 记录链处理开始
	c.LogBusinessLogic("chain_processing_start", map[string]any{
		"chain_name": getChainName(chain),
		"input_type": fmt.Sprintf("%T", input),
	})

	if chain == nil {
		err := fmt.Errorf("chain is nil")
		c.LogError("ProcessWithChain failed: %v", err)
		return input, err
	}

	// 执行处理链
	result, err := chain.Execute(context.Background(), input)
	if err != nil {
		c.LogBusinessLogic("chain_processing_failed", map[string]any{
			"chain_name": getChainName(chain),
			"error":      err.Error(),
		})
		return input, err
	}

	// 记录链处理成功
	c.LogBusinessLogic("chain_processing_completed", map[string]any{
		"chain_name": getChainName(chain),
		"success":    true,
	})

	return result, nil
}

// ProcessWithPipeline 使用管道处理数据
func ProcessWithPipeline[T any, R any](c *GenericController, pipeline *util.GenericPipeline[T, R], input T) (R, error) {
	// 记录管道处理开始
	c.LogBusinessLogic("pipeline_processing_start", map[string]any{
		"pipeline_name": getPipelineName(pipeline),
		"input_type":    fmt.Sprintf("%T", input),
	})

	var zero R
	if pipeline == nil {
		err := fmt.Errorf("pipeline is nil")
		c.LogError("ProcessWithPipeline failed: %v", err)
		return zero, err
	}

	// 执行管道
	result, err := pipeline.Execute(context.Background(), input)
	if err != nil {
		c.LogBusinessLogic("pipeline_processing_failed", map[string]any{
			"pipeline_name": getPipelineName(pipeline),
			"error":         err.Error(),
		})
		return zero, err
	}

	// 记录管道处理成功
	c.LogBusinessLogic("pipeline_processing_completed", map[string]any{
		"pipeline_name": getPipelineName(pipeline),
		"output_type":   fmt.Sprintf("%T", result),
		"success":       true,
	})

	return result, nil
}

// =============== 便捷方法 ===============

// CreateGenericHandler 创建泛型处理器的工厂方法
func CreateGenericHandler[T any, R any](name string) *util.BaseHTTPGenericHandler[T, R] {
	return util.NewBaseHTTPGenericHandler[T, R](name)
}

// =============== 示例控制器方法 ===============

// ExampleUserCreate 用户创建示例（使用泛型处理器）
func (c *GenericController) ExampleUserCreate() {
	// 使用预定义的泛型处理器
	handler := util.CreateUserGenericHandler()
	HandleWithGenericHandler(c, handler)
}

// ExampleUserQuery 用户查询示例（使用泛型处理器）
func (c *GenericController) ExampleUserQuery() {
	// 使用预定义的泛型处理器
	handler := util.QueryUsersGenericHandler()
	HandleWithGenericHandler(c, handler)
}

// ExampleInlineHandler 内联处理器示例
func (c *GenericController) ExampleInlineHandler() {
	// 定义请求和响应结构
	type SimpleRequest struct {
		Message string `json:"message" binding:"required"`
	}

	type SimpleResponse struct {
		Echo      string `json:"echo"`
		Timestamp int64  `json:"timestamp"`
	}

	// 创建内联处理器
	handler := util.NewBaseHTTPGenericHandler[SimpleRequest, SimpleResponse]("EchoMessage")
	handler.BaseGenericHandler = handler.BaseGenericHandler.
		WithValidator(func(req SimpleRequest) error {
			if req.Message == "" {
				return fmt.Errorf("message cannot be empty")
			}
			return nil
		}).
		WithProcessor(func(ctx context.Context, req SimpleRequest) (SimpleResponse, error) {
			c.LogInfo("Processing echo request: %s", req.Message)

			return SimpleResponse{
				Echo:      fmt.Sprintf("Echo: %s", req.Message),
				Timestamp: c.CreateTime(),
			}, nil
		})

	HandleWithGenericHandler(c, handler)
}

// ExampleChainProcessing 处理链示例
func (c *GenericController) ExampleChainProcessing() {
	// 定义输入数据
	type ChainInput struct {
		Value string `json:"value"`
		Count int    `json:"count"`
	}

	// 从请求中解析数据
	var input ChainInput
	if err := c.Ctx.BindJSON(&input); err != nil {
		c.JSONBadRequest("Invalid request format")
		return
	}

	// 创建处理链
	chain := util.NewGenericChain[ChainInput]("ExampleChain").
		Add(func(ctx context.Context, data ChainInput) (ChainInput, error) {
			// 步骤1: 验证
			if data.Value == "" {
				return data, fmt.Errorf("value cannot be empty")
			}
			return data, nil
		}).
		Add(func(ctx context.Context, data ChainInput) (ChainInput, error) {
			// 步骤2: 转换
			data.Value = fmt.Sprintf("processed_%s", data.Value)
			data.Count += 10
			return data, nil
		})

	// 执行处理链
	result, err := ProcessWithChain(c, chain, input)
	if err != nil {
		c.JSONBadRequest(err.Error())
		return
	}

	c.JSONOK("Chain processing completed", result)
}

// ExamplePipelineProcessing 管道处理示例
func (c *GenericController) ExamplePipelineProcessing() {
	// 定义输入和输出结构
	type PipelineInput struct {
		Numbers []int `json:"numbers"`
	}

	type PipelineOutput struct {
		Sum     int     `json:"sum"`
		Average float64 `json:"average"`
		Count   int     `json:"count"`
	}

	// 从请求中解析数据
	var input PipelineInput
	if err := c.Ctx.BindJSON(&input); err != nil {
		c.JSONBadRequest("Invalid request format")
		return
	}

	// 创建处理管道
	pipeline := util.NewGenericPipeline[PipelineInput, PipelineOutput]("StatisticsPipeline").
		AddStage(func(ctx context.Context, data any) (any, error) {
			// 阶段1: 验证和预处理
			input := data.(PipelineInput)
			if len(input.Numbers) == 0 {
				return nil, fmt.Errorf("numbers array cannot be empty")
			}
			return input, nil
		}).
		AddStage(func(ctx context.Context, data any) (any, error) {
			// 阶段2: 计算统计信息
			input := data.(PipelineInput)
			sum := 0
			for _, num := range input.Numbers {
				sum += num
			}

			intermediate := map[string]any{
				"sum":   sum,
				"count": len(input.Numbers),
			}
			return intermediate, nil
		}).
		AddStage(func(ctx context.Context, data any) (any, error) {
			// 阶段3: 构建最终结果
			intermediate := data.(map[string]any)
			sum := intermediate["sum"].(int)
			count := intermediate["count"].(int)

			output := PipelineOutput{
				Sum:     sum,
				Average: float64(sum) / float64(count),
				Count:   count,
			}
			return output, nil
		})

	// 执行管道
	result, err := ProcessWithPipeline(c, pipeline, input)
	if err != nil {
		c.JSONBadRequest(err.Error())
		return
	}

	c.JSONOK("Pipeline processing completed", result)
}

// =============== 辅助函数 ===============

func getChainName[T any](chain *util.GenericChain[T]) string {
	if chain == nil {
		return "unknown"
	}
	// 由于GenericChain的name字段可能不是导出的，我们使用反射或提供getter方法
	// 这里简化处理
	return "chain"
}

func getPipelineName[T any, R any](pipeline *util.GenericPipeline[T, R]) string {
	if pipeline == nil {
		return "unknown"
	}
	// 同样，这里简化处理
	return "pipeline"
}
