package handlers

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"zippifi/backend/api/middleware"
	"zippifi/backend/api/models"
	"zippifi/backend/api/schemas"
	"zippifi/backend/api/services"

	"github.com/gin-gonic/gin"
)

// 扩展内存存储以支持代理、对话和消息
type agentMemoryStore struct {
	agents             map[uint]*models.Agent
	conversations      map[uint]*models.Conversation
	messages           map[uint]*models.Message
	mutex              sync.RWMutex
	nextAgentID        uint
	nextConversationID uint
	nextMessageID      uint
}

var agentMemStore = &agentMemoryStore{
	agents:             make(map[uint]*models.Agent),
	conversations:      make(map[uint]*models.Conversation),
	messages:           make(map[uint]*models.Message),
	nextAgentID:        1,
	nextConversationID: 1,
	nextMessageID:      1,
}

// AgentHandler AI代理处理程序
type AgentHandler struct {
	aiService *services.AIService
}

// NewAgentHandler 创建AI代理处理程序实例
func NewAgentHandler() *AgentHandler {
	return &AgentHandler{
		aiService: services.NewAIService(),
	}
}

// CreateAgent 创建AI代理
func (h *AgentHandler) CreateAgent(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var req schemas.CreateAgentRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, schemas.ErrorResponse{
			Error:   "validation_error",
			Message: err.Error(),
		})
		return
	}

	// 创建代理
	agent := &models.Agent{
		Name:         req.Name,
		SystemPrompt: req.SystemPrompt,
		UserID:       userID,
		CreatedAt:    time.Now().Format(time.RFC3339),
		UpdatedAt:    time.Now().Format(time.RFC3339),
	}

	// 使用内存存储
	agentMemStore.mutex.Lock()
	defer agentMemStore.mutex.Unlock()

	// 分配ID并保存代理
	agent.ID = agentMemStore.nextAgentID
	agentMemStore.nextAgentID++
	agentMemStore.agents[agent.ID] = agent

	c.JSON(http.StatusCreated, schemas.SuccessResponse{
		Message: "代理创建成功",
		Data:    agent,
	})
}

// GetAgents 获取用户的所有代理
func (h *AgentHandler) GetAgents(c *gin.Context) {
	userID := middleware.GetUserID(c)

	// 从内存存储获取用户的所有代理
	agentMemStore.mutex.RLock()
	var agents []models.Agent
	for _, agent := range agentMemStore.agents {
		if agent.UserID == userID {
			agents = append(agents, *agent)
		}
	}
	agentMemStore.mutex.RUnlock()

	c.JSON(http.StatusOK, agents)
}

// CreateConversation 创建对话
func (h *AgentHandler) CreateConversation(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var req schemas.CreateConversationRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, schemas.ErrorResponse{
			Error:   "validation_error",
			Message: err.Error(),
		})
		return
	}

	// 使用内存存储
	agentMemStore.mutex.Lock()
	defer agentMemStore.mutex.Unlock()

	// 验证代理是否属于该用户
	agent, exists := agentMemStore.agents[req.AgentID]
	if !exists || agent.UserID != userID {
		c.JSON(http.StatusNotFound, schemas.ErrorResponse{
			Error:   "agent_not_found",
			Message: "代理不存在或无权限访问",
		})
		return
	}

	// 创建对话
	conversation := &models.Conversation{
		UserID:    userID,
		AgentID:   req.AgentID,
		Title:     req.Title,
		CreatedAt: time.Now().Format(time.RFC3339),
		UpdatedAt: time.Now().Format(time.RFC3339),
	}

	// 分配ID并保存对话
	conversation.ID = agentMemStore.nextConversationID
	agentMemStore.nextConversationID++
	agentMemStore.conversations[conversation.ID] = conversation

	c.JSON(http.StatusCreated, schemas.SuccessResponse{
		Message: "对话创建成功",
		Data:    conversation,
	})
}

