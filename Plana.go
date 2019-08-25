package plana

import (
	"fmt"
	"github.com/rob-johnston/plana/DB"
	"github.com/rob-johnston/plana/job"
	"github.com/rob-johnston/plana/queue"
	"github.com/rob-johnston/plana/worker"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"time"
)

//  the heart of our plana - a scheduler
type Scheduler struct {
	definitions  map[string]func(interface{}) error
	pendingQueue *queue.PendingQueue
	workers      []*worker.Worker
	db *DB.DB
}

var WorkerChannel = make(chan chan job.Job)


// runs the scheduler
func (s *Scheduler) Run() {

	// channels for processing jobs
	processJobs := make(chan struct {}, 100)
	end := make(chan struct{})

	// create a dispatch worker
	go s.dispatcherWorker(processJobs, end)

	// start thread with our main job fetching loop
	go func(){
		for range time.Tick(5 * time.Second) {

			// pull in all jobs
			s.fetchDbJobs(processJobs)

			if s.pendingQueue.Length() > 0 {
				// signal it is time for the dispatch worker to run
				processJobs <- struct{}{}
			}
		}
	}()
}

// fetches jobs from the DB and places them in the pendingQueue
func (s *Scheduler) fetchDbJobs(processJobs chan struct{}) {
	var allResults []*job.Job

	// find all jobs
	allResults = s.db.FindJobs()

	// copy them to our pendingQueue
	for _, v := range allResults {
		s.pendingQueue.Append(v)
	}
}

func enqueueJob (queue []*job.Job, job job.Job) []*job.Job {
	// throw out any duplicate jobs
	// TODO if queue already contains a job, don't add it

	// could do some prioritization pendingQueue stuff here
	return append(queue, &job)
}


// listens on channel for messages when to process jobs
func (s *Scheduler) dispatcherWorker(run <-chan struct{}, close <-chan struct{}) {

	// initialise some workers
	// TODO take worker count as argument?
	workerCount := 10

	i := 0
	for i < workerCount {
		i++
		fmt.Println("starting worker: ", i)
		w := worker.Worker{
			ID: i,
			Channel: make(chan job.Job),
			WorkerChannel: WorkerChannel,
			End: make(chan struct{}),
			Definitions: s.definitions,
			DB: s.db,
		}

		w.Start()
		s.workers = append(s.workers, &w)
	}


	for {
		// loop and listen for signal to run or close
		select {
		case <-run:
			for j := range s.pendingQueue.Iterate() {
				wc := <-WorkerChannel // wait for channel to be available
				wc <- *j // send work to the available channel
			}

			// now clear the queue?


		case <-close:
			for _, wrkr := range s.workers {
				wrkr.Stop()
			}
		}
	}
}

// decodes given bson data into a given struct
func (s *Scheduler) Decode(goal interface{}, data interface{}) error {

	x, err := bson.Marshal(data)
	if err != nil {
		log.Fatal(err)
	}

	err = bson.Unmarshal(x, goal)
	if err != nil {
		log.Fatal(err)
	}

	return nil
}


// adds a user defined job to our definitions map
func (s *Scheduler) Define(name string, function func(data interface{}) error){
	s.definitions[name] = function
}

// schedules a job to run
func (s * Scheduler) Schedule(when string, name string, data interface{}) error {

	// define how we accept dates - could also be a setting??
	layout := "2006-01-02 15:04:05"
	t, err := time.Parse(layout, when)
	if err != nil {
		fmt.Println("error parsing time")
		return err
	}

	payload := job.Job{
		Name:name,
		RunAt:t,
		Data:data,
		Status: "waiting",
	}

	err = s.db.CreateJob(payload)
	if err != nil {
		fmt.Println("error creating job in db")
		return err
	}
	return nil
}

// TODO rename this planner thing - just use Scheduler everywhere
// TODO take options struct as argument, workers, etc
func NewPlana(client *mongo.Client) *Scheduler {

	// set up our lovely Plana object (sorry about the name ¯\_(ツ)_/¯ )
	s := new(Scheduler)
	db := DB.InitDB("go-testing", "jobs", client)
	s.definitions = make(map[string]func(interface{}) error)
	s.pendingQueue = queue.NewPendingQueue()
	s.db = db
	return s
}

