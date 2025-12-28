package proxy

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"go_proxy_every/config"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"regexp"
	"strings"
)

// ProxyManager 代理管理器
type ProxyManager struct {
	configManager *config.ConfigManager
}

// NewProxyManager 创建代理管理器
func NewProxyManager(cm *config.ConfigManager) *ProxyManager {
	return &ProxyManager{
		configManager: cm,
	}
}

// ServeHTTP 处理代理请求
func (pm *ProxyManager) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	// 查找匹配的规则
	rules := pm.configManager.GetEnabledRules()
	for _, rule := range rules {
		prefix := rule.Path
		if !strings.HasPrefix(prefix, "/") {
			prefix = "/" + prefix
		}

		if strings.HasPrefix(path, prefix+"/") || path == prefix {
			pm.handleProxy(w, r, rule, prefix)
			return
		}
	}

	// 没有匹配的规则
	http.Error(w, "No proxy rule matched", http.StatusNotFound)
}

// handleProxy 处理具体的代理请求
func (pm *ProxyManager) handleProxy(w http.ResponseWriter, r *http.Request, rule config.ProxyRule, prefix string) {
	targetURL, err := url.Parse(rule.Target)
	if err != nil {
		http.Error(w, "Invalid target URL", http.StatusInternalServerError)
		return
	}

	proxy := &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			// 移除前缀路径
			originalPath := req.URL.Path
			newPath := strings.TrimPrefix(originalPath, prefix)
			if newPath == "" {
				newPath = "/"
			}

			req.URL.Scheme = targetURL.Scheme
			req.URL.Host = targetURL.Host
			req.URL.Path = singleJoiningSlash(targetURL.Path, newPath)

			// 设置Host头
			req.Host = targetURL.Host

			// 保留查询参数
			if r.URL.RawQuery != "" {
				req.URL.RawQuery = r.URL.RawQuery
			}

			// 设置必要的头
			req.Header.Set("X-Forwarded-Host", r.Host)
			req.Header.Set("X-Forwarded-Proto", "http")
			req.Header.Set("X-Real-IP", getClientIP(r))

			// 删除可能导致问题的头
			req.Header.Del("Accept-Encoding") // 禁用压缩以便修改响应

			log.Printf("[Proxy] %s -> %s%s", originalPath, targetURL.Host, req.URL.Path)
		},
		ModifyResponse: func(resp *http.Response) error {
			// 修改响应中的链接
			return pm.modifyResponse(resp, rule, prefix)
		},
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			log.Printf("[Proxy Error] %v", err)
			http.Error(w, fmt.Sprintf("Proxy error: %v", err), http.StatusBadGateway)
		},
	}

	proxy.ServeHTTP(w, r)
}

// modifyResponse 修改响应内容
func (pm *ProxyManager) modifyResponse(resp *http.Response, rule config.ProxyRule, prefix string) error {
	contentType := resp.Header.Get("Content-Type")

	// 只处理HTML内容
	if !strings.Contains(contentType, "text/html") {
		return nil
	}

	// 读取响应体
	var reader io.ReadCloser
	var err error

	switch resp.Header.Get("Content-Encoding") {
	case "gzip":
		reader, err = gzip.NewReader(resp.Body)
		if err != nil {
			return err
		}
		defer reader.Close()
	default:
		reader = resp.Body
	}

	body, err := io.ReadAll(reader)
	if err != nil {
		return err
	}
	resp.Body.Close()

	// 重写HTML中的链接
	bodyStr := string(body)
	bodyStr = pm.rewriteHTML(bodyStr, rule, prefix)

	// 更新响应
	resp.Body = io.NopCloser(bytes.NewReader([]byte(bodyStr)))
	resp.ContentLength = int64(len(bodyStr))
	resp.Header.Set("Content-Length", fmt.Sprintf("%d", len(bodyStr)))
	resp.Header.Del("Content-Encoding") // 移除压缩标记

	return nil
}

// rewriteHTML 重写HTML中的链接
func (pm *ProxyManager) rewriteHTML(html string, rule config.ProxyRule, prefix string) string {
	targetURL, err := url.Parse(rule.Target)
	if err != nil {
		return html
	}

	// 替换绝对URL
	// 例如：https://www.nsmao.com/xxx -> /nsmao/xxx
	absolutePattern := regexp.MustCompile(fmt.Sprintf(`(href|src|action)=["']%s(/[^"']*)?["']`, regexp.QuoteMeta(rule.Target)))
	html = absolutePattern.ReplaceAllStringFunc(html, func(match string) string {
		return strings.Replace(match, rule.Target, prefix, 1)
	})

	// 替换协议相对URL
	// 例如：//www.nsmao.com/xxx -> /nsmao/xxx
	protoRelativePattern := regexp.MustCompile(fmt.Sprintf(`(href|src|action)=["']//%s(/[^"']*)?["']`, regexp.QuoteMeta(targetURL.Host)))
	html = protoRelativePattern.ReplaceAllStringFunc(html, func(match string) string {
		return strings.Replace(match, "//"+targetURL.Host, prefix, 1)
	})

	return html
}

// singleJoiningSlash 连接路径
func singleJoiningSlash(a, b string) string {
	aslash := strings.HasSuffix(a, "/")
	bslash := strings.HasPrefix(b, "/")
	switch {
	case aslash && bslash:
		return a + b[1:]
	case !aslash && !bslash:
		return a + "/" + b
	}
	return a + b
}

// getClientIP 获取客户端IP
func getClientIP(r *http.Request) string {
	// 尝试从X-Forwarded-For获取
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// 尝试从X-Real-IP获取
	xri := r.Header.Get("X-Real-IP")
	if xri != "" {
		return xri
	}

	// 使用RemoteAddr
	return strings.Split(r.RemoteAddr, ":")[0]
}
