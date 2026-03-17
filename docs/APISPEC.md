# REST API Specification

Base path: `/api/v1`

## Authentication

### Security

**Access token:** Authorization: Bearer <access_token>  
**Refresh token:** Stored in cookie  

### POST /auth/register
**Request**
```json
{
  "username": "alice",
  "email": "alice@example.com",
  "password": "Password1"
}
```

**Response 201**
```json
{
  "id": 42,
  "username": "alice"
}
```

### POST /auth/login
**Request**
```json
{
  "username": "alice",
  "password": "Password1"
}
```

**Response 200**
```json
{
  "access_token": "eyJ...",
  "expires_in": 300,
  "user": {
    "id": 42,
    "username": "alice"
  }
}

```
**Headers**
```
Set-Cookie: refresh_token=...; HttpOnly; Secure; SameSite=Strict; Path=/auth; Max-Age=3600
Set-Cookie: csrf_token=...; Secure; SameSite=Strict; Path=/auth; Max-Age=3600
```

### POST /auth/refresh
Exchange refresh token cookie for a new access token using CSRF protection.

**Request**
```
Cookie: refresh_token=...; csrf_token=C  
X-CSRF-Token: C  
```

**Response 200**
```json
{
  "access_token": "eyJ...",
  "expires_in": 300
}
```

**Headers**
```
Set-Cookie: refresh_token=...; HttpOnly; Secure; SameSite=Strict; Path=/auth; Max-Age=3600
Set-Cookie: csrf_token=...; Secure; SameSite=Strict; Path=/auth; Max-Age=3600
```

### POST /auth/logout
Invalidates the refresh token.

**Request**
```
Cookie: refresh_token=...; csrf_token=C
X-CSRF-Token: C
```

**Response 204** No content.

### GET /auth/sessions
List active sessions (refresh tokens).

**Response 200**
```json
[
  {
    "id": "session_1",
    "created_at": "2026-03-17T10:00:00Z",
    "last_used_at": "2026-03-17T10:05:00Z",
    "user_agent": "Chrome",
    "ip": "192.168.1.1",
    "current": true
  }
]
```

### DELETE /auth/sessions/{id}
Revoke specific session.

--
### Token Details
**Access Token (JWT)**
```
{
  "sub": "42",
  "iss": "auth-service",
  "aud": "api",
  "iat": 1710000000,
  "exp": 1710000300,
  "jti": "f47ac10b-58cc-4372-a567-0e02b2c3d479"
}
```

**Refresh Token**
- Opaque, random string.
- Stored hashed.

**CSRF Token**
base64url(HMAC-SHA256(server_secret, refresh_token_id))
