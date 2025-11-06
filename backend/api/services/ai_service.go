package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"

	"zippifi/backend/config"

	"github.com/go-resty/resty/v2"
)

// DeepSeekMessage DeepSeek消息格式
type DeepSeekMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// DeepSeekRequest DeepSeek API请求
type DeepSeekRequest struct {
	Model    string          `json:"model"`
	Messages []DeepSeekMessage `json:"messages"`
	Stream   bool            `json:"stream"`
}

// DeepSeekResponse DeepSeek API响应
type DeepSeekResponse struct {
	Choices []struct {
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

// AIService AI服务结构体
type AIService struct {
	client *resty.Client
}

// NewAIService 创建AI服务实例
func NewAIService() *AIService {
	client := resty.New()
	return &AIService{
		client: client,
	}
}

// GenerateResponse 生成AI响应
func (s *AIService) GenerateResponse(systemPrompt, userMessage string, conversationHistory []DeepSeekMessage) (string, error) {
	// 构建消息列表
	messages := []DeepSeekMessage{
		{Role: "system", Content: systemPrompt},
	}

	// 添加历史对话
	messages = append(messages, conversationHistory...)
	
	// 添加最新用户消息
	messages = append(messages, DeepSeekMessage{
		Role:    "user",
		Content: userMessage,
	})

	// 构建请求体
	request := DeepSeekRequest{
		Model:    "deepseek-chat",
		Messages: messages,
		Stream:   false,
	}

	requestBody, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// 发送请求到DeepSeek API
	response, err := s.client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Authorization", fmt.Sprintf("Bearer %s", config.AppConfig.DeepSeekAPIKey)).
		SetBody(bytes.NewBuffer(requestBody)).
		Post(config.AppConfig.DeepSeekAPIURL)

	if err != nil {
		return "", fmt.Errorf("failed to send request to DeepSeek API: %w", err)
	}

	if response.StatusCode() != 200 {
		return "", fmt.Errorf("DeepSeek API returned non-200 status: %d, response: %s", 
			response.StatusCode(), response.String())
	}

	// 解析响应
	var deepSeekResponse DeepSeekResponse
	err = json.Unmarshal(response.Body(), &deepSeekResponse)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(deepSeekResponse.Choices) == 0 || len(deepSeekResponse.Choices[0].Message.Content) == 0 {
		return "", fmt.Errorf("invalid response from DeepSeek API: no choices or content")
	}

	log.Printf("AI response generated successfully")
	return deepSeekResponse.Choices[0].Message.Content, nil
}

// GenerateAgentResponse 生成AI代理响应
func (s *AIService) GenerateAgentResponse(agentPrompt, userMessage string) (string, error) {
	return s.GenerateResponse(agentPrompt, userMessage, []DeepSeekMessage{})
}