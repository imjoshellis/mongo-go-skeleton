package main

import (
	"context"
	"mongogo/src/data"
	"mongogo/src/entities"
	"time"

	"github.com/gofiber/fiber/v2"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func main() {
	var user entities.User
	var err error
	user.Username = "imjoshellis"
	user.Email = "josh@imjoshlis.com"
	if user.ID == primitive.NilObjectID {
		log.Info("User has nil ObjectID. Attempting to save...")
	}
	err = user.Save()
	if err != nil {
		log.Fatal("Error reading user from db: %v", err)
	}
	if user.ID != primitive.NilObjectID {
		log.Info("User has a generated ObjectID. Save was successful.")
	}

	var readUser entities.User
	readUser.ID = user.ID
	if readUser.Email == "" {
		log.Info("New user entity created with matching ID. Attempting to read...")
	}
	err = readUser.Get()
	if err != nil {
		log.Fatal("Error reading user from db: %v", err)
	}
	if readUser.Email == user.Email {
		log.Info("New user has the same email as original. Read was successful.")
	}

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
