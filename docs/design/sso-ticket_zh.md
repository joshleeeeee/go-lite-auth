# SSO Ticket 机制设计文档

## 1. 概述

本文档描述了 go-lite-auth 项目中 CAS (Central Authentication Service) 风格 SSO Ticket 机制的设计与实现。

### 1.1 设计目标

- 实现一次登录、多处访问的单点登录体验
- 使用一次性 Service Ticket (ST) 确保安全性
- 提供简洁的 API 供客户端应用集成

---

## 2. 核心概念

### 2.1 票据类型

| 票据 | 全称 | 用途 | 有效期 | 使用次数 |
|------|------|------|--------|---------|
| **ST** | Service Ticket | 单次访问某服务的凭证 | 60 秒 | **一次性** |

> **注**：本实现暂不包含 TGT (Ticket Granting Ticket)，每次 SSO 登录需要重新输入凭据。后续版本可扩展 TGT 支持。

### 2.2 参与角色

```
┌─────────────┐     ┌─────────────────┐     ┌─────────────────┐
│   用户      │     │   客户端应用     │     │   SSO 认证中心   │
│  (Browser)  │     │   (App A/B/C)   │     │  (lite-auth)    │
└─────────────┘     └─────────────────┘     └─────────────────┘
```

---

## 3. 认证流程

### 3.1 时序图

```
用户                 客户端应用 (App)              SSO 认证中心
  │                        │                           │
  │── 1. 访问受保护资源 ──>│                           │
  │                        │                           │
  │<── 2. 重定向到 SSO ────│                           │
  │     ?service=AppURL    │                           │
  │                        │                           │
  │───────── 3. 登录请求 ─────────────────────────────>│
  │                        │                           │
  │<──────── 4. 返回 ST + RedirectURL ─────────────────│
  │                        │                           │
  │── 5. 携带 ?ticket=ST ─>│                           │
  │                        │                           │
  │                        │── 6. 验证 ST ────────────>│
  │                        │                           │
  │                        │<─ 7. 返回用户信息 ────────│
  │                        │                           │
  │<── 8. 创建本地会话 ────│                           │
  │      返回受保护资源     │                           │
```

### 3.2 步骤说明

1. **用户访问应用**：用户尝试访问需要认证的资源
2. **重定向到 SSO**：应用检测到未登录，重定向到 SSO 登录页
3. **用户登录**：用户在 SSO 中心输入凭据
4. **生成 ST**：SSO 验证成功后，生成一次性 Service Ticket
5. **携带 ST 返回**：用户带着 ST 被重定向回应用
6. **验证 ST**：应用后端调用 SSO 验证接口
7. **返回用户信息**：SSO 返回用户数据，同时销毁 ST
8. **建立会话**：应用创建本地会话，用户获得访问权限

---

## 4. API 设计

### 4.1 接口列表

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/sso/login?service=xxx` | 登录入口（检查登录状态） |
| POST | `/sso/login` | 提交登录，生成 ST |
| GET | `/sso/validate?ticket=xxx&service=xxx` | 验证 ST |
| GET | `/sso/logout` | 登出 |

### 4.2 登录接口

**POST /sso/login**

请求：
```json
{
  "username": "user",
  "password": "pass",
  "service": "https://app.example.com/callback"
}
```

响应：
```json
{
  "code": 0,
  "message": "Login successful",
  "data": {
    "ticket": "ST-1737215000-a1b2c3d4e5f6...",
    "redirect_url": "https://app.example.com/callback?ticket=ST-..."
  }
}
```

### 4.3 验证接口

**GET /sso/validate?ticket=xxx&service=xxx**

响应（成功）：
```json
{
  "code": 0,
  "message": "Ticket validated successfully",
  "data": {
    "user_id": 1,
    "username": "user",
    "email": "user@example.com",
    "nickname": "User"
  }
}
```

响应（失败）：
```json
{
  "code": 401,
  "message": "Ticket not found or expired"
}
```

---

## 5. 数据存储

### 5.1 Redis 设计

**Key 格式**：`ticket:{ST-xxx}`

**Value 结构**：
```json
{
  "user_id": 1,
  "username": "user",
  "service": "https://app.example.com/callback"
}
```

**TTL**：60 秒

### 5.2 一次性使用

使用 Redis `GETDEL` 命令实现原子性的"获取并删除"操作，确保 Ticket 只能使用一次。

```go
// 原子操作：获取并删除
result, err := RDB.GetDel(ctx, PrefixTicket+ticketID).Result()
```

---

## 6. 安全考虑

### 6.1 已实现的安全措施

| 措施 | 说明 |
|------|------|
| **一次性使用** | ST 验证后立即删除 |
| **短有效期** | ST 仅 60 秒有效 |
| **Service 校验** | 验证时必须匹配原始 service URL |
| **登录限流** | 防止暴力破解（5 次失败后锁定 5 分钟） |

### 6.2 建议的增强措施

- [ ] 使用 HTTPS 传输
- [ ] ST 与客户端 IP 绑定
- [ ] 增加 TGT 支持（减少重复登录）
- [ ] 实现单点登出 (SLO)

---

## 7. 代码结构

```
internal/
├── database/
│   └── redis.go          # TicketData 结构体、SetTicketWithService、GetAndDeleteTicketData
├── service/
│   └── sso_service.go    # SSOService、GenerateServiceTicket、ValidateServiceTicket
├── handler/
│   └── sso_handler.go    # HTTP 接口处理
└── router/
    └── router.go         # /sso/* 路由注册
```

---

## 8. 客户端集成指南

### 8.1 集成步骤

1. **配置回调 URL**：在应用中配置 SSO 回调地址
2. **重定向到 SSO**：未登录用户重定向到 `{SSO_URL}/sso/login?service={YOUR_CALLBACK}`
3. **处理回调**：在回调接口中提取 `ticket` 参数
4. **验证 Ticket**：调用 `GET {SSO_URL}/sso/validate?ticket=xxx&service=xxx`
5. **建立会话**：使用返回的用户信息创建本地会话

### 8.2 示例代码 (Go)

```go
func handleSSOCallback(w http.ResponseWriter, r *http.Request) {
    ticket := r.URL.Query().Get("ticket")
    service := "https://myapp.com/auth/callback"
    
    // 验证 Ticket
    resp, err := http.Get(fmt.Sprintf(
        "%s/sso/validate?ticket=%s&service=%s",
        ssoBaseURL, ticket, url.QueryEscape(service),
    ))
    if err != nil {
        http.Error(w, "SSO validation failed", 500)
        return
    }
    
    // 解析用户信息并创建会话
    var result struct {
        Code int `json:"code"`
        Data struct {
            UserID   uint   `json:"user_id"`
            Username string `json:"username"`
        } `json:"data"`
    }
    json.NewDecoder(resp.Body).Decode(&result)
    
    if result.Code == 0 {
        // 创建本地 session
        createSession(w, result.Data.UserID, result.Data.Username)
    }
}
```

---

## 9. 版本历史

| 版本 | 日期 | 变更 |
|------|------|------|
| 1.0 | 2026-01-18 | 初始实现 |
