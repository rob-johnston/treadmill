package job

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

// represents a job to run
type Job struct {
	ID          	  primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
	Name string		  `bson:"name" json:"name"`
	Status string	  `bson:"status" json:"status"`
	RunAt time.Time   `bson:"runAt" json:"runAt"`
	Data interface{}  `bson:"data" json:"data"`
	Error interface{} `bson:"error" json:"error"`
}