package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/la1665/task-manager/internal/handler"
)

func main() {
	// Load environment variables
	envErr := godotenv.Load(".env")
	if envErr != nil {
		fmt.Println("Loading env file failed", envErr)
		return
	}
	router := gin.Default()
	router.GET("/health", handler.Healthandler)

	fmt.Println("Server is runnig on port ", os.Getenv("PORT"))
	if err := router.Run(os.Getenv("PORT")); err != nil {
		log.Fatal(err)
	}
}
