package middleware

import (
	"context"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
	"hertz-controller/framework/config"
)

// AuthMiddleware 认证中间件 - 进行身份验证
func AuthMiddleware(skipPaths ...string) Middleware {
	skipMap := make(map[string]bool)
	for _, path := range skipPaths {
		skipMap[path] = true
	}
	
	return func(c context.Context, ctx *app.RequestContext) {
		path := string(ctx.Path())
		
		if skipMap[path] {
			ctx.Next(c)
			return
		}
		
		// 获取请求信息用于日志
		clientIP := ctx.ClientIP()
		userAgent := string(ctx.UserAgent())
		method := string(ctx.Method())
		requestID := ctx.GetString("request_id")
		
		token := ctx.GetHeader("Authorization")
		if string(token) == "" {
			// 使用单例日志系统记录未授权访问
			config.WithFields(map[string]any{
				"event":      "auth_missing_token",
				"client_ip":  clientIP,
				"path":       path,
				"method":     method,
				"user_agent": userAgent,
				"request_id": requestID,
			}).Warn("Authentication failed: missing token")
			
			ctx.JSON(401, map[string]string{
				"error": "未授权访问",
			})
			ctx.Abort()
			return
		}
		
		// 简单的token验证（实际应用中应该使用JWT或其他安全机制）
		tokenValue := strings.TrimPrefix(string(token), "Bearer ")
		if tokenValue != "valid-token" {
			// 使用单例日志系统记录无效token
			config.WithFields(map[string]any{
				"event":      "auth_invalid_token",
				"client_ip":  clientIP,
				"path":       path,
				"method":     method,
				"user_agent": userAgent,
				"request_id": requestID,
				"token_prefix": tokenValue[:min(len(tokenValue), 10)], // 只记录token前缀
			}).Warn("Authentication failed: invalid token")
			
			ctx.JSON(401, map[string]string{
				"error": "无效的令牌",
			})
			ctx.Abort()
			return
		}
		
		// 认证成功，记录日志
		config.WithFields(map[string]any{
			"event":      "auth_success",
			"client_ip":  clientIP,
			"path":       path,
			"method":     method,
			"request_id": requestID,
		}).Debug("Authentication successful")
		
		// 将用户信息设置到上下文（简化示例）
		ctx.Set("user_id", "authenticated_user")
		ctx.Set("authenticated", true)
		
		ctx.Next(c)
	}
}

// JWTAuthMiddleware JWT认证中间件（简化版）
func JWTAuthMiddleware(secretKey string, skipPaths ...string) Middleware {
	skipMap := make(map[string]bool)
	for _, path := range skipPaths {
		skipMap[path] = true
	}
	
	return func(c context.Context, ctx *app.RequestContext) {
		path := string(ctx.Path())
		
		if skipMap[path] {
			ctx.Next(c)
			return
		}
		
		// 获取请求信息
		clientIP := ctx.ClientIP()
		userAgent := string(ctx.UserAgent())
		method := string(ctx.Method())
		requestID := ctx.GetString("request_id")
		
		token := ctx.GetHeader("Authorization")
		if len(token) == 0 {
			config.WithFields(map[string]any{
				"event":      "jwt_missing_token",
				"client_ip":  clientIP,
				"path":       path,
				"method":     method,
				"user_agent": userAgent,
				"request_id": requestID,
			}).Warn("JWT authentication failed: missing token")
			
			ctx.JSON(401, map[string]any{
				"error": "Missing authentication token",
				"code":  "AUTH_TOKEN_REQUIRED",
			})
			ctx.Abort()
			return
		}
		
		tokenString := strings.TrimPrefix(string(token), "Bearer ")
		
		// 这里应该使用真正的JWT验证库，比如github.com/golang-jwt/jwt
		// 为了演示，我们使用简化的验证
		if !validateJWT(tokenString, secretKey) {
			config.WithFields(map[string]any{
				"event":        "jwt_invalid_token", 
				"client_ip":    clientIP,
				"path":         path,
				"method":       method,
				"user_agent":   userAgent,
				"request_id":   requestID,
				"token_prefix": tokenString[:min(len(tokenString), 10)],
			}).Warn("JWT authentication failed: invalid token")
			
			ctx.JSON(401, map[string]any{
				"error": "Invalid authentication token",
				"code":  "AUTH_TOKEN_INVALID",
			})
			ctx.Abort()
			return
		}
		
		// JWT验证成功
		config.WithFields(map[string]any{
			"event":      "jwt_auth_success",
			"client_ip":  clientIP,
			"path":       path,
			"method":     method,
			"request_id": requestID,
		}).Debug("JWT authentication successful")
		
		// 这里应该从JWT中解析用户信息
		ctx.Set("user_id", "jwt_user")
		ctx.Set("authenticated", true)
		ctx.Set("auth_method", "jwt")
		
		ctx.Next(c)
	}
}

// AdminAuthMiddleware 管理员认证中间件
func AdminAuthMiddleware() Middleware {
	return func(c context.Context, ctx *app.RequestContext) {
		// 检查是否已经通过基础认证
		if !ctx.GetBool("authenticated") {
			config.WithFields(map[string]any{
				"event":      "admin_auth_not_authenticated",
				"client_ip":  ctx.ClientIP(),
				"path":       string(ctx.Path()),
				"request_id": ctx.GetString("request_id"),
			}).Warn("Admin access denied: not authenticated")
			
			ctx.JSON(401, map[string]any{
				"error": "Authentication required",
				"code":  "AUTH_REQUIRED",
			})
			ctx.Abort()
			return
		}
		
		// 检查管理员权限（这里使用简化的检查）
		adminToken := ctx.GetHeader("X-Admin-Token")
		if string(adminToken) != "admin-secret-token" {
			config.WithFields(map[string]any{
				"event":      "admin_auth_insufficient_privileges",
				"client_ip":  ctx.ClientIP(),
				"path":       string(ctx.Path()),
				"user_id":    ctx.GetString("user_id"),
				"request_id": ctx.GetString("request_id"),
			}).Warn("Admin access denied: insufficient privileges")
			
			ctx.JSON(403, map[string]any{
				"error": "Admin privileges required",
				"code":  "AUTH_INSUFFICIENT_PRIVILEGES",
			})
			ctx.Abort()
			return
		}
		
		config.WithFields(map[string]any{
			"event":      "admin_auth_success",
			"client_ip":  ctx.ClientIP(),
			"path":       string(ctx.Path()),
			"user_id":    ctx.GetString("user_id"),
			"request_id": ctx.GetString("request_id"),
		}).Info("Admin authentication successful")
		
		ctx.Set("is_admin", true)
		ctx.Next(c)
	}
}

// validateJWT 简化的JWT验证函数（实际应用中应使用专业的JWT库）
func validateJWT(token, secretKey string) bool {
	// 这里应该实现真正的JWT验证逻辑
	// 为了演示，我们只做简单的字符串检查
	return len(token) > 10 && strings.Contains(token, secretKey)
}

// min 辅助函数
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}