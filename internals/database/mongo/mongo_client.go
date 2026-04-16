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

func collection() *mongo.Collection {
	return mongoClient.Database("downtime").Collection("users")
}

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

func RegisterUser(email, passwordHash string) error {
	col := collection()
	if IsMail(email) {
		return context.Canceled // will return normal error
	}
	_, err := col.InsertOne(ctx, bson.M{
		"email":    email,
		"password": passwordHash,
		"verified": false,
		"websites": []string{},
	})
	return err
}

func GetUserAuth(email string) (string, bool, error) {
	col := collection()
	var result bson.M
	err := col.FindOne(ctx, bson.M{"email": email}).Decode(&result)
	if err != nil {
		return "", false, err
	}
	hash, _ := result["password"].(string)
	verified, _ := result["verified"].(bool)
	return hash, verified, nil
}

func AddWebsite(email string, website string) error {
	col := collection()
	filter := bson.M{"email": email}
	update := bson.M{
		"$addToSet": bson.M{
			"websites": website,
		},
	}
	res, err := col.UpdateOne(ctx, filter, update)
	if res.MatchedCount == 0 {
		return os.ErrNotExist
	}
	return err
}

func GetWebsites(email string) ([]string, error) {
	col := collection()

	var result struct {
		Websites []string `bson:"websites"`
	}

	err := col.FindOne(ctx, bson.M{"email": email}).Decode(&result)
	if err != nil {
		return nil, err
	}

	return result.Websites, nil
}

func RemoveWebsite(email string, website string) error {
	col := collection()

	filter := bson.M{"email": email}
	update := bson.M{
		"$pull": bson.M{
			"websites": website,
		},
	}

	_, err := col.UpdateOne(ctx, filter, update)
	return err
}

func IsMail(email string) bool {
	err := collection().FindOne(ctx, bson.M{"email": email}).Err()
	return err == nil
}

func WebsiteExists(email string, website string) bool {
	col := collection()

	filter := bson.M{
		"email":    email,
		"websites": website,
	}

	err := col.FindOne(ctx, filter).Err()
	return err == nil
}

func IsVerified(email string) bool {
	col := collection()

	var result bson.M
	err := col.FindOne(ctx, bson.M{"email": email}).Decode(&result)
	if err != nil {
		return false
	}

	verified, ok := result["verified"].(bool)
	return ok && verified
}

func VerifyEmail(email string) error {
	_, err := collection().UpdateOne(ctx, bson.M{"email": email}, bson.M{"$set": bson.M{"verified": true}})
	return err
}

func GetClient() *mongo.Client {
	return mongoClient
}

func GetAllUsersWithWebsites() ([]struct {
	Email    string   `bson:"email"`
	Websites []string `bson:"websites"`
}, error) {
	col := collection()
	ctx := context.TODO()
	cur, err := col.Find(ctx, bson.M{"verified": true})
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
