# mongo-go-skeleton
A simple starting point for using MongoDB with Go, created for a blog post on Dev.to.

Read: https://dev.to/imjoshellis/how-to-setup-mongodb-with-go-2ccb

---

After struggling through this and combining a few different resources, I wrote a reference guide for next time. My hope in sharing it is that it will help someone else.

Disclaimer: I'm still relatively new to using MongoDB in Go, so there's no guarantee any of this is the _best_ way... I just know it works for me. I'm happily open to suggestions!

## QuickStart

I set up a repo on GitHub if you want to jump straight into the code: <https://github.com/imjoshellis/mongo-go-skeleton>

## Prerequisites

This guide skips the basics. I'm assuming you have an instance of MongoDB ready to connect to and that you know how to set up a Go project with `go mod init`.

I'm also hiding imports in the snippets to save space, but you can check the [GitHub repo](https://github.com/imjoshellis/mongo-go-skeleton) for the files with full import info.

## User Entity

We'll start with the model so you can see where this is headed. I've found annotating a struct to be the simplest way to handle MongoDB in Go. Let's imagine you want to create a simple user type that has an id, username, and email. You'd declare the type like so.

The bson tags are important so the mongo driver can understand how to associate the data. The json tags won't be consumed in this project. But since you're likely to use this in the context of an API that sends json, you can see what both json and bson tags would look like together.

```go
// src/entities/user.go
type User struct {
 ID       primitive.ObjectID `bson:"_id, omitempty" json:"id"`
 Username string             `bson:"username" json:"username"`
 Email    string             `bson:"email" json:"email"`
}
```

We'll add two functions here for writing and reading users later, but for now, the struct is good enough.

## Connecting to the Database

In my case, I've chosen to create a separate package for data. We'll use the special `init` function to get the Mongo client up and running the first time its imported.

Since the focus here is on getting MongoDB up and running asap, we'll be working with the db directly in `main.go`, but I'd normally have an `app.go` with routes/controllers.

Let's start building the db file. [Full file on GitHub](https://github.com/imjoshellis/mongo-go-skeleton/blob/main/src/data/users.db.go)

Note: Some of the conventions are from [this fantastic guide](https://medium.com/@wembleyleach/how-to-use-the-official-mongodb-go-driver-9f8aff716fdb)

To start with, I've chosen to expose the mongo client and collection to the other parts of the app:

```go
// src/data/users.db.go
var (
 Client     *mongo.Client
 Collection *mongo.Collection
)
```

We'll be using environment variables for the database connection info, so let's set up some constants:

```go
// src/data/users.db.go
type key string

const (
 hostKey     = key("hostKey")
 usernameKey = key("usernameKey")
 passwordKey = key("passwordKey")
 databaseKey = key("databaseKey")
)
```

Next, we write an `init` function. Again, this is a special named function that will be run when the package is first imported, so it never has to be manually called as it will run once automatically:

```go
// src/data/users.db.go
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
 Collection = db.Collection("users") // Change me!
 log.Info("Successfully connected to MongoDB")
}
```

The main purpose of the `init` function is to grab the environment variables and set them onto the context. Then, it attempts to call `configDB`, which will actually create the db connection with error checks and return the database if it succeeds. You'll want to customize the database name:

```go
// src/data/users.db.go
func configDB(ctx context.Context) (*mongo.Database, error) {
 uri := fmt.Sprintf(`mongodb://%s:%s@%s/%s`,
  ctx.Value(usernameKey).(string),
  ctx.Value(passwordKey).(string),
  ctx.Value(hostKey).(string),
  ctx.Value(databaseKey).(string),
 )
 Client, err := mongo.NewClient(options.Client().ApplyURI(uri))
 if err != nil {
  return nil, fmt.Errorf("couldn't connect to mongo: %v", err)
 }
 err = Client.Connect(ctx)
 if err != nil {
  return nil, fmt.Errorf("client couldn't connect with context: %v", err)
 }
 db := Client.Database("appName") // Change me!
 return db, nil
}
```

Back in the `init` function, I've also added some temporary functions for development. The first wipes the collection every time `init` is run. For obvious reasons, you don't want this in production:

```go
// src/data/users.db.go
func init() {
// ...
 _, err = Collection.DeleteMany(ctx, bson.M{})
 if err != nil {
  log.Fatalf("Deleting users collection failed %v", err)
 }
 log.Warn("Users collection was reset! You probably don't want this to happen in production...")
// ...
}
```

The next thing I have is an example of how you would make fields unique:

```go
// src/data/users.db.go
func init() {
// ...
keys := []string{"email", "username"}
 for _, k := range keys {
  _, err = Collection.Indexes().CreateOne(
   ctx,
   mongo.IndexModel{
    Keys:    bson.D{{Key: k, Value: 1}},
    Options: options.Index().SetUnique(true),
   },
  )
  if err != nil {
   log.Fatalf("Failed to create unique index on %v: %v", k, err)
  }
  }
// ...
}
```

This will make it so both email and username are unique values. I've seen it argued that it's better to do this in the Mongo shell, and that's a valid way of doing it. I find doing it this way is easier to avoid errors, especially during development.

## Save and Get

Now that we have the database ready to go, let's go back to the User entity file and create two functions for saving to and reading from MongoDB:

```go
// src/entities/user.go
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
```

This part is pretty straightforward. We get context, use the `InsertOne()` function to insert the user, and set the `ID` field on the user based on the result of `InsertOne()`.

Here's the `Get()` function:

```go
// src/entities/user.go
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
```

As you can see, we're setting up a filter based on the user's ID, then calling `FindOne()` to the collection. The `Decode()` function will fill in the blanks on the user if successful.

## Testing it Out in Main

Now, normally you'd set up API routes and controllers and all that jazz, but to quickly test whether this is working, I've set up `main()` to make a user, save them, and query the database for them. I'm using fiber to make an http server to make sure the connection to MongoDB is kept open long enough.

```go
// src/main.go
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
```

## Running Main with Variables

Finally, you'll want to make sure to include the correct variables when you run `go run main.go`.

In my case, it looks like this:

```sh
MONGO_HOST="localhost:27017" MONGO_USERNAME=myUserAdmin MONGO_PASSWORD=admin go run src/main.go
```

If everything went well, you should see a bunch of logs, followed by a message from fiber saying the server is up and running:

```txt
WARN[0000] Users collection was reset! You probably don't want this to happen in production...
INFO[0000] Successfully connected to MongoDB
INFO[0000] User has nil ObjectID. Attempting to save...
INFO[0000] User has a generated ObjectID. Save was successful.
INFO[0000] New user entity created with matching ID. Attempting to read...
INFO[0000] New user has the same email as original. Read was successful.
```

## Conclusion

That's it! You'll of course want to add a lot more functionality, including API endpoints, update/delete methods, more entities, etc... But hopefully this helped you get started with MongoDB and Go!
