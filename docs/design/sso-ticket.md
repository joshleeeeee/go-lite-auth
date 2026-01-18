# SSO Ticket Mechanism Design Document

## 1. Overview

This document describes the design and implementation of the CAS (Central Authentication Service) style SSO Ticket mechanism in the go-lite-auth project.

### 1.1 Design Goals

- Enable single sign-on experience: log in once, access everywhere
- Use one-time Service Tickets (ST) for security
- Provide simple APIs for client application integration

---

## 2. Core Concepts

### 2.1 Ticket Types

| Ticket | Full Name | Purpose | TTL | Usage |
|--------|-----------|---------|-----|-------|
| **ST** | Service Ticket | One-time credential for accessing a service | 60 seconds | **One-time** |

> **Note**: This implementation does not include TGT (Ticket Granting Ticket). Users must re-enter credentials for each SSO login. TGT support can be added in future versions.

### 2.2 Participants

```
┌─────────────┐     ┌─────────────────┐     ┌─────────────────┐
│    User     │     │  Client App     │     │   SSO Server    │
│  (Browser)  │     │  (App A/B/C)    │     │  (lite-auth)    │
└─────────────┘     └─────────────────┘     └─────────────────┘
```

---

## 3. Authentication Flow

### 3.1 Sequence Diagram

```
User                   Client App                    SSO Server
  │                        │                             │
  │── 1. Access resource ─>│                             │
  │                        │                             │
  │<── 2. Redirect to SSO ─│                             │
  │     ?service=AppURL    │                             │
  │                        │                             │
  │──────── 3. Login request ───────────────────────────>│
  │                        │                             │
  │<─────── 4. Return ST + RedirectURL ──────────────────│
  │                        │                             │
  │── 5. With ?ticket=ST ─>│                             │
  │                        │                             │
  │                        │── 6. Validate ST ──────────>│
  │                        │                             │
  │                        │<─ 7. Return user info ──────│
  │                        │                             │
  │<── 8. Create session ──│                             │
  │      Return resource   │                             │
```

### 3.2 Step Description

1. **User accesses app**: User tries to access a protected resource
2. **Redirect to SSO**: App detects unauthenticated user, redirects to SSO login
3. **User logs in**: User enters credentials at SSO server
4. **Generate ST**: SSO validates credentials and generates a one-time Service Ticket
5. **Return with ST**: User is redirected back to app with the ST
6. **Validate ST**: App backend calls SSO validation endpoint
7. **Return user info**: SSO returns user data and destroys the ST
8. **Create session**: App creates local session, user gains access

---

## 4. API Design

### 4.1 Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/sso/login?service=xxx` | Login entry (check login status) |
| POST | `/sso/login` | Submit login, generate ST |
| GET | `/sso/validate?ticket=xxx&service=xxx` | Validate ST |
| GET | `/sso/logout` | Logout |

### 4.2 Login Endpoint

**POST /sso/login**

Request:
```json
{
  "username": "user",
  "password": "pass",
  "service": "https://app.example.com/callback"
}
```

Response:
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

### 4.3 Validation Endpoint

**GET /sso/validate?ticket=xxx&service=xxx**

Response (Success):
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

Response (Failure):
```json
{
  "code": 401,
  "message": "Ticket not found or expired"
}
```

---

## 5. Data Storage

### 5.1 Redis Design

**Key Format**: `ticket:{ST-xxx}`

**Value Structure**:
```json
{
  "user_id": 1,
  "username": "user",
  "service": "https://app.example.com/callback"
}
```

**TTL**: 60 seconds

### 5.2 One-Time Use

Uses Redis `GETDEL` command for atomic "get and delete" operation, ensuring tickets can only be used once.

```go
// Atomic operation: get and delete
result, err := RDB.GetDel(ctx, PrefixTicket+ticketID).Result()
```

---

## 6. Security Considerations

### 6.1 Implemented Security Measures

| Measure | Description |
|---------|-------------|
| **One-time use** | ST is deleted immediately after validation |
| **Short TTL** | ST expires after 60 seconds |
| **Service validation** | Service URL must match during validation |
| **Login rate limiting** | Prevents brute force (locked for 5 min after 5 failures) |

### 6.2 Recommended Enhancements

- [ ] Use HTTPS for transport
- [ ] Bind ST to client IP
- [ ] Add TGT support (reduce repeated logins)
- [ ] Implement Single Logout (SLO)

---

## 7. Code Structure

```
internal/
├── database/
│   └── redis.go          # TicketData struct, SetTicketWithService, GetAndDeleteTicketData
├── service/
│   └── sso_service.go    # SSOService, GenerateServiceTicket, ValidateServiceTicket
├── handler/
│   └── sso_handler.go    # HTTP handlers
└── router/
    └── router.go         # /sso/* route registration
```

---

## 8. Client Integration Guide

### 8.1 Integration Steps

1. **Configure callback URL**: Set up SSO callback address in your app
2. **Redirect to SSO**: Redirect unauthenticated users to `{SSO_URL}/sso/login?service={YOUR_CALLBACK}`
3. **Handle callback**: Extract `ticket` parameter in callback handler
4. **Validate Ticket**: Call `GET {SSO_URL}/sso/validate?ticket=xxx&service=xxx`
5. **Create session**: Use returned user info to create local session

### 8.2 Example Code (Go)

```go
func handleSSOCallback(w http.ResponseWriter, r *http.Request) {
    ticket := r.URL.Query().Get("ticket")
    service := "https://myapp.com/auth/callback"
    
    // Validate Ticket
    resp, err := http.Get(fmt.Sprintf(
        "%s/sso/validate?ticket=%s&service=%s",
        ssoBaseURL, ticket, url.QueryEscape(service),
    ))
    if err != nil {
        http.Error(w, "SSO validation failed", 500)
        return
    }
    
    // Parse user info and create session
    var result struct {
        Code int `json:"code"`
        Data struct {
            UserID   uint   `json:"user_id"`
            Username string `json:"username"`
        } `json:"data"`
    }
    json.NewDecoder(resp.Body).Decode(&result)
    
    if result.Code == 0 {
        // Create local session
        createSession(w, result.Data.UserID, result.Data.Username)
    }
}
```

---

## 9. Version History

| Version | Date | Changes |
|---------|------|---------|
| 1.0 | 2026-01-18 | Initial implementation |
