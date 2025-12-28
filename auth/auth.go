package auth

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"os"
	"sync"
	"time"
)

// AuthConfig 认证配置
type AuthConfig struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Session 会话
type Session struct {
	Token     string
	ExpiresAt time.Time
}

// Captcha 验证码
type Captcha struct {
	Code      string
	ExpiresAt time.Time
}

// AuthManager 认证管理器
type AuthManager struct {
	mu       sync.RWMutex
	config   AuthConfig
	sessions map[string]Session
	captchas map[string]Captcha
	filePath string
}

var (
	authManager *AuthManager
	authOnce    sync.Once
)

// GetAuthManager 获取认证管理器单例
func GetAuthManager() *AuthManager {
	authOnce.Do(func() {
		authManager = &AuthManager{
			filePath: "data/auth.json",
			sessions: make(map[string]Session),
			captchas: make(map[string]Captcha),
			config: AuthConfig{
				Username: "admin",
				Password: "admin123",
			},
		}
		authManager.Load()
	})
	return authManager
}

// Load 加载配置
func (m *AuthManager) Load() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	data, err := os.ReadFile(m.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return m.saveWithoutLock()
		}
		return err
	}

	return json.Unmarshal(data, &m.config)
}

// Save 保存配置
func (m *AuthManager) Save() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.saveWithoutLock()
}

func (m *AuthManager) saveWithoutLock() error {
	data, err := json.MarshalIndent(m.config, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(m.filePath, data, 0644)
}

// GenerateCaptcha 生成验证码
func (m *AuthManager) GenerateCaptcha() (captchaID string, captchaCode string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 清理过期验证码
	now := time.Now()
	for id, c := range m.captchas {
		if now.After(c.ExpiresAt) {
			delete(m.captchas, id)
		}
	}

	// 生成4位数字验证码
	code := ""
	for i := 0; i < 4; i++ {
		n, _ := rand.Int(rand.Reader, big.NewInt(10))
		code += fmt.Sprintf("%d", n.Int64())
	}

	// 生成验证码ID
	idBytes := make([]byte, 16)
	rand.Read(idBytes)
	id := hex.EncodeToString(idBytes)

	m.captchas[id] = Captcha{
		Code:      code,
		ExpiresAt: time.Now().Add(5 * time.Minute),
	}

	return id, code
}

// ValidateCaptcha 验证验证码
func (m *AuthManager) ValidateCaptcha(captchaID, captchaCode string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	captcha, exists := m.captchas[captchaID]
	if !exists {
		return false
	}

	// 删除已使用的验证码
	delete(m.captchas, captchaID)

	if time.Now().After(captcha.ExpiresAt) {
		return false
	}

	return captcha.Code == captchaCode
}

// Login 登录验证
func (m *AuthManager) Login(username, password string) (string, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if username == m.config.Username && password == m.config.Password {
		token := generateToken()
		m.sessions[token] = Session{
			Token:     token,
			ExpiresAt: time.Now().Add(24 * time.Hour),
		}
		return token, true
	}
	return "", false
}

// ValidateToken 验证Token
func (m *AuthManager) ValidateToken(token string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	session, exists := m.sessions[token]
	if !exists {
		return false
	}

	if time.Now().After(session.ExpiresAt) {
		delete(m.sessions, token)
		return false
	}

	return true
}

// Logout 登出
func (m *AuthManager) Logout(token string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.sessions, token)
}

// ChangePassword 修改密码
func (m *AuthManager) ChangePassword(oldPassword, newPassword string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.config.Password == oldPassword {
		m.config.Password = newPassword
		m.saveWithoutLock()
		return true
	}
	return false
}

// generateToken 生成随机Token
func generateToken() string {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// AuthMiddleware 认证中间件
func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("auth_token")
		if err != nil {
			http.Error(w, `{"code":-1,"message":"未授权"}`, http.StatusUnauthorized)
			return
		}

		if !GetAuthManager().ValidateToken(cookie.Value) {
			http.Error(w, `{"code":-1,"message":"未授权"}`, http.StatusUnauthorized)
			return
		}

		next(w, r)
	}
}
