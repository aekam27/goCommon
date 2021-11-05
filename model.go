package trestCommon

import (
	"context"
	"log"
	"os"
	firebase "firebase.google.com/go"
	"firebase.google.com/go/auth"
	"github.com/spf13/viper"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var database *mongo.Database
var fb *firebase.App

func init() {
	LoadConfig()
	db := viper.GetString("mongodb.db")
	dbHost := viper.GetString("mongodb.host")
	dbName := viper.GetString("mongodb.dbname")
	userName := viper.GetString("mongodb.username")
	password := viper.GetString("mongodb.password")
	uri := db + "://" + userName + ":" + password + "@" + dbHost + "/" + dbName + "?retryWrites=true&w=majority"
	clientOptions := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	err = client.Ping(context.Background(), nil)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	log.Println("Connected to MongoDB!")
	database = client.Database(dbName)
}

func AuthClient() (*auth.Client, error) {
	return fb.Auth(context.Background())
}

func FindOne(filter, projection bson.M, collectionName string) *mongo.SingleResult {
	return database.Collection(collectionName).FindOne(context.Background(), filter, options.FindOne().SetProjection(projection))
}
func Find(filter, projection bson.M, collectionName string) (*mongo.Cursor, error) {
	return database.Collection(collectionName).Find(context.Background(), filter, options.Find().SetProjection(projection))
}
func FindSort(filter, projection, filter1 bson.M, collectionName string) (*mongo.Cursor, error) {
	return database.Collection(collectionName).Find(context.Background(), filter, options.Find().SetProjection(projection), options.Find().SetSort(filter1))
}
func Aggregate(pipeline bson.A, collectionName string) (*mongo.Cursor, error) {
	return database.Collection(collectionName).Aggregate(context.Background(), pipeline)
}
func InsertOne(document interface{}, collectionName string) (*mongo.InsertOneResult, error) {
	return database.Collection(collectionName).InsertOne(context.Background(), document)
}
func UpdateOne(filter, update bson.M, collectionName string) (*mongo.UpdateResult, error) {
	return database.Collection(collectionName).UpdateOne(context.Background(), filter, update)
}
func DeleteOne(filter bson.M, collectionName string) (*mongo.DeleteResult, error) {
	return database.Collection(collectionName).DeleteOne(context.Background(), filter)
}
func DeleteMany(filter bson.M, collectionName string) (*mongo.DeleteResult, error) {
	return database.Collection(collectionName).DeleteMany(context.Background(), filter)
}
func FindWithLimitAndOffSet(filter, projection bson.M, limit, offset int, collectionName string) (*mongo.Cursor, error) {
	return database.Collection(collectionName).Find(context.Background(), filter, options.Find().SetProjection(projection), options.Find().SetLimit(int64(limit)), options.Find().SetSkip(int64(offset)))
}
func Count(filter, projection bson.M, collectionName string) (int64, error) {
	return database.Collection(collectionName).CountDocuments(context.Background(), filter)
}
