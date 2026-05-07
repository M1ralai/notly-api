# Auth API

Base URL: `/auth`

## Endpoints

### POST /auth/login
User login
- Auth: Not required
- Body: `LoginRequest`
- Returns: JWT token + user info

### POST /auth/register
User registration
- Auth: Not required
- Body: `RegisterRequest`
- Returns: JWT token + user info

### POST /auth/verify
Verify email with code
- Auth: Not required
- Body: `VerifyEmailRequest`
- Returns: JWT token + refresh token + user info

### POST /auth/refresh
Refresh access token
- Auth: Not required
- Body: `RefreshTokenRequest`
- Returns: New JWT token + new refresh token

### POST /auth/logout
Logout user (invalidate refresh token)
- Auth: Not required (refresh token handles validation)
- Body: `LogoutRequest`
- Returns: Success message

For complete API documentation, see `/api/openapi.yaml`
