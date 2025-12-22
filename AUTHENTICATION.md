# User Authentication API

This document describes the user authentication endpoints available in the Metronom metrics collection service.

## Overview

The authentication system provides user registration and login functionality with token-based authentication. User passwords are securely hashed using bcrypt, and access tokens are stored in the database with expiration times.

## Prerequisites

- PostgreSQL database connection must be configured
- Authentication is only available when running with a database backend (not in-memory mode)

## Endpoints

### 1. Register a New User

**Endpoint:** `POST /register`

**Request Body:**
```json
{
  "username": "myusername",
  "password": "mypassword"
}
```

**Validation Rules:**
- Username must be at least 3 characters long
- Password must be at least 6 characters long
- Username must be unique

**Success Response (201 Created):**
```json
{
  "message": "User created successfully",
  "username": "myusername"
}
```

**Error Responses:**
- `400 Bad Request` - Invalid input (missing fields, too short, etc.)
- `409 Conflict` - Username already exists
- `500 Internal Server Error` - Server error

### 2. Login

**Endpoint:** `POST /login`

**Request Body:**
```json
{
  "username": "myusername",
  "password": "mypassword"
}
```

**Success Response (200 OK):**
```json
{
  "token": "base64-encoded-token-string",
  "username": "myusername"
}
```

The token is valid for 24 hours from the time of login.

**Error Responses:**
- `400 Bad Request` - Missing username or password
- `401 Unauthorized` - Invalid username or password
- `500 Internal Server Error` - Server error

### 3. Using Authentication

To access protected endpoints, include the token in the Authorization header:

```
Authorization: Bearer <token>
```

Example with curl:
```bash
curl -H "Authorization: Bearer YOUR_TOKEN_HERE" http://localhost:8080/protected-endpoint
```

## Example Usage

### Register a new user
```bash
curl -X POST http://localhost:8080/register \
  -H "Content-Type: application/json" \
  -d '{"username": "testuser", "password": "testpass123"}'
```

### Login
```bash
curl -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{"username": "testuser", "password": "testpass123"}'
```

Response:
```json
{
  "token": "AbCdEf123456...",
  "username": "testuser"
}
```

### Access protected resource
```bash
curl -H "Authorization: Bearer AbCdEf123456..." \
  http://localhost:8080/some-protected-endpoint
```

## Database Schema

The authentication system creates two tables:

### users
- `id` (SERIAL PRIMARY KEY)
- `username` (VARCHAR(50) UNIQUE NOT NULL)
- `password_hash` (VARCHAR(100) NOT NULL)
- `created_at` (TIMESTAMP DEFAULT CURRENT_TIMESTAMP)

### tokens
- `id` (SERIAL PRIMARY KEY)
- `user_id` (INTEGER REFERENCES users(id))
- `token` (VARCHAR(100) UNIQUE NOT NULL)
- `created_at` (TIMESTAMP DEFAULT CURRENT_TIMESTAMP)
- `expires_at` (TIMESTAMP NOT NULL)

## Security Considerations

1. **Password Storage**: Passwords are hashed using bcrypt with default cost before storage
2. **Token Expiration**: Tokens automatically expire after 24 hours
3. **HTTPS**: In production, always use HTTPS to protect tokens in transit
4. **Token Storage**: Store tokens securely on the client side (e.g., secure HTTP-only cookies or secure storage)

## Implementation Details

- Password hashing: bcrypt with default cost (10)
- Token generation: 32 random bytes encoded as base64 URL-safe string
- Token lifetime: 24 hours
- Authentication middleware: Available for protecting routes

## Testing

Run the authentication tests:
```bash
go test ./internal/auth/...
```

Note: Tests require a PostgreSQL database connection configured at:
`postgres://postgres:passwordas@localhost:5432/forgo`
