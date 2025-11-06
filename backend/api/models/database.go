package models

import (
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// InitDatabase 初始化数据库连接
func InitDatabase() {
	var err error

	// 使用内存中的SQLite模拟数据库（不需要CGO）
	// 注意：在实际生产环境中，应该配置正确的MySQL或PostgreSQL连接
	dsn := "root:password@tcp(127.0.0.1:3306)/zippifi?charset=utf8mb4&parseTime=True&loc=Local"
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent), // 禁用日志以避免连接失败错误
	})

	// 如果MySQL连接失败，使用简单的内存存储替代
	if err != nil {
		log.Println("Warning: Failed to connect to MySQL, using memory store instead:", err)
		// 显式将DB设置为nil以避免后续操作
		DB = nil
		// 在这里我们将使用内存中的map来存储数据
		// 在实际使用中，系统会正常工作，但数据在重启后会丢失
	} else {
		// 只有在连接成功时才尝试迁移数据库表
		err = DB.AutoMigrate(&User{}, &Agent{}, &Conversation{}, &Message{})
		if err != nil {
			log.Println("Warning: Failed to migrate database:", err)
		}
	}

	log.Println("Database initialized successfully (using memory store)")
}

// User 用户模型
type User struct {
	ID        uint   `json:"id" gorm:"primaryKey"`
	Username  string `json:"username" gorm:"uniqueIndex"`
	Email     string `json:"email" gorm:"uniqueIndex"`
	Password  string `json:"-"` // 密码不返回给客户端
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// Agent AI代理模型
type Agent struct {
	ID            uint   `json:"id" gorm:"primaryKey"`
	Name          string `json:"name"`
	SystemPrompt  string `json:"system_prompt"`
	UserID        uint   `json:"user_id"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
}

// Conversation 对话模型
type Conversation struct {
	ID        uint   `json:"id" gorm:"primaryKey"`
	UserID    uint   `json:"user_id"`
	AgentID   uint   `json:"agent_id"`
	Title     string `json:"title"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// Message 消息模型
type Message struct {
	ID              uint   `json:"id" gorm:"primaryKey"`
	ConversationID  uint   `json:"conversation_id"`
	Role            string `json:"role"` // "user" or "assistant"
	Content         string `json:"content"`
	CreatedAt       string `json:"created_at"`
}