// SendMessage 发送消息并获取AI回复
func (h *AgentHandler) SendMessage(c *gin.Context) {
	userID := middleware.GetUserID(c)
	var req schemas.SendMessageRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, schemas.ErrorResponse{
			Error:   "validation_error",
			Message: err.Error(),
		})
		return
	}

	// 验证对话是否属于该用户
	agentMemStore.mutex.RLock()
	conversation, exists := agentMemStore.conversations[req.ConversationID]
	if !exists || conversation.UserID != userID {
		agentMemStore.mutex.RUnlock()
		c.JSON(http.StatusNotFound, schemas.ErrorResponse{
			Error:   "conversation_not_found",
			Message: "对话不存在或无权限访问",
		})
		return
	}

	// 获取代理信息
	agent, exists := agentMemStore.agents[conversation.AgentID]
	if !exists {
		agentMemStore.mutex.RUnlock()
		c.JSON(http.StatusInternalServerError, schemas.ErrorResponse{
			Error:   "agent_not_found",
			Message: "代理信息获取失败",
		})
		return
	}
	agentMemStore.mutex.RUnlock()

	// 调用AI服务生成回复
	aiResponse, err := h.aiService.GenerateAgentResponse(agent.SystemPrompt, req.Content)
	if err != nil {
		c.JSON(http.StatusInternalServerError, schemas.ErrorResponse{
			Error:   "ai_generation_error",
			Message: "AI生成回复失败: " + err.Error(),
		})
		return
	}

	// 保存用户消息和AI回复
	agentMemStore.mutex.Lock()
	defer agentMemStore.mutex.Unlock()

	// 保存用户消息
	userMessage := &models.Message{
		ConversationID: req.ConversationID,
		Role:           "user",
		Content:        req.Content,
		CreatedAt:      time.Now().Format(time.RFC3339),
	}
	userMessage.ID = agentMemStore.nextMessageID
	agentMemStore.nextMessageID++
	agentMemStore.messages[userMessage.ID] = userMessage

	// 保存AI回复
	assistantMessage := &models.Message{
		ConversationID: req.ConversationID,
		Role:           "assistant",
		Content:        aiResponse,
		CreatedAt:      time.Now().Format(time.RFC3339),
	}
	assistantMessage.ID = agentMemStore.nextMessageID
	agentMemStore.nextMessageID++
	agentMemStore.messages[assistantMessage.ID] = assistantMessage

	// 更新对话时间
	conversation.UpdatedAt = time.Now().Format(time.RFC3339)

	// 返回响应
	c.JSON(http.StatusOK, schemas.ChatResponse{
		UserMessage: schemas.MessageResponse{
			ID:        userMessage.ID,
			Role:      userMessage.Role,
			Content:   userMessage.Content,
			CreatedAt: userMessage.CreatedAt,
		},
		AssistantReply: schemas.MessageResponse{
			ID:        assistantMessage.ID,
			Role:      assistantMessage.Role,
			Content:   assistantMessage.Content,
			CreatedAt: assistantMessage.CreatedAt,
		},
	})
}

// GetConversations 获取用户的所有对话
func (h *AgentHandler) GetConversations(c *gin.Context) {
	userID := middleware.GetUserID(c)

	// 从内存存储获取用户的所有对话
	agentMemStore.mutex.RLock()
	var conversations []models.Conversation
	for _, conv := range agentMemStore.conversations {
		if conv.UserID == userID {
			conversations = append(conversations, *conv)
		}
	}
	agentMemStore.mutex.RUnlock()

	// 简单的时间排序（降序）
	// 在实际应用中应该使用更高效的排序算法
	for i := 0; i < len(conversations); i++ {
		for j := i + 1; j < len(conversations); j++ {
			if conversations[i].UpdatedAt < conversations[j].UpdatedAt {
				conversations[i], conversations[j] = conversations[j], conversations[i]
			}
		}
	}

	c.JSON(http.StatusOK, conversations)
}

// GetConversationMessages 获取对话历史消息
func (h *AgentHandler) GetConversationMessages(c *gin.Context) {
	userID := middleware.GetUserID(c)
	conversationIDStr := c.Param("id")

	// 将字符串转换为uint
	var conversationID uint
	_, err := fmt.Sscanf(conversationIDStr, "%d", &conversationID)
	if err != nil {
		c.JSON(http.StatusBadRequest, schemas.ErrorResponse{
			Error:   "invalid_conversation_id",
			Message: "无效的对话ID",
		})
		return
	}

	// 验证对话是否属于该用户
	agentMemStore.mutex.RLock()
	conversation, exists := agentMemStore.conversations[conversationID]
	if !exists || conversation.UserID != userID {
		agentMemStore.mutex.RUnlock()
		c.JSON(http.StatusNotFound, schemas.ErrorResponse{
			Error:   "conversation_not_found",
			Message: "对话不存在或无权限访问",
		})
		return
	}

	// 获取对话的所有消息
	var messages []models.Message
	for _, msg := range agentMemStore.messages {
		if msg.ConversationID == conversationID {
			messages = append(messages, *msg)
		}
	}
	agentMemStore.mutex.RUnlock()

	// 简单的时间排序（升序）
	for i := 0; i < len(messages); i++ {
		for j := i + 1; j < len(messages); j++ {
			if messages[i].CreatedAt > messages[j].CreatedAt {
				messages[i], messages[j] = messages[j], messages[i]
			}
		}
	}

	c.JSON(http.StatusOK, messages)
}
