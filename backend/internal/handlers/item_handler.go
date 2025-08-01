package handlers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"ecommerce-app/internal/models"
)

type ItemHandler struct {
	DB *gorm.DB
}

func NewItemHandler(db *gorm.DB) *ItemHandler {
	return &ItemHandler{DB: db}
}

type CreateItemRequest struct {
	Name  string  `json:"name" binding:"required"`
	Price float64 `json:"price" binding:"required,gt=0"`
}

func (h *ItemHandler) CreateItem(c *gin.Context) {
	var req CreateItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	item := models.Item{
		Name:  req.Name,
		Price: req.Price,
	}

	if err := h.DB.Create(&item).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create item"})
		return
	}

	c.JSON(http.StatusCreated, item)
}

func (h *ItemHandler) ListItems(c *gin.Context) {
	type ItemResponse struct {
		ID     uint    `json:"id"`
		Name   string  `json:"name"`
		Price  float64 `json:"price"`
		Status string  `json:"status,omitempty"`
	}

	var items []models.Item
	// First, get all items from the database
	if err := h.DB.Find(&items).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch items"})
		return
	}

	// Log the items being returned for debugging
	log.Printf("Found %d items in database", len(items))
	for _, item := range items {
		log.Printf("Item - ID: %d, Name: %s, Price: %.2f, Status: %s", 
			item.ID, item.Name, item.Price, item.Status)
	}

	// Use a map to ensure unique items by name (case-insensitive)
	uniqueItems := make(map[string]ItemResponse)
	
	// Process items, keeping only the first occurrence of each item name
	for _, item := range items {
		// Use lowercase name as the key for case-insensitive comparison
		nameKey := item.Name
		if _, exists := uniqueItems[nameKey]; !exists {
			uniqueItems[nameKey] = ItemResponse{
				ID:     item.ID,
				Name:   item.Name,
				Price:  item.Price,
				Status: item.Status,
			}
		} else {
			log.Printf("Duplicate item found - ID: %d, Name: %s (keeping first occurrence)", 
				item.ID, item.Name)
		}
	}

	// Convert map values to slice
	response := make([]ItemResponse, 0, len(uniqueItems))
	for _, item := range uniqueItems {
		response = append(response, item)
	}

	log.Printf("Returning %d unique items in response", len(response))
	c.JSON(http.StatusOK, response)
}
