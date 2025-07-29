package util

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

// SMSConfig 短信配置
type SMSConfig struct {
	URL      string `json:"url"`
	Username string `json:"username"`
	Password string `json:"password"`
	Auth     string `json:"auth"`
	Salt     string `json:"salt"`
	Alias    string `json:"alias"`
	Subject  string `json:"subject"`
}

// SMSResponse 短信响应结构
type SMSResponse struct {
	ResultCode string `json:"resultCode"`
	Result     int    `json:"result"`
	ResultMsg  string `json:"resultMsg"`
	SerialNum  string `json:"serialNum"`
}

// EmailRequest 邮件请求结构
type EmailRequest struct {
	Address string `json:"address"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

// UploadFileResult 文件上传结果
type UploadFileResult struct {
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

// SMSService 短信服务
type SMSService struct {
	config *SMSConfig
}

// NewSMSService 创建短信服务
func NewSMSService(config *SMSConfig) *SMSService {
	if config == nil {
		config = &SMSConfig{
			URL:     "http://localhost:8080",
			Alias:   "system",
			Subject: "通知",
		}
	}
	return &SMSService{config: config}
}

// SendSMS 发送短信
func (s *SMSService) SendSMS(flowID, phoneNumbers string, content map[string]any) (*SMSResponse, error) {
	return s.send(flowID, phoneNumbers, content, map[string]any{
		"alias":   s.config.Alias,
		"subject": s.config.Subject,
	})
}

// SendEmail 发送邮件
func (s *SMSService) SendEmail(email, title, content string) (*SMSResponse, error) {
	flowID := "MAIL_COMMON"
	body := map[string]any{
		"body": content,
	}
	attr := map[string]any{
		"alias":   s.config.Alias,
		"subject": title,
	}
	return s.send(flowID, email, body, attr)
}

// SendEmailWithRequest 使用请求结构发送邮件
func (s *SMSService) SendEmailWithRequest(req *EmailRequest) (*SMSResponse, error) {
	return s.SendEmail(req.Address, req.Title, req.Content)
}

// SendVerificationCode 发送验证码
func (s *SMSService) SendVerificationCode(contact string, codeType int, opType int) (*SMSResponse, error) {
	flowID := "SMS_244545331"
	if codeType == 2 { // 邮件
		flowID = "MAIL_244545331"
	}
	
	return s.sendCodeType(flowID, contact, strconv.Itoa(opType), map[string]any{
		"alias":   s.config.Alias,
		"subject": s.config.Subject,
	})
}

// VerifyCode 验证验证码
func (s *SMSService) VerifyCode(contact, code string, opType int) (*SMSResponse, error) {
	return s.verifyCodeType(contact, code, opType, map[string]any{
		"alias":   s.config.Alias,
		"subject": s.config.Subject,
	})
}

// 内部方法

// send 发送消息
func (s *SMSService) send(flowID, atUsers string, body map[string]any, attr map[string]any) (*SMSResponse, error) {
	timestamp := time.Now().Format("200601021504")
	password := s.generatePassword(timestamp)
	
	attrJSON, err := json.Marshal(attr)
	if err != nil {
		return nil, fmt.Errorf("marshal attr error: %w", err)
	}
	
	bodyJSON, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal body error: %w", err)
	}
	
	data := map[string]any{
		"username":  s.config.Username,
		"timestamp": timestamp,
		"password":  password,
		"flowid":    flowID,
		"atusers":   atUsers,
		"attribute": string(attrJSON),
		"body":      base64.StdEncoding.EncodeToString(bodyJSON),
	}
	
	return s.httpPost(data, s.config.URL+"/api/message/v1/send")
}

// sendCodeType 发送验证码类型消息
func (s *SMSService) sendCodeType(flowID, atUsers, codeType string, attr map[string]any) (*SMSResponse, error) {
	timestamp := time.Now().Format("200601021504")
	password := s.generatePassword(timestamp)
	
	attrJSON, err := json.Marshal(attr)
	if err != nil {
		return nil, fmt.Errorf("marshal attr error: %w", err)
	}
	
	data := map[string]any{
		"username":  s.config.Username,
		"timestamp": timestamp,
		"password":  password,
		"flowId":    flowID,
		"type":      codeType,
		"atusers":   atUsers,
		"attribute": base64.StdEncoding.EncodeToString(attrJSON),
	}
	
	return s.httpPost(data, s.config.URL+"/api/message/v1/codetype/send")
}

// verifyCodeType 验证验证码类型
func (s *SMSService) verifyCodeType(atUsers, code string, opType int, attr map[string]any) (*SMSResponse, error) {
	timestamp := time.Now().Format("200601021504")
	password := s.generatePassword(timestamp)
	
	attrJSON, err := json.Marshal(attr)
	if err != nil {
		return nil, fmt.Errorf("marshal attr error: %w", err)
	}
	
	data := map[string]any{
		"username":  s.config.Username,
		"timestamp": timestamp,
		"password":  password,
		"code":      code,
		"type":      strconv.Itoa(opType),
		"atusers":   atUsers,
		"attribute": base64.StdEncoding.EncodeToString(attrJSON),
	}
	
	return s.httpPost(data, s.config.URL+"/api/message/v1/codetype/verify")
}

// generatePassword 生成密码
func (s *SMSService) generatePassword(timestamp string) string {
	return MD5String(timestamp + s.config.Password + s.config.Salt)
}

// httpPost 发送HTTP POST请求
func (s *SMSService) httpPost(data map[string]any, url string) (*SMSResponse, error) {
	// 添加随机参数
	pm := s.randomHexString(6)
	rm := s.randomHexString(18)
	separator := "?"
	if strings.Contains(url, "?") {
		separator = "&"
	}
	url += separator + pm + "=" + rm
	
	// 序列化数据
	jsonData, err := json.Marshal(data)
	if err != nil {
		return &SMSResponse{
			Result:    500,
			ResultMsg: err.Error(),
		}, err
	}
	
	// 创建HTTP客户端
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	
	// 创建请求
	req, err := http.NewRequest("POST", url, bytes.NewReader(jsonData))
	if err != nil {
		return &SMSResponse{
			Result:    500,
			ResultMsg: err.Error(),
		}, err
	}
	
	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	if s.config.Auth != "" {
		req.Header.Set("Authorization", s.config.Auth)
	}
	
	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		return &SMSResponse{
			Result:    500,
			ResultMsg: err.Error(),
		}, err
	}
	defer resp.Body.Close()
	
	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return &SMSResponse{
			Result:    500,
			ResultMsg: err.Error(),
		}, err
	}
	
	// 解析响应
	var response SMSResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return &SMSResponse{
			Result:    500,
			ResultMsg: err.Error(),
		}, err
	}
	
	return &response, nil
}

// randomHexString 生成随机十六进制字符串
func (s *SMSService) randomHexString(length int) string {
	bytes := make([]byte, length/2)
	if _, err := rand.Read(bytes); err != nil {
		return ""
	}
	return hex.EncodeToString(bytes)
}

// GenerateVerificationCode 生成验证码
func GenerateVerificationCode(length int) string {
	if length <= 0 {
		length = 6
	}
	
	const charset = "0123456789"
	b := make([]byte, length)
	for i := range b {
		num := make([]byte, 1)
		rand.Read(num)
		b[i] = charset[num[0]%byte(len(charset))]
	}
	return string(b)
}

// GenerateRandomString 生成随机字符串
func GenerateRandomString(length int, charset string) string {
	if charset == "" {
		charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	}
	
	b := make([]byte, length)
	for i := range b {
		num := make([]byte, 1)
		rand.Read(num)
		b[i] = charset[num[0]%byte(len(charset))]
	}
	return string(b)
}

// Base64Encode Base64编码
func Base64Encode(data string) string {
	return base64.StdEncoding.EncodeToString([]byte(data))
}

// Base64Decode Base64解码
func Base64Decode(data string) (string, error) {
	decoded, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return "", err
	}
	return string(decoded), nil
}

// PathExistsWithError 检查路径是否存在并返回错误信息
func PathExistsWithError(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// 全局默认短信服务实例
var DefaultSMSService *SMSService

// InitSMSService 初始化默认短信服务
func InitSMSService(config *SMSConfig) {
	DefaultSMSService = NewSMSService(config)
}

// 便捷函数
func SendSMS(flowID, phoneNumbers string, content map[string]any) (*SMSResponse, error) {
	if DefaultSMSService == nil {
		return nil, fmt.Errorf("SMS service not initialized")
	}
	return DefaultSMSService.SendSMS(flowID, phoneNumbers, content)
}

func SendEmail(email, title, content string) (*SMSResponse, error) {
	if DefaultSMSService == nil {
		return nil, fmt.Errorf("SMS service not initialized")
	}
	return DefaultSMSService.SendEmail(email, title, content)
}

func SendVerificationCode(contact string, codeType int, opType int) (*SMSResponse, error) {
	if DefaultSMSService == nil {
		return nil, fmt.Errorf("SMS service not initialized")
	}
	return DefaultSMSService.SendVerificationCode(contact, codeType, opType)
}

func VerifyCode(contact, code string, opType int) (*SMSResponse, error) {
	if DefaultSMSService == nil {
		return nil, fmt.Errorf("SMS service not initialized")
	}
	return DefaultSMSService.VerifyCode(contact, code, opType)
}