package handlers

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"ecommerce-app/internal/models"
)

type CartHandler struct {
	DB *gorm.DB
}

// generateSessionID creates a new random session ID
func generateSessionID() string {
	bytes := make([]byte, 16) // 128 bits
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to timestamp if crypto/rand fails (shouldn't happen)
		return "sess_" + fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return "sess_" + hex.EncodeToString(bytes)
}

func NewCartHandler(db *gorm.DB) *CartHandler {
	return &CartHandler{DB: db}
}

type AddToCartRequest struct {
	ItemID uint `json:"item_id" binding:"required"`
	// Note: The JSON tag must match exactly what's sent from the frontend (snake_case)
}

func (h *CartHandler) AddToCart(c *gin.Context) {
	// Read and restore request body
	bodyBytes, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		log.Printf("Error reading request body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}
	c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))

	// Log request
	log.Printf("AddToCart - Headers: %+v", c.Request.Header)
	log.Printf("Request body: %s", string(bodyBytes))

	// Get user ID from context if authenticated
	userIDVal, isAuthenticated := c.Get("userID")
	var userID uint
	var userIDPtr *uint

	if isAuthenticated && userIDVal != nil {
		// The auth middleware sets this as uint, but let's handle both float64 and uint
		switch v := userIDVal.(type) {
		case float64:
			userID = uint(v)
		case uint:
			userID = v
		case int:
			userID = uint(v)
		default:
			log.Printf("Unexpected user ID type in context: %T", userIDVal)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID format"})
			return
		}
		userIDPtr = &userID
		log.Printf("User is authenticated with ID: %d (type: %T)", userID, userIDVal)
	} else {
		log.Println("User is not authenticated or userID is nil")
	}

	// Get or generate session ID
	sessionID := c.GetHeader("X-Session-ID")
	if sessionID == "" && !isAuthenticated {
		sessionID = generateSessionID()
		c.Header("X-Session-ID", sessionID)
	}

	// Parse request body
	var input AddToCartRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	log.Printf("AddToCart request - UserID: %v, SessionID: %s, ItemID: %d", 
		userID, sessionID, input.ItemID)

	// Start transaction
	tx := h.DB.Begin()
	if tx.Error != nil {
		log.Printf("Error starting transaction: %v", tx.Error)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction"})
		return
	}

	// Ensure we rollback in case of a panic
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			log.Printf("Recovered from panic in AddToCart: %v", r)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		}
	}()

	// Get or create cart
	var cart models.Cart
	var cartErr error

	if isAuthenticated && userIDPtr != nil {
		// Handle authenticated user cart
		cartErr = tx.Where("user_id = ? AND status = ?", userID, "active").First(&cart).Error

		if cartErr != nil {
			if cartErr == gorm.ErrRecordNotFound {
				// Create new cart for authenticated user
				cart = models.Cart{
					UserID:    userIDPtr,
					SessionID: "",
					Status:    "active",
				}
				if err := tx.Create(&cart).Error; err != nil {
					tx.Rollback()
					log.Printf("Error creating user cart: %v", err)
					c.JSON(http.StatusInternalServerError, gin.H{
						"error":   "Failed to create cart",
						"details": err.Error(),
					})
					return
				}
				log.Printf("Created new cart ID: %d for user ID: %d", cart.ID, userID)
			} else {
				tx.Rollback()
				log.Printf("Error finding user's cart: %v", cartErr)
				c.JSON(http.StatusInternalServerError, gin.H{
					"error":   "Failed to find cart",
					"details": cartErr.Error(),
				})
				return
			}
		} else {
			log.Printf("Using existing cart ID: %d for user ID: %d", cart.ID, userID)
		}
	} else if sessionID != "" {
		// Handle unauthenticated user with session
		cartErr = tx.Where("session_id = ? AND status = ?", sessionID, "active").First(&cart).Error

		if cartErr != nil {
			if cartErr == gorm.ErrRecordNotFound {
				// Create new cart for unauthenticated user
				cart = models.Cart{
					SessionID: sessionID,
					Status:    "active",
				}
				if err := tx.Create(&cart).Error; err != nil {
					tx.Rollback()
					log.Printf("Error creating session cart: %v", err)
					c.JSON(http.StatusInternalServerError, gin.H{
						"error":   "Failed to create cart",
						"details": err.Error(),
					})
					return
				}
				log.Printf("Created new cart ID: %d for session ID: %s", cart.ID, sessionID)
			} else {
				tx.Rollback()
				log.Printf("Error finding session cart: %v", cartErr)
				c.JSON(http.StatusInternalServerError, gin.H{
					"error":   "Failed to find cart",
					"details": cartErr.Error(),
				})
				return
			}
		} else {
			log.Printf("Using existing cart ID: %d for session ID: %s", cart.ID, sessionID)
		}
	} else {
		tx.Rollback()
		log.Println("Neither user ID nor session ID provided")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Authentication or session ID required"})
		return
	}

	// Check if item exists and is available
	var item models.Item
	if err := tx.First(&item, input.ItemID).Error; err != nil {
		tx.Rollback()
		if err == gorm.ErrRecordNotFound {
			log.Printf("Item with ID %d not found", input.ItemID)
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Item not found",
			})
		} else {
			log.Printf("Error checking item: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Error checking item",
				"details": err.Error(),
			})
		}
		return
	}

	// Check if item is available
	if item.Status != "available" {
		tx.Rollback()
		log.Printf("Item with ID %d is not available for purchase. Status: %s", item.ID, item.Status)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Item is not available for purchase",
			"status": item.Status,
		})
		return
	}

	// Add item to cart or update quantity if already exists
	var cartItem models.CartItem
	itemErr := tx.
		Where("cart_id = ? AND item_id = ?", cart.ID, input.ItemID).
		First(&cartItem).Error

	if itemErr != nil {
		if itemErr == gorm.ErrRecordNotFound {
			// Create new cart item
			log.Printf("Creating new cart item - CartID: %d, ItemID: %d", cart.ID, input.ItemID)
			cartItem = models.CartItem{
				CartID:   cart.ID,
				ItemID:   input.ItemID,
				Quantity: 1,
			}
			if err := tx.Create(&cartItem).Error; err != nil {
				tx.Rollback()
				log.Printf("Error creating cart item: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Failed to add item to cart",
					"details": err.Error(),
				})
				return
			}
			log.Printf("Added new item %d to cart %d", input.ItemID, cart.ID)
		} else {
			tx.Rollback()
			log.Printf("Error checking cart items: %v", itemErr)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to process cart items",
				"details": itemErr.Error(),
			})
			return
		}
	} else {
		// Update quantity
		cartItem.Quantity++
		if err := tx.Model(&cartItem).
			Update("quantity", cartItem.Quantity).
			Error; err != nil {
			tx.Rollback()
			log.Printf("Error updating cart item quantity: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to update cart",
				"details": err.Error(),
			})
			return
		}
		log.Printf("Updated quantity for item %d in cart %d to %d", 
			input.ItemID, cart.ID, cartItem.Quantity)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		log.Printf("Error committing transaction: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update cart",
			"details": err.Error(),
		})
		return
	}

	// Get updated cart with items to return
	var updatedCart models.Cart
	if err := h.DB.Preload("Items").First(&updatedCart, cart.ID).Error; err != nil {
		log.Printf("Error fetching updated cart: %v", err)
		c.JSON(http.StatusOK, gin.H{
			"message": "Item added to cart, but could not fetch updated cart",
			"cart_id": cart.ID,
		})
		return
	}

	// Return success response with updated cart
	c.JSON(http.StatusOK, gin.H{
		"message": "Item added to cart successfully",
		"cart_id": cart.ID,
		"cart":    updatedCart,
	})
}

