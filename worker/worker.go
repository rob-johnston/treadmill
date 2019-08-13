package worker

import (
	"fmt"
	"time"
)

// represents a job to run
type Job struct {
	Name string
	RunAt time.Time
	Data interface{}
}

type Worker struct {
	ID int
	WorkerChannel chan chan Job
	Channel chan Job
	End chan struct{}
	Definitions  map[string]func(interface{})(string, error)
}

func (w *Worker) Start() {
	go func() {
		for{
			// when the worker is free, send this workers channel to the worker channel queue
			w.WorkerChannel <-w.Channel
			select {
			case job := <-w.Channel:
				// do the work here
				fmt.Println("executing super important job herr")
				fmt.Println(job)

				// execute the job function
				w.runJob(job)
				// use definitions to run function
			case <-w.End:
				return
			}
		}
	}()
}

func (w *Worker) runJob(job Job)  {
	res, err := w.Definitions[job.Name](job.Data)
	if err != nil {
		// handle a failed job
	}

	//handle a successfully completed job

	//ie access mongodb and update the record with
	//the results
}

func (w *Worker) Stop(){
	fmt.Printf("worker with id %d is stopping", w.ID)
	w.End <- struct{}{}
}