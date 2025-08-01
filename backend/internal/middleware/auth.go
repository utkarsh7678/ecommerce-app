package middleware

import (
	"net/http"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

var jwtSecret = []byte("your-secure-jwt-secret-key-123") // In production, use environment variable

type Claims struct {
	UserID uint `json:"user_id"`
	jwt.StandardClaims
}

// MapClaims is used for parsing the JWT token claims
type MapClaims map[string]interface{}

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get token from Authorization header
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header is required",
			})
			c.Abort()
			return
		}

		// Remove 'Bearer ' prefix if present
		tokenString = strings.TrimSpace(strings.TrimPrefix(tokenString, "Bearer "))
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid authorization token format",
			})
			c.Abort()
			return
		}

		// Parse and validate the token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Validate the alg is what you expect
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return jwtSecret, nil
		})

		// Handle token parsing errors
		if err != nil {
			errMsg := "Invalid or expired token"
			if ve, ok := err.(*jwt.ValidationError); ok {
				switch {
				case ve.Errors&jwt.ValidationErrorMalformed != 0:
					errMsg = "Malformed token"
				case ve.Errors&(jwt.ValidationErrorExpired|jwt.ValidationErrorNotValidYet) != 0:
					errMsg = "Token is either expired or not active yet"
				default:
					errMsg = "Error processing token"
				}
			}
			c.JSON(http.StatusUnauthorized, gin.H{"error": errMsg})
			c.Abort()
			return
		}

		// Check if token is valid
		if !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		// Extract claims
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			c.Abort()
			return
		}

		// Get user ID from claims
		userID, ok := claims["user_id"].(float64) // JSON numbers are float64
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in token"})
			c.Abort()
			return
		}

		// Set user ID in context for downstream handlers
		c.Set("userID", uint(userID))
		c.Next()
	}
}
