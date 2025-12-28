package main

import (
	"embed"
	"go_frame/auth"
	"go_frame/config"
	"go_frame/handlers"
	"go_frame/proxy"
	"io/fs"
	"log"
	"net/http"
	"strings"
)

//go:embed static/*
var staticFiles embed.FS

func main() {
	// 初始化配置管理器
	configManager := config.GetManager()

	// 初始化认证管理器
	_ = auth.GetAuthManager()

	// 创建API处理器
	apiHandler := handlers.NewAPIHandler(configManager)

	// 创建代理管理器
	proxyManager := proxy.NewProxyManager(configManager)

	// 创建路由
	mux := http.NewServeMux()

	// CORS中间件
	corsMiddleware := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusOK)
				return
			}
			next(w, r)
		}
	}

	// 认证相关路由（无需登录）
	mux.HandleFunc("/api/captcha", corsMiddleware(apiHandler.GetCaptcha))
	mux.HandleFunc("/api/login", corsMiddleware(apiHandler.Login))
	mux.HandleFunc("/api/logout", corsMiddleware(apiHandler.Logout))
	mux.HandleFunc("/api/check-auth", corsMiddleware(apiHandler.CheckAuth))

	// 需要认证的API路由
	mux.HandleFunc("/api/rules", corsMiddleware(auth.AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			apiHandler.ListRules(w, r)
		case http.MethodPost:
			apiHandler.CreateRule(w, r)
		case http.MethodPut:
			apiHandler.UpdateRule(w, r)
		case http.MethodDelete:
			apiHandler.DeleteRule(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})))

	mux.HandleFunc("/api/rules/toggle", corsMiddleware(auth.AuthMiddleware(apiHandler.ToggleRule)))
	mux.HandleFunc("/api/change-password", corsMiddleware(auth.AuthMiddleware(apiHandler.ChangePassword)))

	// 管理面板路由
	mux.HandleFunc("/admin", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/admin/", http.StatusMovedPermanently)
	})

	// 静态文件服务
	staticFS, _ := fs.Sub(staticFiles, "static")
	fileServer := http.FileServer(http.FS(staticFS))
	mux.Handle("/admin/", http.StripPrefix("/admin/", fileServer))

	// 代理路由（处理所有其他请求）
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		// 检查是否匹配任何代理规则
		rules := configManager.GetEnabledRules()
		for _, rule := range rules {
			prefix := rule.Path
			if !strings.HasPrefix(prefix, "/") {
				prefix = "/" + prefix
			}

			if strings.HasPrefix(path, prefix+"/") || path == prefix {
				proxyManager.ServeHTTP(w, r)
				return
			}
		}

		// 如果是根路径，显示欢迎页面（不显示管理按钮）
		if path == "/" {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Write([]byte(`<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>反向代理服务</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'SF Pro Display', sans-serif;
            min-height: 100vh;
            background: #F2F2F7;
            display: flex;
            justify-content: center;
            align-items: center;
        }
        .container {
            text-align: center;
            padding: 60px;
            background: rgba(255,255,255,0.7);
            backdrop-filter: blur(40px);
            border-radius: 40px;
            border: 1px solid rgba(255,255,255,0.6);
            box-shadow: 0 24px 48px -12px rgba(0,0,0,0.08);
        }
        h1 {
            font-size: 2.5rem;
            font-weight: 700;
            color: #1C1C1E;
            margin-bottom: 16px;
        }
        p {
            color: #8E8E93;
            font-size: 1.1rem;
        }
    </style>
</head>
<body>
    <div class="container">
        <h1>反向代理服务</h1>
        <p>服务运行中</p>
    </div>
</body>
</html>`))
			return
		}

		http.NotFound(w, r)
	})

	// 启动服务器
	addr := ":8080"
	log.Printf("========================================")
	log.Printf("  反向代理服务已启动")
	log.Printf("  服务地址: http://localhost%s", addr)
	log.Printf("  管理面板: http://localhost%s/admin/", addr)
	log.Printf("  默认账号: admin / admin123")
	log.Printf("========================================")

	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal("服务器启动失败:", err)
	}
}
