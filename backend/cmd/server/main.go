package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	_ "modernc.org/sqlite"
	"ecommerce-app/internal/config"
	"ecommerce-app/internal/handlers"
	"ecommerce-app/internal/middleware"
	"ecommerce-app/internal/models"
)

func main() {
	// Initialize database
	db, err := config.InitDB()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Auto-migrate the schema
	migrateDB(db)

	// Initialize handlers
	userHandler := handlers.NewUserHandler(db)
	itemHandler := handlers.NewItemHandler(db)
	cartHandler := handlers.NewCartHandler(db)
	orderHandler := handlers.NewOrderHandler(db)

	// Create Gin router
	r := gin.Default()

	// CORS middleware
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// API routes
	api := r.Group("/api")
	{
		// Public routes
		api.POST("/users", userHandler.Signup)
		api.POST("/users/login", userHandler.Login)
		api.GET("/users", userHandler.ListUsers)

		// Protected routes
		auth := api.Group("/")
		auth.Use(middleware.AuthMiddleware())
		{
			// User routes
			auth.GET("/users/me", userHandler.GetCurrentUser)

			// Items
			auth.POST("/items", itemHandler.CreateItem)
			auth.GET("/items", itemHandler.ListItems)

			// Carts
			auth.POST("/carts", cartHandler.AddToCart)
			auth.GET("/carts", cartHandler.GetCart)

			// Orders
			auth.POST("/orders", orderHandler.CreateOrder)
			auth.GET("/orders", orderHandler.ListOrders)
		}
	}

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server running on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func migrateDB(db *gorm.DB) {
	// Enable foreign key constraints for SQLite
	db.Exec("PRAGMA foreign_keys = ON")

	// Auto-migrate the models
	db.AutoMigrate(
		&models.User{},
		&models.Item{},
		&models.Cart{},
		&models.CartItem{},
		&models.Order{},
	)

	// Add any initial data if needed
	seedInitialData(db)
}

func seedInitialData(db *gorm.DB) {
	// Define our items with proper IDs
	items := []models.Item{
		{ID: 1, Name: "Laptop", Price: 999.99, Status: "available"},
		{ID: 2, Name: "Smartphone", Price: 699.99, Status: "available"},
		{ID: 3, Name: "Headphones", Price: 199.99, Status: "available"},
		{ID: 4, Name: "Keyboard", Price: 99.99, Status: "available"},
		{ID: 5, Name: "Mouse", Price: 49.99, Status: "available"},
	}

	// Update or create each item
	for _, item := range items {
		// Try to find existing item by name
		var existingItem models.Item
		if err := db.Where("name = ?", item.Name).First(&existingItem).Error; err == nil {
			// Item exists, update it with correct ID and status
			existingItem.ID = item.ID
			existingItem.Price = item.Price
			existingItem.Status = item.Status
			if err := db.Save(&existingItem).Error; err != nil {
				log.Printf("Failed to update item %s: %v", item.Name, err)
			}
		} else {
			// Item doesn't exist, create it
			if err := db.Create(&item).Error; err != nil {
				log.Printf("Failed to create item %s: %v", item.Name, err)
			}
		}
	}

	// Ensure the sequence is set correctly
	db.Exec("UPDATE sqlite_sequence SET seq = (SELECT MAX(id) FROM items) WHERE name='items';")
}
