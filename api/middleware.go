package main

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware(c *gin.Context) (int, error) {
	log.Printf("AuthMiddleware: Starting authentication for request to %s", c.Request.URL.Path)
	
	authHeader := c.GetHeader("Authorization")
    if authHeader == "" {
        log.Printf("AuthMiddleware: Authorization header missing")
        c.JSON(http.StatusUnauthorized, gin.H{
            "error": "Authorization header required",
            "code": "MISSING_AUTH_HEADER",
        })
        return -1, errors.New("authorization header required")
    }
    
    log.Printf("AuthMiddleware: Authorization header found: %s", authHeader[:min(len(authHeader), 20)] + "...")
    
    // Check if it's a Bearer token
    if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
        log.Printf("AuthMiddleware: Invalid authorization format. Expected 'Bearer <token>', got: %s", authHeader[:min(len(authHeader), 20)] + "...")
        c.JSON(http.StatusUnauthorized, gin.H{
            "error": "Invalid authorization format. Use 'Bearer <token>'",
            "code": "INVALID_AUTH_FORMAT",
        })
        return -1, errors.New("invalid authorization format. Use 'Bearer <token>'")
    }
    
    tokenString := authHeader[7:] // Remove "Bearer " prefix
    log.Printf("AuthMiddleware: Token extracted (first 20 chars): %s...", tokenString[:min(len(tokenString), 20)])
    
    // Verify JWT and extract user ID
    tokenUserID, err := VerifyJWT(tokenString)
    if err != nil {
        log.Printf("AuthMiddleware: JWT verification failed: %v", err)
        c.JSON(http.StatusUnauthorized, gin.H{
            "error": "Invalid or expired token",
            "code": "INVALID_TOKEN",
            "details": err.Error(),
        })
        return -1, errors.New("invalid or expired token")
    }
    
    log.Printf("AuthMiddleware: Authentication successful for user ID: %d", tokenUserID)
    return tokenUserID, nil
}

// Helper function to safely get minimum of two integers
func min(a, b int) int {
    if a < b {
        return a
    }
    return b
}