func (h *CartHandler) GetCart(c *gin.Context) {
	// Get user ID from context if authenticated
	userIDVal, isAuthenticated := c.Get("userID")
	var userID uint
	var userIDPtr *uint

	if isAuthenticated && userIDVal != nil {
		// Handle different possible types for user ID
		switch v := userIDVal.(type) {
		case float64:
			userID = uint(v)
		case uint:
			userID = v
		case int:
			userID = uint(v)
		default:
			log.Printf("Unexpected user ID type in GetCart: %T", userIDVal)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID format"})
			return
		}
		userIDPtr = &userID
		log.Printf("GetCart - User is authenticated with ID: %d (type: %T)", userID, userIDVal)
	} else {
		log.Println("GetCart - User is not authenticated or userID is nil")
	}

	// Get session ID from header
	sessionID := c.GetHeader("X-Session-ID")
	if sessionID != "" {
		log.Printf("GetCart - Using session ID: %s", sessionID)
	}

	if !isAuthenticated && sessionID == "" {
		// No session ID and not authenticated - return empty cart
		log.Println("GetCart - No session ID and not authenticated")
		c.JSON(http.StatusOK, gin.H{
			"message": "No active cart found",
			"cart":    nil,
		})
		return
	}

	// Start a transaction for read consistency
	tx := h.DB.Begin()
	defer tx.Rollback()

	// Try to find active cart with items
	var cart models.Cart
	query := tx.Preload("Items").
		Joins("LEFT JOIN cart_items ON carts.id = cart_items.cart_id")

	if isAuthenticated && userIDPtr != nil {
		// For authenticated users, look for their cart by user ID
		log.Printf("Looking for cart for user ID: %d", *userIDPtr)
		query = query.Where("carts.user_id = ? AND carts.status = ?", *userIDPtr, "active")
	} else if sessionID != "" {
		// For unauthenticated users, look for cart by session ID
		log.Printf("Looking for cart with session ID: %s", sessionID)
		query = query.Where("carts.session_id = ? AND carts.status = ?", sessionID, "active")
	} else {
		tx.Rollback()
		log.Println("No user ID or session ID provided")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Authentication or session ID required"})
		return
	}

	err := query.First(&cart).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			tx.Rollback()
			log.Printf("No active cart found - Authenticated: %v, UserID: %v, SessionID: %s", 
				isAuthenticated, userID, sessionID)
			c.JSON(http.StatusOK, gin.H{
				"message": "No active cart found",
				"cart":    nil,
			})
			return
		}
		tx.Rollback()
		log.Printf("Error fetching cart - Authenticated: %v, UserID: %v, SessionID: %s, Error: %v", 
			isAuthenticated, userID, sessionID, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch cart",
			"details": err.Error(),
		})
		return
	}

	// First, let's verify the cart items exist in the database
	var rawItems []map[string]interface{}
	if err := tx.Raw("SELECT * FROM cart_items WHERE cart_id = ?", cart.ID).Scan(&rawItems).Error; err != nil {
		tx.Rollback()
		log.Printf("Error fetching raw cart items: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to verify cart items",
			"details": err.Error(),
		})
		return
	}
	log.Printf("Raw cart items for cart ID %d: %+v", cart.ID, rawItems)

	// Let's check if the items in the cart exist in the items table
	for _, item := range rawItems {
		itemID, ok := item["item_id"].(int64)
		if !ok {
			log.Printf("Could not convert item_id to int64: %v", item["item_id"])
			continue
		}

		// Check if item exists in items table
		var count int64
		if err := tx.Raw("SELECT COUNT(*) FROM items WHERE id = ?", itemID).Scan(&count).Error; err != nil {
			log.Printf("Error checking if item %d exists: %v", itemID, err)
			continue
		}

		if count == 0 {
			log.Printf("Item with ID %d not found in items table", itemID)
		} else {
			log.Printf("Item with ID %d exists in items table", itemID)
			
			// Get item details for debugging
			var itemDetails map[string]interface{}
			if err := tx.Raw("SELECT * FROM items WHERE id = ?", itemID).Scan(&itemDetails).Error; err != nil {
				log.Printf("Error fetching item %d details: %v", itemID, err)
			} else {
				log.Printf("Item %d details: %+v", itemID, itemDetails)
			}
		}
	}

	// Now let's check the items table structure
	var itemColumns []string
	if err := tx.Raw("SELECT column_name FROM information_schema.columns WHERE table_name = 'items'").Pluck("column_name", &itemColumns).Error; err != nil {
		log.Printf("Error fetching items table columns: %v", err)
	} else {
		log.Printf("Items table columns: %v", itemColumns)
	}

	// First, let's check the structure of the items table
	var itemsColumns []string
	if err := tx.Raw(`
		SELECT column_name, data_type 
		FROM information_schema.columns 
		WHERE table_name = 'items' 
		ORDER BY ordinal_position
	`).Scan(&itemsColumns).Error; err != nil {
		log.Printf("Error fetching items table columns: %v", err)
	} else {
		log.Printf("Items table structure: %+v", itemsColumns)
	}

	// Get sample data from items table
	var rawItemData []map[string]interface{}
	if err := tx.Raw("SELECT * FROM items LIMIT 5").Scan(&rawItemData).Error; err != nil {
		log.Printf("Error fetching items: %v", err)
	} else {
		log.Printf("Sample items from database: %+v", rawItemData)
	}

	// Check if there are any items in the cart that don't exist in the items table
	var missingItems []map[string]interface{}
	if err := tx.Raw(`
		SELECT ci.* 
		FROM cart_items ci
		LEFT JOIN items i ON i.id = ci.item_id
		WHERE ci.cart_id = ? AND i.id IS NULL
	`, cart.ID).Scan(&missingItems).Error; err == nil && len(missingItems) > 0 {
		log.Printf("Found %d cart items with no matching item in items table: %+v", len(missingItems), missingItems)
	}

	// Get cart items with item details
	type CartItemWithDetails struct {
		ID        uint    `gorm:"column:id"`
		CartID    uint    `gorm:"column:cart_id"`
		ItemID    uint    `gorm:"column:item_id"`
		Quantity  int     `gorm:"column:quantity"`
		Name      string  `gorm:"column:name"`
		Price     float64 `gorm:"column:price"`
	}

	var cartItems []CartItemWithDetails
	err = tx.Raw(`
		SELECT ci.*, i.name, i.price
		FROM cart_items ci
		INNER JOIN items i ON i.id = ci.item_id
		WHERE ci.cart_id = ?
	`, cart.ID).Scan(&cartItems).Error

	if err != nil {
		tx.Rollback()
		log.Printf("Error in raw SQL query: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch cart items",
			"details": err.Error(),
		})
		return
	}

	log.Printf("Cart items with details: %+v", cartItems)

	if len(cartItems) == 0 {
		log.Printf("No items found in cart %d", cart.ID)
		c.JSON(http.StatusOK, gin.H{
			"message": "Cart is empty",
			"cart_id": cart.ID,
		})
		return
	}

	// Convert to the expected format
	items := make([]map[string]interface{}, 0, len(cartItems))
	total := 0.0

	for _, item := range cartItems {
		total += float64(item.Quantity) * item.Price
		items = append(items, map[string]interface{}{
			"id":       item.ItemID,
			"name":     item.Name,
			"price":    item.Price,
			"quantity": item.Quantity,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"cart_id": cart.ID,
		"items":   items,
		"total":   total,
	})
}
