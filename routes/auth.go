package routes

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"database/sql"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AuthRequest struct {
	Username string `json:"username"`
}

type AuthResponse struct {
	EncryptedToken string `json:"encryptedToken"`
}

type TokenData struct {
	Token     string    `json:"token"`
	Timestamp time.Time `json:"timestamp"`
	Username  string    `json:"username"`
}

var tokenStorage = make(map[string]TokenData)

// generateToken generates a random token.
func generateToken() (string, error) {
	newToken := uuid.New().String()

	return newToken, nil
}

// encryptToken encrypts the token with the user's public key.
func encryptToken(token string, publicKeyPEM string) (string, error) {
	// Decode the PEM-encoded public key
	block, _ := pem.Decode([]byte(publicKeyPEM))
	if block == nil {
		return "", fmt.Errorf("failed to decode PEM block containing public key")
	}

	// Parse the public key
	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return "", fmt.Errorf("failed to parse public key: %w", err)
	}

	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		return "", errors.New("public key is not an RSA key")
	}

	// Encrypt the token
	encryptedTokenBytes, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, rsaPub, []byte(token), nil)
	if err != nil {
		return "", fmt.Errorf("failed to encrypt token: %w", err)
	}

	// Transform base 64
	encryptedToken := base64.StdEncoding.EncodeToString(encryptedTokenBytes)

	return encryptedToken, nil
}

// verifyToken verifies the token.
func verifyToken(token string) (string, bool) {
	tokenData, ok := tokenStorage[token]
	if !ok {
		return "", false
	}
	// Check if the token has expired (e.g., after 1 hour)
	if time.Since(tokenData.Timestamp) > time.Hour {
		delete(tokenStorage, token)
		return "", false
	}
	return tokenData.Username, true
}

func getUserFromToken(token string) (string, bool) {
	return verifyToken(token)
}

// requestToken godoc
// @Summary Request a token
// @Description Request a token for authentication
// @Tags auth
// @Accept json
// @Produce json
// @Param authRequest body AuthRequest true "Authentication request"
// @Success 200 {object} AuthResponse
// @Router /auth/token [post]
func RequestToken(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var authRequest AuthRequest
		if err := c.ShouldBindJSON(&authRequest); err != nil {
			c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if authRequest.Username == "" {
			c.IndentedJSON(http.StatusBadRequest, gin.H{"error": "username is required"})
			return
		}

		// Get the user's public key from the database
		var publicKeyPEM string
		err := db.QueryRow("SELECT public_key FROM users WHERE username = ?", authRequest.Username).Scan(&publicKeyPEM)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				c.IndentedJSON(http.StatusNotFound, gin.H{"error": "user not found"})
			} else {
				log.Println(err)
				c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "failed to get user public key"})
			}
			return
		}

		// Generate a new token
		token, err := generateToken()
		if err != nil {
			log.Println(err)
			c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
			return
		}

		// Encrypt the token with the user's public key
		encryptedToken, err := encryptToken(token, publicKeyPEM)
		if err != nil {
			log.Println(err)
			c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": "failed to encrypt token"})
			return
		}

		// Store the token in the storage
		tokenStorage[token] = TokenData{Token: token, Timestamp: time.Now(), Username: authRequest.Username}

		c.IndentedJSON(http.StatusOK, AuthResponse{EncryptedToken: encryptedToken})
	}
}

// AuthMiddleware is a middleware to check for valid tokens.
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authorization header is required"})
			return
		}

		// Check if the header is in the correct format
		if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header format"})
			return
		}

		token := authHeader[7:]
		usernameToken, ok := getUserFromToken(token)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		usernameHeader := c.GetHeader("X-User")
		if usernameHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "X-User header is required"})
			return
		}

		if usernameToken != usernameHeader {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid user for this token"})
			return
		}

		c.Next()
	}
}

func SetupAuthRoutes(router *gin.Engine, db *sql.DB) {
	authRoutes := router.Group("/auth")
	{
		authRoutes.POST("/token", RequestToken(db))
	}
}


// Comments for Swagger concerning Auth
// @securitydefinitions.basicAuth
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
// @description Type "Bearer {token}" to correctly authenticate.
