package plana

import (
	"fmt"
	"github.com/rob-johnston/plana/DB"
	"github.com/rob-johnston/plana/job"
	"github.com/rob-johnston/plana/worker"
	"go.mongodb.org/mongo-driver/mongo"
	"time"
)

//  the heart of our plana - a scheduler
type Scheduler struct {
	definitions  map[string]func(interface{}) error
	pendingQueue []*job.Job
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
	var allResults []*job.Job

	allResults = s.db.FindJobs()
	fmt.Println("current state of value 1 in pending queue:::")
	fmt.Println(allResults[0])
	// copy all results to our pendingQueue
	s.pendingQueue = append(s.pendingQueue, allResults...)
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
	workerCount := 5

	i := 0
	for i < workerCount {
		i++
		fmt.Println("starting worker: ", i)
		worker := worker.Worker{
			ID: i,
			Channel: make(chan job.Job),
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


// adds a user defined job to our definitions map
func (s *Scheduler) Define(name string, function func(data interface{}) error){
	s.definitions[name] = function
}

// schedules a job to run
func (s * Scheduler) Schedule(when string, name string, data interface{}) error {

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
// TODO remove default function from here
func NewPlana(client *mongo.Client) *Scheduler {

	// set up our lovely Plana object (sorry about the name ¯\_(ツ)_/¯ )
	s := new(Scheduler)
	db := DB.InitDB("go-testing", "jobs", client)
	s.definitions = make(map[string]func(interface{}) error)
	s.pendingQueue = []*job.Job{}
	s.db = db
	return s
}

