package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type SignupRequest struct {
	Name     string `json:"name" binding:"required,min=2"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type AuthResponse struct {
	AccessToken  string      `json:"access_token"`
	RefreshToken string      `json:"refresh_token"`
	ExpiresIn    int         `json:"expires_in"`
	User         interface{} `json:"user"`
}

// Login godoc
// @Summary User login
// @Description Authenticate user and return JWT tokens
// @Tags auth
// @Accept json
// @Produce json
// @Param login body LoginRequest true "Login credentials"
// @Success 200 {object} AuthResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /v1/auth/login [post]
func (h *Handlers) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: Implement actual authentication
	// For now, return mock response for demo purposes
	if req.Email == "demo@househelper.app" && req.Password == "password123" {
		// Generate JWT token
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"user_id":      "demo-user-123",
			"household_id": "demo-household-456",
			"email":        req.Email,
			"exp":          time.Now().Add(time.Hour * 24).Unix(),
			"iat":          time.Now().Unix(),
		})

		tokenString, err := token.SignedString([]byte("your-super-secret-jwt-key-change-this-in-production"))
		if err != nil {
			h.logger.Error("Failed to sign JWT token", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
			return
		}

		refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"user_id": "demo-user-123",
			"exp":     time.Now().Add(time.Hour * 24 * 7).Unix(), // 7 days
			"iat":     time.Now().Unix(),
		})

		refreshTokenString, err := refreshToken.SignedString([]byte("your-super-secret-jwt-key-change-this-in-production"))
		if err != nil {
			h.logger.Error("Failed to sign refresh token", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate refresh token"})
			return
		}

		c.JSON(http.StatusOK, AuthResponse{
			AccessToken:  tokenString,
			RefreshToken: refreshTokenString,
			ExpiresIn:    86400, // 24 hours
			User: gin.H{
				"id":            "demo-user-123",
				"email":         req.Email,
				"name":          "Demo User",
				"household_id":  "demo-household-456",
				"created_at":    time.Now().Format(time.RFC3339),
			},
		})
		return
	}

	c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
}

// Signup godoc
// @Summary User signup
// @Description Create a new user account
// @Tags auth
// @Accept json
// @Produce json
// @Param signup body SignupRequest true "Signup data"
// @Success 201 {object} AuthResponse
// @Failure 400 {object} map[string]string
// @Router /v1/auth/signup [post]
func (h *Handlers) Signup(c *gin.Context) {
	var req SignupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// TODO: Implement actual user creation
	// For now, return mock response
	userID := "new-user-" + time.Now().Format("20060102150405")
	householdID := "household-" + time.Now().Format("20060102150405")

	// Generate JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":      userID,
		"household_id": householdID,
		"email":        req.Email,
		"exp":          time.Now().Add(time.Hour * 24).Unix(),
		"iat":          time.Now().Unix(),
	})

	tokenString, err := token.SignedString([]byte("your-super-secret-jwt-key-change-this-in-production"))
	if err != nil {
		h.logger.Error("Failed to sign JWT token", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Hour * 24 * 7).Unix(),
		"iat":     time.Now().Unix(),
	})

	refreshTokenString, err := refreshToken.SignedString([]byte("your-super-secret-jwt-key-change-this-in-production"))
	if err != nil {
		h.logger.Error("Failed to sign refresh token", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate refresh token"})
		return
	}

	c.JSON(http.StatusCreated, AuthResponse{
		AccessToken:  tokenString,
		RefreshToken: refreshTokenString,
		ExpiresIn:    86400,
		User: gin.H{
			"id":           userID,
			"email":        req.Email,
			"name":         req.Name,
			"household_id": householdID,
			"created_at":   time.Now().Format(time.RFC3339),
		},
	})
}

// RefreshToken godoc
// @Summary Refresh JWT token
// @Description Refresh an expired JWT token
// @Tags auth
// @Accept json
// @Produce json
// @Param refresh body map[string]string true "Refresh token"
// @Success 200 {object} AuthResponse
// @Failure 400 {object} map[string]string
// @Failure 401 {object} map[string]string
// @Router /v1/auth/refresh [post]
func (h *Handlers) RefreshToken(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Parse and validate refresh token
	token, err := jwt.Parse(req.RefreshToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte("your-super-secret-jwt-key-change-this-in-production"), nil
	})

	if err != nil || !token.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
		return
	}

	userID := claims["user_id"].(string)

	// Generate new access token
	newToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":      userID,
		"household_id": "demo-household-456", // TODO: Get from user record
		"email":        "demo@househelper.app", // TODO: Get from user record
		"exp":          time.Now().Add(time.Hour * 24).Unix(),
		"iat":          time.Now().Unix(),
	})

	tokenString, err := newToken.SignedString([]byte("your-super-secret-jwt-key-change-this-in-production"))
	if err != nil {
		h.logger.Error("Failed to sign new JWT token", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token": tokenString,
		"expires_in":   86400,
	})
}
