package main

import (
	"cashcontrol/internal/config"
	"cashcontrol/internal/database"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}

	if err := database.Init(cfg); err != nil {
		panic(err)
	}

	if err := database.Migrate(); err != nil {
		panic(err)
	}

	defer database.Close()

	router := gin.Default()

	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Сервер работает",
		})
	})

	router.Run(cfg.ServerAddress)
}
