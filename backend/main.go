package main

import (
	"fmt"
	"log"

	"zippifi/backend/api/handlers"
	"zippifi/backend/api/middleware"
	"zippifi/backend/api/models"
	"zippifi/backend/config"

	"github.com/gin-gonic/gin"
)

func main() {
	// 加载配置
	config.LoadConfig()

	// 初始化数据库
	models.InitDatabase()

	// 设置Gin模式
	if config.AppConfig.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	// 创建Gin引擎
	r := gin.Default()

	// 添加CORS中间件
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"message": "ZippiFi backend service is running",
		})
	})

	// 创建处理程序
	authHandler := handlers.NewAuthHandler()
	agentHandler := handlers.NewAgentHandler()

	// 公开路由
	auth := r.Group("/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
		auth.GET("/profile", middleware.AuthMiddleware(), authHandler.GetProfile)
	}

	// 需要认证的路由
	protected := r.Group("/api")
	protected.Use(middleware.AuthMiddleware())
	{
		// 代理管理
		protected.POST("/agents", agentHandler.CreateAgent)
		protected.GET("/agents", agentHandler.GetAgents)

		// 对话管理
		protected.POST("/conversations", agentHandler.CreateConversation)
		protected.GET("/conversations", agentHandler.GetConversations)
		protected.GET("/conversations/:id/messages", agentHandler.GetConversationMessages)

		// 消息交互
		protected.POST("/messages", agentHandler.SendMessage)
	}

	// 启动服务器
	address := fmt.Sprintf(":%s", config.AppConfig.Port)
	log.Printf("Server starting on %s", address)
	log.Printf("Environment: %s", config.AppConfig.Environment)

	if err := r.Run(address); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
