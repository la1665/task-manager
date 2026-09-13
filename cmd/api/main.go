// @title           Task Manager API
// @version         1.0
// @description     A simple task management REST API built with Gin, PostgreSQL, and TDD.
// @host            localhost:8080
// @BasePath        /
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	"github.com/la1665/task-manager/internal/handler"
	"github.com/la1665/task-manager/internal/metrics"
	"github.com/la1665/task-manager/internal/repository"
	"github.com/la1665/task-manager/internal/service"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "github.com/la1665/task-manager/docs"
)

func main() {
	// Load environment variables
	envErr := godotenv.Load(".env")
	if envErr != nil {
		fmt.Println("Loading env file failed", envErr)
	}

	// Database connection
	db, err := sqlx.Connect("pgx", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	taskRepo := repository.NewPostgresTaskRepository(db)
	taskService := service.NewTaskService(taskRepo)
	taskHandler := handler.NewTaskHandler(taskService)

	// Initialize tasks_count metric
	if existing, err := taskService.ListTasks(context.Background()); err == nil {
		metrics.TasksCount.Set(float64(len(existing)))
	} else {
		log.Printf("warning: failed to initialize tasks_count metric: %v", err)
	}

	router := gin.Default()
	router.Use(handler.TracingMiddleware())
	router.Use(handler.MetricsMiddleware())
	router.GET("/health", handler.HealthHandler)
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	tasks := router.Group("/tasks")
	{
		tasks.POST("", taskHandler.CreateTask)
		tasks.GET("", taskHandler.ListTasks)
		tasks.GET("/:id", taskHandler.GetTask)
		tasks.PUT("/:id", taskHandler.UpdateTask)
		tasks.DELETE("/:id", taskHandler.DeleteTask)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	fmt.Println("Server is running on port", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}
