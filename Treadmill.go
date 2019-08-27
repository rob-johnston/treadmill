package treadmill

import (
	"fmt"
	"github.com/rob-johnston/treadmill/DB"
	"github.com/rob-johnston/treadmill/job"
	"github.com/rob-johnston/treadmill/options"
	"github.com/rob-johnston/treadmill/queue"
	"github.com/rob-johnston/treadmill/worker"
	"go.mongodb.org/mongo-driver/bson"
	"log"
	"time"
)

//  this is the heart of our treadmill object
type Scheduler struct {
	definitions  map[string]func(interface{}) error
	pendingQueue *queue.PendingQueue
	workers      []*worker.Worker
	db *DB.DB
	options options.Options
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
		for range time.Tick(s.options.RunFrequency) {

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
		s.enqueue(v)
	}
}

func (s *Scheduler) enqueue(j *job.Job) {
	s.pendingQueue.Append(j)
	err := s.db.UpdateJobByObjectId(j.ID, struct{ Status string }{ Status: "running" })
	if err != nil {
		fmt.Printf("unable to set job: %v status to running\n", j.ID)
	}
}


// listens on channel for messages when to process jobs
func (s *Scheduler) dispatcherWorker(run <-chan struct{}, close <-chan struct{}) {

	// initialise some workers
	i := 0
	for i < s.options.WorkerCount {
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
	t, err := time.Parse(s.options.DateFormat, when)
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

func NewTreadmill(options options.Options) *Scheduler {

	checkOptions(&options)
	// set up our treadmill object thing ¯\_(ツ)_/¯
	s := new(Scheduler)
	db := DB.InitDB(options.Database, options.Collection, options.Client)
	s.definitions = make(map[string]func(interface{}) error)
	s.pendingQueue = queue.NewPendingQueue()
	s.db = db
	s.options = options
	return s
}

// set some default values if they aren't included
func checkOptions(options *options.Options) {
	if options.Client == nil {
		panic("You must provide a mongo client in the options")
	}

	if options.Collection == "" || options.Database == "" {
		panic("You must provide database and collection names")
	}

	if options.WorkerCount == 0 {
		options.WorkerCount = 5
	}

	if options.DateFormat == "" {
		options.DateFormat = "2006-01-02 15:04:05"
	}

	if options.RunFrequency == 0 {
		options.RunFrequency = time.Second * 5
	}
}
