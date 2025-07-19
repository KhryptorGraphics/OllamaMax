# Authentication System

This directory contains a comprehensive authentication and authorization system for the Ollama Distributed System. The system provides JWT-based authentication, API key management, role-based access control (RBAC), and comprehensive security features.

## Overview

The authentication system consists of several key components:

- **Auth Manager** - Core authentication and user management
- **JWT Manager** - Advanced JWT token handling with refresh tokens
- **Middleware Manager** - HTTP middleware for API protection
- **Routes** - HTTP endpoints for authentication operations
- **Integration** - Easy integration with existing API servers

## Features

### 🔐 Authentication Methods
- **JWT Tokens** - Secure JSON Web Tokens with RSA-256 signing
- **API Keys** - Long-lived API keys for service integration
- **Refresh Tokens** - Secure token refresh mechanism
- **Service Tokens** - Special tokens for service-to-service communication

### 👥 User Management
- **User Registration** - Self-service user registration
- **User Profiles** - User metadata and profile management
- **Password Management** - Secure password hashing with bcrypt
- **Account Status** - Active/inactive user management

### 🔑 Authorization
- **Role-Based Access Control (RBAC)** - Predefined roles with permissions
- **Fine-Grained Permissions** - Granular permission system
- **Dynamic Permission Checking** - Runtime permission validation
- **Resource-Level Security** - Protect specific API endpoints

### 🛡️ Security Features
- **Token Blacklisting** - Revoke compromised tokens
- **Session Management** - Secure session tracking
- **Rate Limiting** - Prevent abuse and DoS attacks
- **Security Headers** - Comprehensive HTTP security headers
- **Audit Logging** - Complete audit trail of authentication events

## Quick Start

### 1. Basic Setup

```go
package main

import (
    "log"
    "github.com/ollama/ollama-distributed/internal/auth"
    "github.com/ollama/ollama-distributed/internal/config"
)

func main() {
    // Configure authentication
    cfg := &config.AuthConfig{
        Enabled:     true,
        Method:      "jwt",
        TokenExpiry: 24 * time.Hour,
        SecretKey:   "your-secret-key",
        Issuer:      "ollama-distributed",
        Audience:    "ollama-api",
    }
    
    // Create authentication integration
    authIntegration, err := auth.NewIntegration(cfg)
    if err != nil {
        log.Fatalf("Failed to create auth integration: %v", err)
    }
    defer authIntegration.Close()
    
    // Setup router with authentication
    router := authIntegration.SetupRouter()
    
    // Start server
    log.Println("Starting server with authentication")
    router.Run(":8080")
}
```

### 2. Protecting API Endpoints

```go
// Protect existing API routes
authIntegration.ProtectAPIRoutes(router)

// Or protect specific routes
api := router.Group("/api/v1")
api.Use(authIntegration.MiddlewareManager.AuthRequired())

// Require specific permissions
api.GET("/models", 
    authIntegration.MiddlewareManager.RequirePermission(auth.PermissionModelRead),
    getModelsHandler)

api.POST("/models/:name/download",
    authIntegration.MiddlewareManager.RequirePermission(auth.PermissionModelWrite),
    downloadModelHandler)

// Require admin role
api.DELETE("/models/:name",
    authIntegration.MiddlewareManager.RequireRole(auth.RoleAdmin),
    deleteModelHandler)
```

### 3. Creating Service Tokens

```go
// Create a service token for internal communication
serviceToken, err := authIntegration.CreateServiceToken("node-1", "Ollama Node 1")
if err != nil {
    log.Printf("Failed to create service token: %v", err)
}

// Use the token in requests
req.Header.Set("Authorization", "Bearer " + serviceToken)
```

## API Endpoints

### Authentication
- `POST /api/v1/login` - User login
- `POST /api/v1/register` - User registration
- `POST /api/v1/refresh` - Refresh access token
- `POST /api/v1/user/logout` - Logout and revoke session

### User Management
- `GET /api/v1/user/profile` - Get user profile
- `PUT /api/v1/user/profile` - Update user profile
- `POST /api/v1/user/change-password` - Change password
- `GET /api/v1/user/sessions` - List user sessions
- `DELETE /api/v1/user/sessions/:session_id` - Revoke session

### API Key Management
- `GET /api/v1/api-keys` - List user's API keys
- `POST /api/v1/api-keys` - Create new API key
- `DELETE /api/v1/api-keys/:key_id` - Revoke API key

### Admin Operations
- `GET /api/v1/admin/users` - List all users
- `POST /api/v1/admin/users` - Create user
- `GET /api/v1/admin/users/:user_id` - Get user details
- `PUT /api/v1/admin/users/:user_id` - Update user
- `DELETE /api/v1/admin/users/:user_id` - Delete user
- `POST /api/v1/admin/users/:user_id/reset-password` - Reset user password
- `GET /api/v1/admin/stats` - Authentication statistics

## Roles and Permissions

### Predefined Roles

- **admin** - Full system access
- **operator** - Node and model management
- **user** - Standard user access
- **readonly** - Read-only access
- **service** - Service-to-service communication

### Permission System

#### Node Permissions
- `node:read` - View node information
- `node:write` - Manage nodes
- `node:admin` - Full node administration

