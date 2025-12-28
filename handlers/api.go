package handlers

import (
	"bytes"
	"encoding/json"
	"go_proxy_every/auth"
	"go_proxy_every/config"
	"image"
	"image/color"
	"image/png"
	"math/rand"
	"net/http"
	"time"

	"github.com/google/uuid"
)

// APIHandler API处理器
type APIHandler struct {
	configManager *config.ConfigManager
	authManager   *auth.AuthManager
}

// NewAPIHandler 创建API处理器
func NewAPIHandler(cm *config.ConfigManager) *APIHandler {
	return &APIHandler{
		configManager: cm,
		authManager:   auth.GetAuthManager(),
	}
}

// Response 统一响应格式
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// writeJSON 写入JSON响应
func writeJSON(w http.ResponseWriter, code int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(data)
}

// success 成功响应
func success(w http.ResponseWriter, data interface{}) {
	writeJSON(w, http.StatusOK, Response{
		Code:    0,
		Message: "success",
		Data:    data,
	})
}

// fail 失败响应
func fail(w http.ResponseWriter, code int, message string) {
	writeJSON(w, code, Response{
		Code:    -1,
		Message: message,
	})
}

// GetCaptcha 获取验证码
func (h *APIHandler) GetCaptcha(w http.ResponseWriter, r *http.Request) {
	captchaID, captchaCode := h.authManager.GenerateCaptcha()

	// 生成验证码图片
	img := generateCaptchaImage(captchaCode)

	// 将图片编码为PNG
	var buf bytes.Buffer
	png.Encode(&buf, img)

	// 返回JSON包含captchaId和base64图片
	imgBase64 := "data:image/png;base64," + encodeBase64(buf.Bytes())

	success(w, map[string]string{
		"captcha_id":  captchaID,
		"captcha_img": imgBase64,
	})
}

// generateCaptchaImage 生成验证码图片
func generateCaptchaImage(code string) image.Image {
	width, height := 120, 50
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// 背景色 - 浅灰色
	bgColor := color.RGBA{245, 245, 250, 255}
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			img.Set(x, y, bgColor)
		}
	}

	// 添加干扰点
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < 100; i++ {
		x := r.Intn(width)
		y := r.Intn(height)
		gray := uint8(r.Intn(200) + 50)
		img.Set(x, y, color.RGBA{gray, gray, gray, 255})
	}

	// 添加干扰线
	for i := 0; i < 4; i++ {
		x1 := r.Intn(width)
		y1 := r.Intn(height)
		x2 := r.Intn(width)
		y2 := r.Intn(height)
		lineColor := color.RGBA{uint8(r.Intn(150) + 100), uint8(r.Intn(150) + 100), uint8(r.Intn(150) + 100), 255}
		drawLine(img, x1, y1, x2, y2, lineColor)
	}

	// 绘制数字
	textColor := color.RGBA{40, 40, 50, 255}
	for i, c := range code {
		drawDigit(img, int(c-'0'), 15+i*25, 10, textColor)
	}

	return img
}

// drawLine 画线
func drawLine(img *image.RGBA, x1, y1, x2, y2 int, c color.Color) {
	dx := abs(x2 - x1)
	dy := abs(y2 - y1)
	sx, sy := 1, 1
	if x1 >= x2 {
		sx = -1
	}
	if y1 >= y2 {
		sy = -1
	}
	err := dx - dy

	for {
		img.Set(x1, y1, c)
		if x1 == x2 && y1 == y2 {
			break
		}
		e2 := err * 2
		if e2 > -dy {
			err -= dy
			x1 += sx
		}
		if e2 < dx {
			err += dx
			y1 += sy
		}
	}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// drawDigit 绘制数字
func drawDigit(img *image.RGBA, digit, x, y int, c color.Color) {
	// 简单的7段数字显示
	patterns := [][]string{
		{" ### ", "#   #", "#   #", "#   #", "#   #", "#   #", " ### "}, // 0
		{"  #  ", " ##  ", "  #  ", "  #  ", "  #  ", "  #  ", " ### "}, // 1
		{" ### ", "#   #", "    #", " ### ", "#    ", "#    ", "#####"}, // 2
		{" ### ", "#   #", "    #", "  ## ", "    #", "#   #", " ### "}, // 3
		{"#   #", "#   #", "#   #", "#####", "    #", "    #", "    #"}, // 4
		{"#####", "#    ", "#    ", " ### ", "    #", "#   #", " ### "}, // 5
		{" ### ", "#    ", "#    ", "#### ", "#   #", "#   #", " ### "}, // 6
		{"#####", "    #", "   # ", "  #  ", " #   ", " #   ", " #   "}, // 7
		{" ### ", "#   #", "#   #", " ### ", "#   #", "#   #", " ### "}, // 8
		{" ### ", "#   #", "#   #", " ####", "    #", "    #", " ### "}, // 9
	}

	if digit < 0 || digit > 9 {
		return
	}

	pattern := patterns[digit]
	for dy, row := range pattern {
		for dx, ch := range row {
			if ch == '#' {
				// 画粗一点的点
				for px := 0; px < 3; px++ {
					for py := 0; py < 4; py++ {
						img.Set(x+dx*3+px, y+dy*4+py, c)
					}
				}
			}
		}
	}
}

// encodeBase64 base64编码
func encodeBase64(data []byte) string {
	const base64Chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"
	result := make([]byte, 0, (len(data)+2)/3*4)

	for i := 0; i < len(data); i += 3 {
		var n uint32
		remaining := len(data) - i

		if remaining >= 3 {
			n = uint32(data[i])<<16 | uint32(data[i+1])<<8 | uint32(data[i+2])
			result = append(result, base64Chars[n>>18&0x3F], base64Chars[n>>12&0x3F], base64Chars[n>>6&0x3F], base64Chars[n&0x3F])
		} else if remaining == 2 {
			n = uint32(data[i])<<16 | uint32(data[i+1])<<8
			result = append(result, base64Chars[n>>18&0x3F], base64Chars[n>>12&0x3F], base64Chars[n>>6&0x3F], '=')
		} else {
			n = uint32(data[i]) << 16
			result = append(result, base64Chars[n>>18&0x3F], base64Chars[n>>12&0x3F], '=', '=')
		}
	}

	return string(result)
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username   string `json:"username"`
	Password   string `json:"password"`
	CaptchaID  string `json:"captcha_id"`
	CaptchaCode string `json:"captcha_code"`
}

// Login 登录
func (h *APIHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		fail(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		fail(w, http.StatusBadRequest, "请求格式错误")
		return
	}

	// 验证验证码
	if !h.authManager.ValidateCaptcha(req.CaptchaID, req.CaptchaCode) {
		fail(w, http.StatusBadRequest, "验证码错误")
		return
	}

	token, ok := h.authManager.Login(req.Username, req.Password)
	if !ok {
		fail(w, http.StatusUnauthorized, "用户名或密码错误")
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   86400,
	})

	success(w, map[string]string{"token": token})
}

