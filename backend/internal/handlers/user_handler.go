package handlers

import (
	"log"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"golang.org/x/crypto/bcrypt"
	"ecommerce-app/internal/models"
)

type UserHandler struct {
	DB *gorm.DB
}

func NewUserHandler(db *gorm.DB) *UserHandler {
	return &UserHandler{DB: db}
}

type SignupRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required,min=6"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (h *UserHandler) Signup(c *gin.Context) {
	var req SignupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request format",
			"details": err.Error(),
		})
		return
	}
	
	// Log the incoming request for debugging
	log.Printf("Signup request - Username: %s, Password length: %d\n", req.Username, len(req.Password))

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	// Create user with null token
	tx := h.DB.Begin()
	
	user := models.User{
		Username: req.Username,
		Password: string(hashedPassword),
		Token:    nil, // Explicitly set to null
	}

	if err := tx.Create(&user).Error; err != nil {
		tx.Rollback()
		log.Printf("Error creating user: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Could not create user",
			"details": err.Error(),
		})
		return
	}

	// Generate JWT token with the correct user ID
	claims := jwt.MapClaims{
		"user_id": user.ID,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	jwtSecret := "your-secure-jwt-secret-key-123"
	tokenString, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	// Update user with the generated token
	tokenStr := tokenString // Convert to string pointer
	if err := tx.Model(&user).Update("token", &tokenStr).Error; err != nil {
		tx.Rollback()
		log.Printf("Error updating user token: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update user token",
			"details": err.Error(),
		})
		return
	}

	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to complete registration"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "User created successfully",
		"user_id":  user.ID,
		"username": user.Username,
	})
}

func (h *UserHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Find user by username
	var user models.User
	if err := h.DB.Where("username = ?", req.Username).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		}
		return
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}

	// Log the user details for debugging
	log.Printf("Found user - ID: %d, Username: %s", user.ID, user.Username)
	if user.ID == 0 {
		log.Printf("WARNING: User ID is 0 for username: %s", req.Username)
		// Try to find the user again with a different query
		var userCheck models.User
		if err := h.DB.Raw("SELECT id, username FROM users WHERE username = ?", req.Username).Scan(&userCheck).Error; err == nil {
			log.Printf("User check - ID: %d, Username: %s", userCheck.ID, userCheck.Username)
			if userCheck.ID != 0 {
				// Update the user ID if found
				user.ID = userCheck.ID
				log.Printf("Updated user ID to: %d", user.ID)
			}
		}
	}

	// Generate JWT token with the correct user ID
	claims := jwt.MapClaims{
		"user_id": user.ID,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	jwtSecret := "your-secure-jwt-secret-key-123"
	tokenString, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		log.Printf("Error generating token: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	log.Printf("Generated token for user ID %d: %s", user.ID, tokenString)

	// Update user token in database using a transaction
	tx := h.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Log the user ID and token for debugging
	log.Printf("Updating token for user ID: %d, Token: %s", user.ID, tokenString)

	// Convert tokenString to a pointer
	tokenStr := tokenString
	result := tx.Model(&models.User{}).Where("id = ?", user.ID).Update("token", &tokenStr)
	if result.Error != nil {
		tx.Rollback()
		log.Printf("Error updating user token: %v", result.Error)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update user session",
			"details": result.Error.Error(),
		})
		return
	}

	// Log the number of rows affected
	log.Printf("Updated %d rows with new token", result.RowsAffected)

	if result.RowsAffected == 0 {
		tx.Rollback()
		log.Printf("No rows were updated for user ID: %d", user.ID)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update user session: user not found",
		})
		return
	}

	if err := tx.Commit().Error; err != nil {
		log.Printf("Error committing transaction: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to complete login",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token":   tokenString,
		"user_id": user.ID,
	})
}

// GetCurrentUser returns the currently authenticated user's information
func (h *UserHandler) GetCurrentUser(c *gin.Context) {
	// Get user ID from context (set by AuthMiddleware)
	userID, exists := c.Get("userID")
	if !exists {
		log.Println("GetCurrentUser: userID not found in context")
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	// Log the user ID for debugging
	log.Printf("GetCurrentUser: Fetching user with ID: %v", userID)

	// Convert userID to uint
	userIDUint, ok := userID.(uint)
	if !ok {
		log.Printf("GetCurrentUser: Invalid userID type: %T, value: %v", userID, userID)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Internal server error",
		})
		return
	}

	// Find user by ID
	var user models.User
	if err := h.DB.Where("id = ?", userIDUint).First(&user).Error; err != nil {
		log.Printf("GetCurrentUser: Error fetching user: %v", err)
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "User not found",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to fetch user",
			})
		}
		return
	}

	// Remove sensitive information
	user.Password = ""

	log.Printf("GetCurrentUser: Successfully retrieved user: %+v", user)
	c.JSON(http.StatusOK, user)
}

func (h *UserHandler) ListUsers(c *gin.Context) {
	var users []models.User
	if err := h.DB.Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
		return
	}

	// Remove sensitive information
	for i := range users {
		users[i].Password = ""
		users[i].Token = nil // Set to nil instead of empty string for *string
	}

	c.JSON(http.StatusOK, users)
}
