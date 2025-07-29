package util

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// MessageResponse 消息响应结构
type MessageResponse struct {
	ResultCode string `json:"resultCode"` // 操作业务码
	Result     int    `json:"result"`     // 操作业务码
	ResultMsg  string `json:"resultMsg"`  // 描述信息
	SerialNum  string `json:"serialNum"`  // 业务流水号
}

// UploadResult 文件上传结果
type UploadResult struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	File struct {
		Name1 string `json:"name1"` // 原始文件名称
		Name2 string `json:"name2"` // 保存文件名称
		Ext   string `json:"ext"`   // 文件后缀
		Size  int64  `json:"size"`  // 文件大小
		Url1  string `json:"url1"`  // 文件相对路径
		Url2  string `json:"url2"`  // 文件绝对路径
	} `json:"file"`
}

// MessageService 消息服务
type MessageService struct {
	BaseURL   string
	Username  string
	Password  string
	Salt      string
	Auth      string
	Timeout   time.Duration
	client    *http.Client
}

// NewMessageService 创建消息服务
func NewMessageService(baseURL, username, password, salt, auth string) *MessageService {
	return &MessageService{
		BaseURL:  strings.TrimRight(baseURL, "/"),
		Username: username,
		Password: password,
		Salt:     salt,
		Auth:     auth,
		Timeout:  5 * time.Second,
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// SendMessage 发送消息
func (ms *MessageService) SendMessage(flowId, atUsers string, body map[string]any, attr map[string]any) *MessageResponse {
	timestamp := time.Now().Format("200601021504")
	password := MD5String(timestamp + ms.Password + ms.Salt)
	
	attrJSON, _ := json.Marshal(attr)
	bodyJSON, _ := json.Marshal(body)
	bodyBase64 := Base64EncodeMessage(string(bodyJSON))
	
	data := map[string]any{
		"username":  ms.Username,
		"timestamp": timestamp,
		"password":  password,
		"flowid":    flowId,
		"atusers":   atUsers,
		"attribute": string(attrJSON),
		"body":      bodyBase64,
	}
	
	return ms.post(data, ms.BaseURL+"/api/message/v1/send")
}

// SendCodeType 发送验证码
func (ms *MessageService) SendCodeType(flowId, atUsers, codeType string, attr map[string]any) *MessageResponse {
	timestamp := time.Now().Format("200601021504")
	password := MD5String(timestamp + ms.Password + ms.Salt)
	
	attrJSON, _ := json.Marshal(attr)
	
	data := map[string]any{
		"username":  ms.Username,
		"timestamp": timestamp,
		"password":  password,
		"flowid":    flowId,
		"atusers":   atUsers,
		"attribute": string(attrJSON),
		"type":      codeType,
	}
	
	return ms.post(data, ms.BaseURL+"/api/message/v1/codetype/send")
}

// VerifyCodeType 验证验证码
func (ms *MessageService) VerifyCodeType(atUsers, code, codeType string, attr map[string]any) *MessageResponse {
	timestamp := time.Now().Format("200601021504")
	password := MD5String(timestamp + ms.Password + ms.Salt)
	
	attrJSON, _ := json.Marshal(attr)
	
	data := map[string]any{
		"username":  ms.Username,
		"timestamp": timestamp,
		"password":  password,
		"atusers":   atUsers,
		"attribute": string(attrJSON),
		"type":      codeType,
		"code":      code,
	}
	
	return ms.post(data, ms.BaseURL+"/api/message/v1/codetype/verify")
}

// SendEmail 发送邮件
func (ms *MessageService) SendEmail(address, subject, content string, attr map[string]any) *MessageResponse {
	flowId := "MAIL_COMMON"
	body := map[string]any{
		"body":    content,
		"subject": subject,
	}
	
	if attr == nil {
		attr = make(map[string]any)
	}
	attr["subject"] = subject
	
	return ms.SendMessage(flowId, address, body, attr)
}

// SendSMS 发送短信
func (ms *MessageService) SendSMS(phone, content string, attr map[string]any) *MessageResponse {
	flowId := "SMS_COMMON"
	body := map[string]any{
		"body": content,
	}
	
	return ms.SendMessage(flowId, phone, body, attr)
}

// SendSMSCode 发送短信验证码
func (ms *MessageService) SendSMSCode(phone, opType string, attr map[string]any) *MessageResponse {
	flowId := "SMS_244545331"
	return ms.SendCodeType(flowId, phone, opType, attr)
}

// SendEmailCode 发送邮件验证码
func (ms *MessageService) SendEmailCode(email, opType string, attr map[string]any) *MessageResponse {
	flowId := "MAIL_244545331"
	return ms.SendCodeType(flowId, email, opType, attr)
}

// VerifySMSCode 验证短信验证码
func (ms *MessageService) VerifySMSCode(phone, code, opType string, attr map[string]any) *MessageResponse {
	return ms.VerifyCodeType(phone, code, opType, attr)
}

// VerifyEmailCode 验证邮件验证码
func (ms *MessageService) VerifyEmailCode(email, code, opType string, attr map[string]any) *MessageResponse {
	return ms.VerifyCodeType(email, code, opType, attr)
}

// post 发送POST请求
func (ms *MessageService) post(data map[string]any, url string) *MessageResponse {
	// 添加随机参数防止缓存
	randomParam := RandomString(6)
	randomValue := RandomString(18)
	separator := "?"
	if strings.Contains(url, "?") {
		separator = "&"
	}
	url += separator + randomParam + "=" + randomValue
	
	response := &MessageResponse{}
	
	// 序列化数据
	jsonData, err := json.Marshal(data)
	if err != nil {
		response.Result = 500
		response.ResultMsg = err.Error()
		return response
	}
	
	// 创建请求
	req, err := http.NewRequest("POST", url, bytes.NewReader(jsonData))
	if err != nil {
		response.Result = 500
		response.ResultMsg = err.Error()
		return response
	}
	
	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	if ms.Auth != "" {
		req.Header.Set("Authorization", ms.Auth)
	}
	
	// 发送请求
	resp, err := ms.client.Do(req)
	if err != nil {
		response.Result = 500
		response.ResultMsg = err.Error()
		return response
	}
	defer resp.Body.Close()
	
	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		response.Result = 500
		response.ResultMsg = err.Error()
		return response
	}
	
	// 解析响应
	err = json.Unmarshal(body, response)
	if err != nil {
		response.Result = 500
		response.ResultMsg = err.Error()
		return response
	}
	
	return response
}

// 便捷函数

// Base64EncodeMessage base64编码（消息服务专用）
func Base64EncodeMessage(data string) string {
	return base64.StdEncoding.EncodeToString([]byte(data))
}

// Base64DecodeMessage base64解码（消息服务专用）
func Base64DecodeMessage(data string) (string, error) {
	decoded, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return "", err
	}
	return string(decoded), nil
}

// RandomString 生成随机字符串
func RandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[rand.Intn(len(charset))]
	}
	return string(result)
}

// GetOpTypeName 获取操作类型名称
func GetOpTypeName(opType int) string {
	switch opType {
	case 1:
		return "注册"
	case 2:
		return "忘记密码"
	case 3:
		return "绑定"
	default:
		return "未知"
	}
}

// IsValidOpType 验证操作类型是否有效
func IsValidOpType(opType int) bool {
	return opType >= 1 && opType <= 3
}

// FormatOpType 格式化操作类型
func FormatOpType(opType int) string {
	return strconv.Itoa(opType)
}

// ParseOpType 解析操作类型
func ParseOpType(opType string) (int, error) {
	return strconv.Atoi(opType)
}