// Logout 登出
func (h *APIHandler) Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("auth_token")
	if err == nil {
		h.authManager.Logout(cookie.Value)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	})

	success(w, nil)
}

// CheckAuth 检查登录状态
func (h *APIHandler) CheckAuth(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("auth_token")
	if err != nil {
		fail(w, http.StatusUnauthorized, "未登录")
		return
	}

	if !h.authManager.ValidateToken(cookie.Value) {
		fail(w, http.StatusUnauthorized, "登录已过期")
		return
	}

	success(w, map[string]bool{"authenticated": true})
}

// ChangePasswordRequest 修改密码请求
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

// ChangePassword 修改密码
func (h *APIHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		fail(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	var req ChangePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		fail(w, http.StatusBadRequest, "请求格式错误")
		return
	}

	if !h.authManager.ChangePassword(req.OldPassword, req.NewPassword) {
		fail(w, http.StatusBadRequest, "原密码错误")
		return
	}

	success(w, nil)
}

// ListRules 获取所有规则
func (h *APIHandler) ListRules(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		fail(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	rules := h.configManager.GetRules()
	success(w, rules)
}

// CreateRuleRequest 创建规则请求
type CreateRuleRequest struct {
	Name    string `json:"name"`
	Path    string `json:"path"`
	Target  string `json:"target"`
	Enabled bool   `json:"enabled"`
}

// CreateRule 创建规则
func (h *APIHandler) CreateRule(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		fail(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	var req CreateRuleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		fail(w, http.StatusBadRequest, "请求格式错误")
		return
	}

	if req.Name == "" || req.Path == "" || req.Target == "" {
		fail(w, http.StatusBadRequest, "名称、路径和目标地址不能为空")
		return
	}

	existingRules := h.configManager.GetRules()
	for _, rule := range existingRules {
		if rule.Path == req.Path {
			fail(w, http.StatusBadRequest, "路径已存在")
			return
		}
	}

	rule := config.ProxyRule{
		ID:        uuid.New().String(),
		Name:      req.Name,
		Path:      req.Path,
		Target:    req.Target,
		Enabled:   req.Enabled,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := h.configManager.AddRule(rule); err != nil {
		fail(w, http.StatusInternalServerError, "保存规则失败")
		return
	}

	success(w, rule)
}

// UpdateRuleRequest 更新规则请求
type UpdateRuleRequest struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Path    string `json:"path"`
	Target  string `json:"target"`
	Enabled bool   `json:"enabled"`
}

// UpdateRule 更新规则
func (h *APIHandler) UpdateRule(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		fail(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	var req UpdateRuleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		fail(w, http.StatusBadRequest, "请求格式错误")
		return
	}

	if req.ID == "" {
		fail(w, http.StatusBadRequest, "ID不能为空")
		return
	}

	rule := config.ProxyRule{
		ID:      req.ID,
		Name:    req.Name,
		Path:    req.Path,
		Target:  req.Target,
		Enabled: req.Enabled,
	}

	if err := h.configManager.UpdateRule(rule); err != nil {
		fail(w, http.StatusInternalServerError, "更新规则失败")
		return
	}

	success(w, rule)
}

// DeleteRuleRequest 删除规则请求
type DeleteRuleRequest struct {
	ID string `json:"id"`
}

// DeleteRule 删除规则
func (h *APIHandler) DeleteRule(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		fail(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	var req DeleteRuleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		fail(w, http.StatusBadRequest, "请求格式错误")
		return
	}

	if req.ID == "" {
		fail(w, http.StatusBadRequest, "ID不能为空")
		return
	}

	if err := h.configManager.DeleteRule(req.ID); err != nil {
		fail(w, http.StatusInternalServerError, "删除规则失败")
		return
	}

	success(w, nil)
}

// ToggleRule 切换规则状态
func (h *APIHandler) ToggleRule(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		fail(w, http.StatusMethodNotAllowed, "方法不允许")
		return
	}

	var req struct {
		ID      string `json:"id"`
		Enabled bool   `json:"enabled"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		fail(w, http.StatusBadRequest, "请求格式错误")
		return
	}

	rules := h.configManager.GetRules()
	for _, rule := range rules {
		if rule.ID == req.ID {
			rule.Enabled = req.Enabled
			if err := h.configManager.UpdateRule(rule); err != nil {
				fail(w, http.StatusInternalServerError, "切换状态失败")
				return
			}
			success(w, rule)
			return
		}
	}

	fail(w, http.StatusNotFound, "规则不存在")
}
