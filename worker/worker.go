package worker

import (
	"fmt"
	"github.com/rob-johnston/plana/DB"
	"github.com/rob-johnston/plana/job"
)

type Worker struct {
	ID int
	WorkerChannel chan chan job.Job
	Channel chan job.Job
	End chan struct{}
	Definitions  map[string]func(interface{}) error
	DB.DB
}

func (w *Worker) Start() {
	go func() {
		for{
			// when the worker is free, send this workers channel to the worker channel queue
			w.WorkerChannel <-w.Channel
			select {
			case job := <-w.Channel:

				// execute the job function
				w.runJob(&job)
				// use definitions to run function
			case <-w.End:
				return
			}
		}
	}()
}

func (w *Worker) runJob(job *job.Job)  {
	err := w.Definitions[job.Name](job.Data)
	if err != nil {
		// handle a failed job
		job.Status = "failed"
		job.Error = err
	}

	// TODO handle a completed job, update result to DB

	//handle a successfully completed job

	//ie access mongodb and update the record with
	//the results
}

func (w *Worker) Stop(){
	fmt.Printf("worker with id %d is stopping", w.ID)
	w.End <- struct{}{}
}