package mongo

import (
	"context"
	"os"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var (
	ctx         = context.TODO()
	mongoClient *mongo.Client
)

func Init() {
	uri := os.Getenv("MONGODB_URI")
	if uri == "" {
		panic("MONGODB_URI environment variable is not set")
	}

	var err error
	mongoClient, err = mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		panic(err)
	}
}

func AddWebsite(email string, website string) error {
	collection := mongoClient.Database("downtimetracker").Collection("users")

	filter := bson.M{"email": email}

	update := bson.M{
		"$setOnInsert": bson.M{
			"verified": false,
		},
		"$addToSet": bson.M{
			"websites": website,
		},
	}

	opts := options.UpdateOne().SetUpsert(true)

	_, err := collection.UpdateOne(ctx, filter, update, opts)
	return err
}

func GetWebsites(email string) ([]string, error) {
	collection := mongoClient.Database("downtimetracker").Collection("users")

	var result struct {
		Websites []string `bson:"websites"`
	}

	err := collection.FindOne(ctx, bson.M{"email": email}).Decode(&result)
	if err != nil {
		return nil, err
	}

	return result.Websites, nil
}

func RemoveWebsite(email string, website string) error {
	collection := mongoClient.Database("downtimetracker").Collection("users")

	filter := bson.M{"email": email}
	update := bson.M{
		"$pull": bson.M{
			"websites": website,
		},
	}

	_, err := collection.UpdateOne(ctx, filter, update)
	return err
}

func IsMail(email string) bool {
	collection := mongoClient.Database("downtimetracker").Collection("users")
	err := collection.FindOne(ctx, bson.M{"email": email}).Err()
	return err == nil
}

func WebsiteExists(email string, website string) bool {
	collection := mongoClient.Database("downtimetracker").Collection("users")

	filter := bson.M{
		"email":    email,
		"websites": website,
	}

	err := collection.FindOne(ctx, filter).Err()
	return err == nil
}

func IsVerified(email string) bool {
	collection := mongoClient.Database("downtimetracker").Collection("users")

	var result bson.M
	err := collection.FindOne(ctx, bson.M{"email": email}).Decode(&result)
	if err != nil {
		return false
	}

	verified, ok := result["verified"].(bool)
	return ok && verified
}

func VerifyEmail(email string) error {
	collection := mongoClient.Database("downtimetracker").Collection("users")
	_, err := collection.UpdateOne(ctx, bson.M{"email": email}, bson.M{"$set": bson.M{"verified": true}})
	return err
}

func GetClient() *mongo.Client {
	return mongoClient
}

func GetAllUsersWithWebsites() ([]struct {
	Email    string   `bson:"email"`
	Websites []string `bson:"websites"`
}, error) {
	collection := mongoClient.Database("downtimetracker").Collection("users")
	ctx := context.TODO()
	cur, err := collection.Find(ctx, bson.M{"verified": true})
	if err != nil {
		return nil, err
	}
	defer func() { _ = cur.Close(ctx) }()
	var users []struct {
		Email    string   `bson:"email"`
		Websites []string `bson:"websites"`
	}
	for cur.Next(ctx) {
		var user struct {
			Email    string   `bson:"email"`
			Websites []string `bson:"websites"`
		}
		if err := cur.Decode(&user); err != nil {
			continue
		}
		users = append(users, user)
	}
	return users, nil
}
