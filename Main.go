package plana

import (
	"context"
	"fmt"
	"github.com/rob-johnston/plana/worker"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"time"
)

//  the heart of our plana - a scheduler
type Scheduler struct {
	db           *mongo.Client
	definitions  map[string]func(interface{})(string, error)
	pendingQueue []*worker.Job
	workers      []*worker.Worker
}

var WorkerChannel = make(chan chan worker.Job)

// runs the scheduler
func (s *Scheduler) Run() {

	// channels for processing jobs
	processJobs := make(chan struct {}, 100)
	end := make(chan struct{})

	// create a dispatch worker
	go s.dispatcherWorker(processJobs, end)

	// create job workers

	// start thread with our main job fetching loop
	go func(){
		for range time.Tick(5 * time.Second) {

			// pull in all jobs
			s.fetchDbJobs(processJobs)

			// signal it is time for the dispatch worker to run
			processJobs <- struct{}{}
		}
	}()
}

// fetches jobs from the DB and places them in the pendingQueue
func (s *Scheduler) fetchDbJobs(processJobs chan struct{}) {
	collection := s.db.Database("go-testing").Collection("jobs")

	var allResults []*worker.Job

	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	cur, err := collection.Find(ctx, bson.D{})
	if err != nil {
		log.Fatal(err)
	}

	for cur.Next(context.Background()) {
		fmt.Println("decoding results...")
		var result worker.Job
		err := cur.Decode(&result)

		if err != nil {
			log.Fatal(err)
		}
		allResults = append(allResults, &result)
	}
	// copy all results to our pendingQueue
	s.pendingQueue = append(s.pendingQueue, allResults...)
	fmt.Println(s.pendingQueue)

	err = cur.Close(context.Background())
	if err := cur.Err(); err != nil {
		log.Fatal(err)
	}
}

func enqueueJob (queue []*worker.Job, job worker.Job) []*worker.Job {
	// throw out any duplicate jobs

	// could do some prioritization pendingQueue stuff here
	return append(queue, &job)
}


// listens on channel for messages when to process jobs
func (s *Scheduler) dispatcherWorker(run <-chan struct{}, close <-chan struct{}) {

	// initialise some workers
	workerCount := 5

	i := 0
	for i < workerCount {
		i++
		fmt.Println("starting worker: ", i)
		worker := worker.Worker{
			ID: i,
			Channel: make(chan worker.Job),
			WorkerChannel: WorkerChannel,
			End: make(chan struct{}),
			Definitions: s.definitions,
		}

		worker.Start()
		s.workers = append(s.workers, &worker)
	}


	for {
		// loop and listen for signal to run or close
		select {
		case <-run:
			fmt.Println(s.pendingQueue)
			fmt.Println("running inside dispatch worker loop")
			for _, job := range s.pendingQueue {
				fmt.Println("about to dispatch job to worker..")
				worker := <-WorkerChannel // wait for channel to be available
				worker <- *job
			}
		case <-close:
			return
		}
	}
}

func (s *Scheduler) ProcessJob(job *worker.Job) {
	if f, ok := s.definitions[job.Name]; ok {
		res, err := f(job.Data)
		if err != nil {
			log.Fatal("failed job")
		}
		fmt.Println(res)
		fmt.Println("at end of processJob")
	}
}

// adds a user defined job to our definitions map
func (s *Scheduler) define(name string, function func(interface{})(string, error)){
	s.definitions[name] = function
}

// schedules a job to run
func (s * Scheduler) schedule(when string, name string, data interface{}) error {

	t, err := time.Parse(time.UnixDate, when)
	if err != nil { // Always check errors even if they should not happen.
		panic(err)
	}

	job := worker.Job{
		Name:name,
		RunAt:t,
		Data:data,
	}

	collection := s.db.Database("go-testing").Collection("jobs")
	_, err = collection.InsertOne(context.Background(), job)
	if err != nil {
		return err
	}
	return nil
}

func NewPlana(db *mongo.Client) *Scheduler {

	// set up our lovely Plana object (sorry about the name ¯\_(ツ)_/¯ )
	s := new(Scheduler)
	s.db = db
	s.definitions = map[string]func(interface{})(string, error){
		"test": func(data interface{})(string, error){
			fmt.Println(data)
			fmt.Println("successfully running the job")
			return "", nil
		},
	}
	s.pendingQueue = []*worker.Job{}
	return s
}

