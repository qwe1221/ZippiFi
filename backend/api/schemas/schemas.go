package schemas

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// RegisterRequest 注册请求
type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	Token string      `json:"token"`
	User  UserProfile `json:"user"`
}

// UserProfile 用户信息
type UserProfile struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

// CreateAgentRequest 创建代理请求
type CreateAgentRequest struct {
	Name         string `json:"name" binding:"required"`
	SystemPrompt string `json:"system_prompt" binding:"required"`
}

// CreateConversationRequest 创建对话请求
type CreateConversationRequest struct {
	AgentID uint   `json:"agent_id" binding:"required"`
	Title   string `json:"title" binding:"required"`
}

// SendMessageRequest 发送消息请求
type SendMessageRequest struct {
	ConversationID uint   `json:"conversation_id" binding:"required"`
	Content        string `json:"content" binding:"required"`
}

// MessageResponse 消息响应
type MessageResponse struct {
	ID        uint   `json:"id"`
	Role      string `json:"role"`
	Content   string `json:"content"`
	CreatedAt string `json:"created_at"`
}

// ChatResponse 聊天响应
type ChatResponse struct {
	UserMessage    MessageResponse `json:"user_message"`
	AssistantReply MessageResponse `json:"assistant_reply"`
}

// ErrorResponse 错误响应
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

// SuccessResponse 成功响应
type SuccessResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}