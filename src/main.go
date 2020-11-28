package main

import (
	"context"
	"mongogo/src/data"
	"mongogo/src/entities"
	"time"

	"github.com/gofiber/fiber/v2"
	log "github.com/sirupsen/logrus"
)

func main() {
	var user entities.User
	var err error
	user.Username = "imjoshellis"
	user.Email = "josh@imjoshlis.com"

	err = user.Save()
	if err != nil {
		log.Fatal("Error reading user from db: %v", err)
	}
	log.Info("User successfully saved to db.")

	err = user.Get()
	if err != nil {
		log.Fatal("Error reading user from db: %v", err)
	}
	log.Info("User successfully read from db.")

	defer func() {
		log.Println("Disconnecting from MongoDB...")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := data.Client.Disconnect(ctx); err != nil {
			log.Fatal(err)
		}
	}()

	app := fiber.New()
	log.Fatal(app.Listen("localhost:8080"))
}
