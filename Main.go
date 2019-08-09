package plana

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"time"
)

type Job struct {
	Name string
	RunAt time.Time
	Data string
}

// define plana struct (needs a DB)
type Scheduler struct {
	db *mongo.Client
	definitions map[string]func(string)
}

// runs the plana
func (s Scheduler) run() {

	for range time.Tick(10 * time.Second) {
		collection := s.db.Database("go-testing").Collection("jobs")

		var allResults []*Job

		ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
		cur, err := collection.Find(ctx, bson.D{})
		if err != nil {
			log.Fatal(err)
		}

		for cur.Next(context.Background()) {
			fmt.Println("decoding results...")
			var result Job
			err := cur.Decode(&result)

			if err != nil {
				log.Fatal(err)
			}

			allResults = append(allResults, &result)
			fmt.Println(result)
		}

		for _, element := range allResults {
			job := *element
			s.definitions[job.Name](job.Data)
		}

		err = cur.Close(context.Background())
		if err := cur.Err(); err != nil {
			log.Fatal(err)
		}
	}
	// every so often, process events
}

// gets all jobs
func (s Scheduler) processJobs() {
	// search DB for all relevant jobs

	// for each result decode into a job struct

	// process each job
}

// handle a single job
func (s Scheduler) processJob(job Job) {

}

func (s Scheduler) scheduleJob() {

}

func (s Scheduler) Testing() {
	fmt.Println("hey hey shitbirds")
}

// my first job
//func robsJob() {
//
//}

func NewPlana(db *mongo.Client) *Scheduler {

	s := new(Scheduler)
	s.db = db
	s.definitions = map[string]func(string){
		"test": func(s string){
			fmt.Println(s)
		},
	}
	return s
}

