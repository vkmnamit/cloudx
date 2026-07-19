package main

import (
	"context"
	"log"

	"github.com/vkmnamit/cloudx/configs"
	"github.com/vkmnamit/cloudx/internal/router"
	"github.com/vkmnamit/cloudx/pkg/database"
)

func main() {

	configs.LoadEnv()

	conn, err := database.Connect()
	if err != nil {
		log.Fatal(err)
	}

	defer conn.Close(context.Background())

	r := router.Setup(conn)

	log.Println("🚀 CloudX Started")

	r.Run(":" + configs.AppConfig.Port)

}
