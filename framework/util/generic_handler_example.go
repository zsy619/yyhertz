package util

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/zsy619/yyhertz/framework/config"
)

// =============== 示例数据结构 ===============

// UserCreateRequest 用户创建请求
type UserCreateRequest struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Age      int    `json:"age" binding:"min=1,max=150"`
	Password string `json:"password" binding:"required,min=6"`
}

// UserCreateResponse 用户创建响应
type UserCreateResponse struct {
	ID       string    `json:"id"`
	Name     string    `json:"name"`
	Email    string    `json:"email"`
	Age      int       `json:"age"`
	CreateAt time.Time `json:"create_at"`
}

// UserQueryRequest 用户查询请求
type UserQueryRequest struct {
	Page     int    `json:"page" form:"page"`
	PageSize int    `json:"page_size" form:"page_size"`
	Keyword  string `json:"keyword" form:"keyword"`
	Status   string `json:"status" form:"status"`
}

// UserQueryResponse 用户查询响应
type UserQueryResponse struct {
	Users []UserCreateResponse `json:"users"`
	Total int                  `json:"total"`
	Page  int                  `json:"page"`
	Size  int                  `json:"size"`
}

// =============== 示例处理器实现 ===============

// CreateUserGenericHandler 创建用户的泛型处理器示例
func CreateUserGenericHandler() *BaseHTTPGenericHandler[UserCreateRequest, UserCreateResponse] {
	handler := NewBaseHTTPGenericHandler[UserCreateRequest, UserCreateResponse]("CreateUser").
		WithAuth(true).
		WithRateLimit(10, time.Minute)
	
	handler.BaseGenericHandler = handler.BaseGenericHandler.
		WithValidator(func(req UserCreateRequest) error {
			if req.Name == "" {
				return errors.New("name is required")
			}
			if req.Email == "" {
				return errors.New("email is required")
			}
			if req.Age < 1 || req.Age > 150 {
				return errors.New("age must be between 1 and 150")
			}
			if len(req.Password) < 6 {
				return errors.New("password must be at least 6 characters")
			}
			return nil
		}).
		WithPreProcessor(func(ctx context.Context, req UserCreateRequest) (UserCreateRequest, error) {
			// 预处理：密码加密、数据清理等
			config.Debug("Pre-processing user creation request")
			
			// 清理空格
			req.Name = fmt.Sprintf("%s", req.Name)
			req.Email = fmt.Sprintf("%s", req.Email)
			
			// 这里可以添加密码加密逻辑
			// req.Password = hashPassword(req.Password)
			
			return req, nil
		}).
		WithProcessor(func(ctx context.Context, req UserCreateRequest) (UserCreateResponse, error) {
			// 核心处理：创建用户
			config.WithFields(map[string]any{
				"name":  req.Name,
				"email": req.Email,
				"age":   req.Age,
			}).Info("Creating user")
			
			// 模拟数据库操作
			time.Sleep(100 * time.Millisecond)
			
			// 模拟邮箱重复检查
			if req.Email == "duplicate@example.com" {
				return UserCreateResponse{}, errors.New("email already exists")
			}
			
			// 创建用户响应
			response := UserCreateResponse{
				ID:       ShortID(),
				Name:     req.Name,
				Email:    req.Email,
				Age:      req.Age,
				CreateAt: time.Now(),
			}
			
			return response, nil
		}).
		WithPostProcessor(func(ctx context.Context, resp UserCreateResponse) (UserCreateResponse, error) {
			// 后处理：发送欢迎邮件、日志记录等
			config.WithFields(map[string]any{
				"user_id": resp.ID,
				"email":   resp.Email,
			}).Info("User created successfully, sending welcome email")
			
			// 这里可以添加异步任务，如发送邮件
			// emailService.SendWelcomeEmail(resp.Email)
			
			return resp, nil
		})
	
	return handler
}

// QueryUsersGenericHandler 用户查询的泛型处理器示例
func QueryUsersGenericHandler() *BaseHTTPGenericHandler[UserQueryRequest, UserQueryResponse] {
	handler := NewBaseHTTPGenericHandler[UserQueryRequest, UserQueryResponse]("QueryUsers").
		WithAuth(false).
		WithCache(5 * time.Minute)
	
	handler.BaseGenericHandler = handler.BaseGenericHandler.
		WithValidator(func(req UserQueryRequest) error {
			if req.Page < 1 {
				req.Page = 1
			}
			if req.PageSize < 1 || req.PageSize > 100 {
				req.PageSize = 10
			}
			return nil
		}).
		WithProcessor(func(ctx context.Context, req UserQueryRequest) (UserQueryResponse, error) {
			config.WithFields(map[string]any{
				"page":      req.Page,
				"page_size": req.PageSize,
				"keyword":   req.Keyword,
				"status":    req.Status,
			}).Debug("Querying users")
			
			// 模拟数据库查询
			time.Sleep(50 * time.Millisecond)
			
			// 模拟查询结果
			users := []UserCreateResponse{
				{
					ID:       "user1",
					Name:     "John Doe",
					Email:    "john@example.com",
					Age:      25,
					CreateAt: time.Now().Add(-24 * time.Hour),
				},
				{
					ID:       "user2",
					Name:     "Jane Smith",
					Email:    "jane@example.com",
					Age:      30,
					CreateAt: time.Now().Add(-48 * time.Hour),
				},
			}
			
			// 如果有关键词，进行过滤
			if req.Keyword != "" {
				filtered := make([]UserCreateResponse, 0)
				for _, user := range users {
					if user.Name == req.Keyword || user.Email == req.Keyword {
						filtered = append(filtered, user)
					}
				}
				users = filtered
			}
			
			response := UserQueryResponse{
				Users: users,
				Total: len(users),
				Page:  req.Page,
				Size:  req.PageSize,
			}
			
			return response, nil
		})
	
	return handler
}

