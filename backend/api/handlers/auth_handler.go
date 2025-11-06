package handlers

import (
	"net/http"
	"sync"
	"time"

	"zippifi/backend/api/middleware"
	"zippifi/backend/api/models"
	"zippifi/backend/api/schemas"
	"zippifi/backend/config"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// 内存存储（当数据库不可用时使用）
type memoryStore struct {
	users  map[uint]*models.User
	mutex  sync.RWMutex
	nextID uint
}

var memStore = &memoryStore{
	users:  make(map[uint]*models.User),
	nextID: 1,
}

// AuthHandler 认证处理程序
type AuthHandler struct{}

// NewAuthHandler 创建认证处理程序实例
func NewAuthHandler() *AuthHandler {
	return &AuthHandler{}
}

// Register 用户注册
func (h *AuthHandler) Register(c *gin.Context) {
	var req schemas.RegisterRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, schemas.ErrorResponse{
			Error:   "validation_error",
			Message: err.Error(),
		})
		return
	}

	// 哈希密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, schemas.ErrorResponse{
			Error:   "password_hash_error",
			Message: "密码加密失败",
		})
		return
	}

	// 创建新用户
	user := models.User{
		Username:  req.Username,
		Email:     req.Email,
		Password:  string(hashedPassword),
		CreatedAt: time.Now().Format(time.RFC3339),
		UpdatedAt: time.Now().Format(time.RFC3339),
	}

	// 使用内存存储
	memStore.mutex.Lock()
	defer memStore.mutex.Unlock()

	// 检查用户名是否已存在
	for _, u := range memStore.users {
		if u.Username == req.Username || u.Email == req.Email {
			c.JSON(http.StatusConflict, schemas.ErrorResponse{
				Error:   "user_exists",
				Message: "用户名或邮箱已存在",
			})
			return
		}
	}

	// 分配ID并保存用户
	user.ID = memStore.nextID
	memStore.nextID++
	memStore.users[user.ID] = &user

	c.JSON(http.StatusCreated, schemas.SuccessResponse{
		Message: "注册成功",
		Data: schemas.UserProfile{
			ID:       user.ID,
			Username: user.Username,
			Email:    user.Email,
		},
	})
}

// Login 用户登录
func (h *AuthHandler) Login(c *gin.Context) {
	var req schemas.LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, schemas.ErrorResponse{
			Error:   "validation_error",
			Message: err.Error(),
		})
		return
	}

	// 从内存存储查找用户
	memStore.mutex.RLock()
	var user *models.User
	for _, u := range memStore.users {
		if u.Username == req.Username {
			user = u
			break
		}
	}
	memStore.mutex.RUnlock()

	if user == nil {
		c.JSON(http.StatusUnauthorized, schemas.ErrorResponse{
			Error:   "invalid_credentials",
			Message: "用户名或密码错误",
		})
		return
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, schemas.ErrorResponse{
			Error:   "invalid_credentials",
			Message: "用户名或密码错误",
		})
		return
	}

	// 生成JWT token
	token, err := h.generateToken(user.ID, user.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, schemas.ErrorResponse{
			Error:   "token_generation_error",
			Message: "生成token失败",
		})
		return
	}

	c.JSON(http.StatusOK, schemas.LoginResponse{
		Token: token,
		User: schemas.UserProfile{
			ID:       user.ID,
			Username: user.Username,
			Email:    user.Email,
		},
	})
}

// generateToken 生成JWT token
func (h *AuthHandler) generateToken(userID uint, username string) (string, error) {
	// 设置过期时间
	expirationTime, _ := time.ParseDuration(config.AppConfig.JWTExpiration)
	claims := middleware.JWTClaims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expirationTime)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	// 创建token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// 签名token
	tokenString, err := token.SignedString([]byte(config.AppConfig.JWTSecret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// GetProfile 获取用户信息
func (h *AuthHandler) GetProfile(c *gin.Context) {
	userID := middleware.GetUserID(c)

	// 从内存存储获取用户信息
	memStore.mutex.RLock()
	user, exists := memStore.users[userID]
	memStore.mutex.RUnlock()

	if !exists {
		c.JSON(http.StatusNotFound, schemas.ErrorResponse{
			Error:   "user_not_found",
			Message: "用户不存在",
		})
		return
	}

	c.JSON(http.StatusOK, schemas.UserProfile{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
	})
}
