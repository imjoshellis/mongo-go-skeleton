package data

import (
	"context"
	"fmt"
	"os"
	"time"

	log "github.com/sirupsen/logrus"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type key string

var (
	Client     *mongo.Client
	Collection *mongo.Collection
)

const (
	hostKey     = key("hostKey")
	usernameKey = key("usernameKey")
	passwordKey = key("passwordKey")
	databaseKey = key("databaseKey")
)

func init() {
	var err error
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	ctx = context.WithValue(ctx, hostKey, os.Getenv("MONGO_HOST"))
	ctx = context.WithValue(ctx, usernameKey, os.Getenv("MONGO_USERNAME"))
	ctx = context.WithValue(ctx, passwordKey, os.Getenv("MONGO_PASSWORD"))
	ctx = context.WithValue(ctx, databaseKey, os.Getenv("MONGO_DATABASE"))
	db, err := configDB(ctx)
	if err != nil {
		log.Fatalf("Database configuration failed: %v", err)
	}
	Collection = db.Collection("users")
	_, err = Collection.DeleteMany(ctx, bson.M{})
	if err != nil {
		log.Fatalf("Deleting users collection failed %v", err)
	}
	log.Warn("Users collection was reset! You probably don't want this to happen in production...")
	_, err = Collection.Indexes().CreateOne(
		ctx,
		mongo.IndexModel{
			Keys:    bson.D{{Key: "email", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
	)
	if err != nil {
		log.Fatalf("Failed to create unique index on emails", err)
	}
	_, err = Collection.Indexes().CreateOne(
		ctx,
		mongo.IndexModel{
			Keys:    bson.D{{Key: "username", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
	)
	if err != nil {
		log.Fatalf("Failed to create unique index on usernames", err)
	}

	log.Info("Successfully connected to MongoDB")
}

func configDB(ctx context.Context) (*mongo.Database, error) {
	uri := fmt.Sprintf(`mongodb://%s:%s@%s/%s`,
		ctx.Value(usernameKey).(string),
		ctx.Value(passwordKey).(string),
		ctx.Value(hostKey).(string),
		ctx.Value(databaseKey).(string),
	)
	client, err := mongo.NewClient(options.Client().ApplyURI(uri))
	if err != nil {
		return nil, fmt.Errorf("couldn't connect to mongo: %v", err)
	}
	err = client.Connect(ctx)
	if err != nil {
		return nil, fmt.Errorf("mongo client couldn't connect with background context: %v", err)
	}
	db := client.Database("appName")
	return db, nil
}