// =============== 处理链示例 ===============

// UserValidationChain 用户验证链示例
func UserValidationChain() *GenericChain[UserCreateRequest] {
	return NewGenericChain[UserCreateRequest]("UserValidation").
		Add(func(ctx context.Context, req UserCreateRequest) (UserCreateRequest, error) {
			// 步骤1: 基础验证
			if req.Name == "" || req.Email == "" {
				return req, errors.New("name and email are required")
			}
			return req, nil
		}).
		Add(func(ctx context.Context, req UserCreateRequest) (UserCreateRequest, error) {
			// 步骤2: 邮箱格式验证
			if !isValidEmail(req.Email) {
				return req, errors.New("invalid email format")
			}
			return req, nil
		}).
		Add(func(ctx context.Context, req UserCreateRequest) (UserCreateRequest, error) {
			// 步骤3: 年龄验证
			if req.Age < 1 || req.Age > 150 {
				return req, errors.New("invalid age")
			}
			return req, nil
		})
}

// =============== 管道示例 ===============

// UserProcessingPipeline 用户处理管道示例
func UserProcessingPipeline() *GenericPipeline[UserCreateRequest, UserCreateResponse] {
	return NewGenericPipeline[UserCreateRequest, UserCreateResponse]("UserProcessing").
		AddStage(func(ctx context.Context, input any) (any, error) {
			// 阶段1: 数据验证和清理
			req := input.(UserCreateRequest)
			config.Debug("Pipeline stage 1: Data validation and cleaning")
			
			// 清理数据
			req.Name = fmt.Sprintf("%s", req.Name)
			req.Email = fmt.Sprintf("%s", req.Email)
			
			return req, nil
		}).
		AddStage(func(ctx context.Context, input any) (any, error) {
			// 阶段2: 业务逻辑处理
			req := input.(UserCreateRequest)
			config.Debug("Pipeline stage 2: Business logic processing")
			
			// 模拟业务处理
			time.Sleep(50 * time.Millisecond)
			
			// 转换为中间结果
			intermediate := map[string]any{
				"id":         ShortID(),
				"name":       req.Name,
				"email":      req.Email,
				"age":        req.Age,
				"create_at":  time.Now(),
				"processed":  true,
			}
			
			return intermediate, nil
		}).
		AddStage(func(ctx context.Context, input any) (any, error) {
			// 阶段3: 结果构建
			intermediate := input.(map[string]any)
			config.Debug("Pipeline stage 3: Result building")
			
			response := UserCreateResponse{
				ID:       intermediate["id"].(string),
				Name:     intermediate["name"].(string),
				Email:    intermediate["email"].(string),
				Age:      intermediate["age"].(int),
				CreateAt: intermediate["create_at"].(time.Time),
			}
			
			return response, nil
		})
}

// =============== HTTP处理器包装函数 ===============

// WrapHTTPGenericHandler 将泛型处理器包装为Hertz处理函数
func WrapHTTPGenericHandler[T any, R any](handler HTTPGenericHandler[T, R]) app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		response, err := handler.HandleHTTP(c)
		if err != nil {
			// 错误已经在HandleError中处理了
			return
		}
		
		// 设置响应
		statusCode := 200
		if !response.Success {
			statusCode = response.Code
		}
		
		c.JSON(statusCode, response)
	}
}

// =============== 使用示例函数 ===============

// ExampleUsage 展示如何使用泛型处理器
func ExampleUsage() {
	// 1. 创建处理器实例
	createHandler := CreateUserGenericHandler()
	queryHandler := QueryUsersGenericHandler()
	
	// 2. 包装为HTTP处理函数
	createUserFunc := WrapHTTPGenericHandler[UserCreateRequest, UserCreateResponse](createHandler)
	queryUsersFunc := WrapHTTPGenericHandler[UserQueryRequest, UserQueryResponse](queryHandler)
	
	// 3. 注册路由（示例）
	_ = createUserFunc
	_ = queryUsersFunc
	
	config.Info("Generic handlers created and wrapped successfully")
	
	// 4. 演示处理链使用
	chain := UserValidationChain()
	testRequest := UserCreateRequest{
		Name:     "Test User",
		Email:    "test@example.com",
		Age:      25,
		Password: "password123",
	}
	
	validatedRequest, err := chain.Execute(context.Background(), testRequest)
	if err != nil {
		config.Errorf("Chain validation failed: %v", err)
	} else {
		config.WithFields(map[string]any{
			"name":  validatedRequest.Name,
			"email": validatedRequest.Email,
		}).Info("Chain validation successful")
	}
	
	// 5. 演示管道使用
	pipeline := UserProcessingPipeline()
	result, err := pipeline.Execute(context.Background(), testRequest)
	if err != nil {
		config.Errorf("Pipeline processing failed: %v", err)
	} else {
		config.WithFields(map[string]any{
			"user_id": result.ID,
			"name":    result.Name,
		}).Info("Pipeline processing successful")
	}
}

// isValidEmail 简单的邮箱验证
func isValidEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}