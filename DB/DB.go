package DB

import (
	"context"
	"fmt"
	"github.com/rob-johnston/plana/job"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"time"
)

type DB struct {
	name string
	collection string
	client *mongo.Client
}

func InitDB(name string, collection string, client *mongo.Client) *DB{
	db := new(DB)
	db.name = name
	db.collection = collection
	db.client = client

	return db
}


func (db *DB) CreateJob(job job.Job) error {
	collection := db.client.Database(db.name).Collection(db.collection)
	_, err := collection.InsertOne(context.Background(), job)
	if err != nil {
		return err
	}
	return nil
}


func (db *DB) UpdateJobById(id string, data interface{}) error {
	collection := db.client.Database(db.name).Collection(db.collection)
	filter := bson.D{{ "_id", primitive.ObjectIDFromHex(id)}}
	_, err := collection.UpdateOne(context.Background(), filter, data)
	if err != nil {
		return err
	}
	return nil
}

func (db *DB) FindJobs() []*job.Job {
	collection := db.client.Database("go-testing").Collection("jobs")

	var allResults []*job.Job

	query := bson.D{{
		"$and", []bson.D{
			{{"status", "waiting"}},
			{{"runAt", bson.D{{"$lte", time.Now()}}}},
		},
	}}

	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	cur, err := collection.Find(ctx, query)
	if err != nil {
		log.Fatal(err)
	}

	for cur.Next(context.Background()) {
		fmt.Println("decoding results...")
		var result job.Job
		err := cur.Decode(&result)

		if err != nil {
			log.Fatal(err)
		}
		allResults = append(allResults, &result)
	}

	err = cur.Close(context.Background())
	if err := cur.Err(); err != nil {
		log.Fatal(err)
	}

	return allResults
}