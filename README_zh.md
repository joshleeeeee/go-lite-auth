# go-lite-auth

[![Go CI](https://github.com/joshleeeeee/go-lite-auth/actions/workflows/go.yml/badge.svg)](https://github.com/joshleeeeee/go-lite-auth/actions/workflows/go.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/github/go-mod/go-version/joshleeeeee/go-lite-auth)](https://golang.org/)

轻量级 SSO (单点登录) 系统，使用 Go 语言构建。

[English README](README.md) | [贡献指南](CONTRIBUTING.md) | [更新日志](CHANGELOG.md)

## 技术栈

- **Web 框架**: Gin
- **数据库**: MySQL 8 + GORM  
- **缓存**: Redis
- **认证**: JWT (Access Token + Refresh Token)
- **配置**: Viper

## 项目结构

```
lite-auth/
├── cmd/
│   └── server/
│       └── main.go           # 程序入口
├── config/
│   └── config.yaml           # 配置文件
├── internal/
│   ├── config/               # 配置加载
│   ├── database/             # MySQL & Redis 初始化
│   ├── handler/              # HTTP 处理器
│   ├── middleware/           # 中间件 (JWT验证, CORS)
│   ├── model/                # 数据模型
│   ├── repository/           # 数据访问层
│   ├── router/               # 路由配置
│   └── service/              # 业务逻辑层
├── pkg/
│   └── jwt/                  # JWT 工具包
├── go.mod
└── README.md
```

## 快速开始

### 1. 环境准备

确保已安装:
- Go 1.21+
- MySQL 8.0+
- Redis 6.0+

### 2. 创建数据库

```sql
CREATE DATABASE lite_auth CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
```

### 3. 修改配置

编辑 `config/config.yaml`，配置你的 MySQL 和 Redis 连接信息。

### 4. 安装依赖

```bash
go mod tidy
```

### 5. 启动服务

```bash
go run cmd/server/main.go
```

服务将在 http://localhost:8080 启动。

## API 接口

### 认证相关

| 方法 | 路径 | 说明 | 认证 |
|------|------|------|------|
| POST | `/api/auth/register` | 用户注册 | ❌ |
| POST | `/api/auth/login` | 用户登录 | ❌ |
| POST | `/api/auth/logout` | 用户登出 | ✅ |
| POST | `/api/auth/refresh` | 刷新令牌 | ❌ |
| GET | `/api/auth/validate` | 验证令牌 | ❌ |

### 用户相关

| 方法 | 路径 | 说明 | 认证 |
|------|------|------|------|
| GET | `/api/user/info` | 获取当前用户信息 | ✅ |

### 请求示例

**注册**
```bash
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"test","email":"test@example.com","password":"123456"}'
```

**登录**
```bash
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"test","password":"123456"}'
```

**获取用户信息**
```bash
curl http://localhost:8080/api/user/info \
  -H "Authorization: Bearer <access_token>"
```

## Redis 键设计

| 前缀 | 用途 | 过期时间 |
|------|------|----------|
| `session:` | 用户会话 | 24小时 |
| `blacklist:` | Token 黑名单 | Token剩余有效期 |
| `ticket:` | SSO Ticket | 60秒 |
| `login_fail:` | 登录失败计数 | 5分钟 |

## 后续扩展

- [ ] SSO Ticket 机制 (CAS 风格)
- [ ] OAuth 2.0 授权码模式
- [ ] 前端登录页面
- [ ] 客户端应用管理
- [ ] 用户管理后台

## 参与贡献

欢迎任何形式的贡献！请阅读 [CONTRIBUTING.md](CONTRIBUTING.md) 了解详细指南。

## 开源协议

MIT
