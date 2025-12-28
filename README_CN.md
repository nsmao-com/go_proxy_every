# Go 反向代理

一个轻量级、易于使用的反向代理服务器，配备精美的 Web 管理面板。

![Go Version](https://img.shields.io/badge/Go-1.23+-00ADD8?style=flat&logo=go)
![Docker](https://img.shields.io/badge/Docker-Ready-2496ED?style=flat&logo=docker)
![License](https://img.shields.io/badge/License-MIT-green?style=flat)

**[English](README.md)** | **官方网站: [www.nsmao.com](https://www.nsmao.com)**

## 功能特性

- **Web 管理面板** - 精美的 Zen-iOS Hybrid 设计，玻璃态效果
- **动态配置** - 无需重启即可添加/编辑/删除代理规则
- **身份验证** - 安全的管理面板登录系统
- **JSON 存储** - 无需数据库，配置存储在 JSON 文件中
- **Docker 支持** - 支持 Docker 和 docker-compose 轻松部署
- **多平台支持** - 支持 linux/amd64 和 linux/arm64

## 快速开始

### 使用 Docker（推荐）

```bash
# 从 GitHub Container Registry 拉取
docker pull ghcr.io/nsmao-com/go_proxy_every:latest

# 运行容器
docker run -d \
  --name go-proxy-every \
  -p 8080:8080 \
  -v $(pwd)/data:/app/data \
  ghcr.io/nsmao-com/go_proxy_every:latest
```

### 使用 Docker Compose

```bash
# 克隆仓库
git clone https://github.com/nsmao-com/go_proxy_every.git
cd go_proxy_every

# 使用 docker-compose 启动
docker-compose up -d
```

### 从源码构建

```bash
# 克隆仓库
git clone https://github.com/nsmao-com/go_proxy_every.git
cd go_proxy_every

# 构建
go build -o proxy .

# 运行
./proxy
```

## 使用说明

### 访问服务

- **主页**: http://localhost:8080
- **管理面板**: http://localhost:8080/admin/

### 默认账号

| 字段 | 值 |
|------|------|
| 用户名 | `admin` |
| 密码 | `admin123` |

> **重要提示**: 首次登录后请立即修改默认密码！

### 添加代理规则

1. 登录管理面板
2. 点击"添加规则"按钮
3. 填写表单：
   - **名称**: 规则的友好名称
   - **路径**: 本地路径前缀（如 `nsmao`）
   - **目标地址**: 要代理的目标 URL（如 `https://www.nsmao.com`）
4. 启用规则并保存

### 示例

如果添加一条规则：
- 路径: `nsmao`
- 目标: `https://www.nsmao.com`

那么访问 `http://localhost:8080/nsmao` 将会代理到 `https://www.nsmao.com`

## 配置文件

配置存储在 `data/rules.json` 中：

```json
{
  "rules": [
    {
      "id": "uuid",
      "name": "示例网站",
      "path": "example",
      "target": "https://example.com",
      "enabled": true,
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T00:00:00Z"
    }
  ]
}
```

身份验证配置存储在 `data/auth.json` 中：

```json
{
  "username": "admin",
  "password": "admin123"
}
```

## API 接口

| 接口 | 方法 | 描述 | 需要认证 |
|------|------|------|----------|
| `/api/captcha` | GET | 获取验证码 | 否 |
| `/api/login` | POST | 登录 | 否 |
| `/api/logout` | POST | 退出登录 | 否 |
| `/api/check-auth` | GET | 检查登录状态 | 否 |
| `/api/rules` | GET | 获取所有规则 | 是 |
| `/api/rules` | POST | 创建规则 | 是 |
| `/api/rules` | PUT | 更新规则 | 是 |
| `/api/rules` | DELETE | 删除规则 | 是 |
| `/api/rules/toggle` | POST | 切换规则状态 | 是 |
| `/api/change-password` | POST | 修改密码 | 是 |

## 项目结构

```
go-reverse-proxy/
├── main.go              # 入口文件
├── go.mod               # Go 模块
├── Dockerfile           # Docker 构建文件
├── docker-compose.yml   # Docker Compose 配置
├── auth/
│   └── auth.go          # 身份验证逻辑
├── config/
│   └── config.go        # 配置管理
├── handlers/
│   └── api.go           # API 处理器
├── proxy/
│   └── reverse_proxy.go # 反向代理核心
├── static/
│   └── index.html       # Web 界面
├── data/
│   ├── rules.json       # 代理规则
│   └── auth.json        # 认证配置
└── .github/
    └── workflows/
        └── docker.yml   # GitHub Actions
```

## Docker 镜像

推送到 GitHub 后，Docker 镜像会自动构建并推送到 GitHub Container Registry。

```bash
# 拉取镜像
docker pull ghcr.io/nsmao-com/go_proxy_every:latest

# 或者指定版本
docker pull ghcr.io/nsmao-com/go_proxy_every:v1.0.0
```

## 环境变量

| 变量 | 默认值 | 描述 |
|------|--------|------|
| `TZ` | `Asia/Shanghai` | 时区 |

## 安全注意事项

1. 部署后立即修改默认密码
2. 生产环境请使用 HTTPS（部署在 nginx/caddy 后面并启用 SSL）
3. 考虑使用防火墙规则限制对管理面板的访问

## 开源协议

MIT License

## 贡献

欢迎提交 Pull Request。如有重大更改，请先开 Issue 讨论。

## 官方网站

更多信息请访问：[www.nsmao.com](https://www.nsmao.com)
