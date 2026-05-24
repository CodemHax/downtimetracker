package mongo

import (
	"context"
	"errors"
	"os"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var mongoClient *mongo.Client

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
	ctx := context.Background()
	col := collection()
	if IsMail(email) {
		if IsVerified(email) {
			return errors.New("email is already registered")
		}
		_, err := col.DeleteOne(ctx, bson.M{"email": email, "verified": false})
		if err != nil {
			return err
		}
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
	ctx := context.Background()
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
	ctx := context.Background()
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
	ctx := context.Background()
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
	ctx := context.Background()
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
	err := collection().FindOne(context.Background(), bson.M{"email": email}).Err()
	return err == nil
}

func WebsiteExists(email string, website string) bool {
	col := collection()

	filter := bson.M{
		"email":    email,
		"websites": website,
	}

	err := col.FindOne(context.Background(), filter).Err()
	return err == nil
}

func IsVerified(email string) bool {
	col := collection()

	var result bson.M
	err := col.FindOne(context.Background(), bson.M{"email": email}).Decode(&result)
	if err != nil {
		return false
	}

	verified, ok := result["verified"].(bool)
	return ok && verified
}

func VerifyEmail(email string) error {
	_, err := collection().UpdateOne(context.Background(), bson.M{"email": email}, bson.M{"$set": bson.M{"verified": true}})
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
	ctx := context.Background()
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

func UpdatePassword(email, passwordHash string) error {
	col := collection()
	_, err := col.UpdateOne(context.Background(), bson.M{"email": email}, bson.M{"$set": bson.M{"password": passwordHash}})
	return err
}
