package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"ecommerce-app/internal/models"
)

type OrderHandler struct {
	DB *gorm.DB
}

func NewOrderHandler(db *gorm.DB) *OrderHandler {
	return &OrderHandler{DB: db}
}

func (h *OrderHandler) CreateOrder(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Start a transaction
	tx := h.DB.Begin()

	// Get user's active cart with items and their product details
	var cart models.Cart
	if err := tx.Where("user_id = ? AND status = ?", userID, "active").First(&cart).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "No active cart found",
			"details": err.Error(),
		})
		return
	}

	// Get cart items with product details
	var cartItems []struct {
		ID       uint    `gorm:"column:id"`
		Name     string  `gorm:"column:name"`
		Price    float64 `gorm:"column:price"`
		Quantity int     `gorm:"column:quantity"`
	}

	if err := tx.Table("cart_items").
		Select("items.id, items.name, items.price, cart_items.quantity").
		Joins("JOIN items ON items.id = cart_items.item_id").
		Where("cart_items.cart_id = ?", cart.ID).
		Scan(&cartItems).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to fetch cart items",
			"details": err.Error(),
		})
		return
	}

	// Check if cart is empty
	if len(cartItems) == 0 {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot create order with empty cart"})
		return
	}

	// Create order
	order := models.Order{
		UserID: userID.(uint),
		CartID: cart.ID,
		Status: "completed",
	}

	if err := tx.Create(&order).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create order",
			"details": err.Error(),
		})
		return
	}

	// Update cart status to 'ordered'
	if err := tx.Model(&models.Cart{}).Where("id = ?", cart.ID).Update("status", "ordered").Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to update cart status",
			"details": err.Error(),
		})
		return
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Transaction failed",
			"details": err.Error(),
		})
		return
	}

	// Prepare response with order details
	response := gin.H{
		"message":    "Order created successfully",
		"order_id":   order.ID,
		"cart_id":    order.CartID,
		"status":     order.Status,
		"created_at": order.CreatedAt,
		"items":      cartItems,
	}

	c.JSON(http.StatusCreated, response)
}

func (h *OrderHandler) ListOrders(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var orders []models.Order
	if err := h.DB.Where("user_id = ?", userID).Find(&orders).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch orders"})
		return
	}

	// Get order details with cart items
	var orderDetails []map[string]interface{}

	for _, order := range orders {
		var cart models.Cart
		if err := h.DB.Preload("Items").First(&cart, order.CartID).Error; err != nil {
			continue
		}

		// Get cart items with details
		var items []map[string]interface{}
		rows, err := h.DB.Table("cart_items").
			Select("items.id, items.name, items.price, cart_items.quantity").
			Joins("JOIN items ON items.id = cart_items.item_id").
			Where("cart_items.cart_id = ?", cart.ID).
			Rows()

		if err == nil {
			for rows.Next() {
				var id uint
				var name string
				var price float64
				var quantity int
				rows.Scan(&id, &name, &price, &quantity)
				items = append(items, map[string]interface{}{
					"id":       id,
					"name":     name,
					"price":    price,
					"quantity": quantity,
				})
			}
			rows.Close()
		}

		orderDetails = append(orderDetails, map[string]interface{}{
			"order_id":   order.ID,
			"cart_id":    order.CartID,
			"status":     order.Status,
			"created_at": order.CreatedAt,
			"items":      items,
		})
	}

	c.JSON(http.StatusOK, orderDetails)
}
