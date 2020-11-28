package entities

import (
	"context"
	"fmt"
	"mongogo/src/data"
	"time"

	"github.com/labstack/gommon/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID       primitive.ObjectID `bson:"_id, omitempty" json:"id"`
	Username string             `bson:"username" json:"username"`
	Email    string             `bson:"email" json:"email"`
}

// Get tries to get a user from the database
func (u *User) Get() error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	filter := bson.D{{Key: "_id", Value: u.ID}}
	err := data.Collection.FindOne(ctx, filter).Decode(&u)
	if err != nil {
		return fmt.Errorf("user %s not found", u.ID.Hex())
	}
	return nil
}

// Save tries to save a user to the database
func (u *User) Save() error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	res, err := data.Collection.InsertOne(ctx, bson.M{"username": u.Username, "email": u.Email})
	if err != nil {
		log.Error(err)
		return fmt.Errorf("there was a problem saving to the db")
	}
	u.ID = res.InsertedID.(primitive.ObjectID)
	return nil
}
