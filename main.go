package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"

	"task-api/db"
	"task-api/task"
)

func main() {
	database, err := db.Connect()
	if err != nil {
		log.Fatalf("database connection failed: %v", err)
	}

	repository := task.NewRepository(database)
	service := task.NewService(repository)
	handler := task.NewHandler(service)

	instance := os.Getenv("INSTANCE_NAME")
	if instance == "" {
		instance = "api-1"
	}

	router := gin.Default()

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":   "ok",
			"instance": instance,
		})
	})

	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message":  "Task API is running",
			"instance": instance,
		})
	})

	router.GET("/tasks", handler.GetAll)
	router.POST("/tasks", handler.Create)

	if err := router.Run(":8080"); err != nil {
		log.Fatalf("server failed to start: %v", err)
	}
}