#### Model Permissions
- `model:read` - View models
- `model:write` - Download/manage models
- `model:admin` - Full model administration

#### Cluster Permissions
- `cluster:read` - View cluster status
- `cluster:write` - Manage cluster
- `cluster:admin` - Full cluster administration

#### Inference Permissions
- `inference:read` - View inference jobs
- `inference:write` - Submit inference requests

#### System Permissions
- `metrics:read` - View system metrics
- `system:admin` - Full system administration
- `user:admin` - User management

## Security Best Practices

### 1. Token Security
- Use strong secret keys (32+ characters)
- Implement token rotation
- Set appropriate expiration times
- Blacklist compromised tokens

### 2. Password Security
- Enforce strong password policies
- Use bcrypt for password hashing
- Implement password expiration
- Prevent password reuse

### 3. API Key Security
- Generate cryptographically secure keys
- Implement key rotation
- Monitor key usage
- Revoke unused keys

### 4. Network Security
- Use HTTPS in production
- Implement proper CORS policies
- Add security headers
- Enable rate limiting

### 5. Monitoring and Auditing
- Log all authentication events
- Monitor failed login attempts
- Track permission violations
- Implement alerting

## Configuration

### Environment Variables

```bash
# Authentication
OLLAMA_AUTH_ENABLED=true
OLLAMA_AUTH_METHOD=jwt
OLLAMA_AUTH_TOKEN_EXPIRY=24h
OLLAMA_AUTH_SECRET_KEY=your-secret-key
OLLAMA_AUTH_ISSUER=ollama-distributed
OLLAMA_AUTH_AUDIENCE=ollama-api

# Security
OLLAMA_SECURITY_TLS_ENABLED=true
OLLAMA_SECURITY_TLS_CERT_FILE=/path/to/cert.pem
OLLAMA_SECURITY_TLS_KEY_FILE=/path/to/key.pem
```

### Configuration File

```yaml
security:
  auth:
    enabled: true
    method: jwt
    token_expiry: 24h
    secret_key: your-secret-key
    issuer: ollama-distributed
    audience: ollama-api
  tls:
    enabled: true
    cert_file: /path/to/cert.pem
    key_file: /path/to/key.pem
```

## Advanced Features

### Custom Middleware

```go
// Create custom permission middleware
func RequireModelAccess(modelName string) gin.HandlerFunc {
    return func(c *gin.Context) {
        user := auth.GetCurrentUser(c)
        if user == nil {
            c.JSON(401, gin.H{"error": "Authentication required"})
            c.Abort()
            return
        }
        
        // Custom logic to check model access
        if !hasModelAccess(user, modelName) {
            c.JSON(403, gin.H{"error": "Access denied to model"})
            c.Abort()
            return
        }
        
        c.Next()
    }
}
```

### Short-Lived Tokens

```go
// Create token for specific operations
token, err := jwtManager.CreateShortLivedToken(user, 5*time.Minute, "model-download")
```

### Token Validation in Other Services

```go
// Get public key for token verification
publicKey := jwtManager.GetPublicKey()

// Verify token in another service
token, err := jwt.ParseWithClaims(tokenString, &auth.Claims{}, func(token *jwt.Token) (interface{}, error) {
    return publicKey, nil
})
```

## Testing

Run the test suite:

```bash
go test ./internal/auth/ -v
```

Test with coverage:

```bash
go test ./internal/auth/ -cover
```

## Migration from Existing Auth

If you have an existing authentication system, follow these steps:

1. **Analyze Current System** - Document existing users, roles, and permissions
2. **Create Migration Plan** - Map existing roles to new permission system
3. **Implement Gradual Migration** - Support both systems during transition
4. **Update API Endpoints** - Apply new authentication middleware
5. **Test Thoroughly** - Verify all endpoints work correctly
6. **Remove Old System** - Clean up deprecated authentication code

## Troubleshooting

### Common Issues

1. **Token Validation Fails**
   - Check secret key configuration
   - Verify token expiration
   - Ensure proper signing method

2. **Permission Denied Errors**
   - Verify user has required permissions
   - Check role assignments
   - Review permission constants

3. **API Key Issues**
   - Ensure proper API key format
   - Check key expiration
   - Verify key is active

4. **Session Problems**
   - Check session expiration
   - Verify session storage
   - Review cleanup routines

### Debug Mode

Enable debug logging:

```go
gin.SetMode(gin.DebugMode)
```

## Performance Considerations

1. **Token Validation** - Use RSA keys for better security
2. **Session Storage** - Consider Redis for distributed deployments
3. **Rate Limiting** - Implement distributed rate limiting
4. **Caching** - Cache permission checks for better performance
5. **Cleanup** - Regular cleanup of expired tokens and sessions

## Production Deployment

### Security Checklist

- [ ] Use HTTPS everywhere
- [ ] Strong secret keys configured
- [ ] Rate limiting enabled
- [ ] Audit logging configured
- [ ] Security headers enabled
- [ ] Token expiration configured
- [ ] Regular security updates

### Monitoring

Monitor these metrics:
- Authentication success/failure rates
- Token validation latency
- API key usage patterns
- Permission violation attempts
- Session duration statistics

## Support

For questions or issues:
1. Check the test files for usage examples
2. Review the integration examples
3. Consult the API documentation
4. Create an issue in the repository