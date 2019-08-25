package worker

import (
	"fmt"
	"github.com/rob-johnston/plana/DB"
	"github.com/rob-johnston/plana/job"
	"time"
)

type Worker struct {
	ID int
	WorkerChannel chan chan job.Job
	Channel chan job.Job
	End chan struct{}
	Definitions  map[string]func(interface{}) error
	DB *DB.DB
}

func (w *Worker) Start() {
	go func() {
		for{
			// when the worker is free, send this workers channel to the worker channel queue
			w.WorkerChannel <-w.Channel
			select {
			case j := <-w.Channel:

				// execute the job function
				w.runJob(&j)
				// use definitions to run function
			case <-w.End:
				return
			}
		}
	}()
}

func (w *Worker) runJob(j *job.Job)  {
	err := w.Definitions[j.Name](j.Data)

	payload := job.Job{}
	if err != nil {

		// handle a failed job
		payload.Status = "failed"
		payload.Error = err
		payload.CompletedAt = time.Now()

	} else {

		// handle successfully run job
		payload.Status = "completed"
		payload.CompletedAt = time.Now()
	}


	err = w.DB.UpdateJobByObjectId(j.ID, payload)

	if err != nil {
		fmt.Println(err)
	}
}

func (w *Worker) Stop(){
	fmt.Printf("worker with id %d is stopping", w.ID)
	w.End <- struct{}{}
}