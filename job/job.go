package job

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

// represents a job to run
type Job struct {
	ID          	  		primitive.ObjectID `bson:"_id,omitempty" json:"_id"`
	Name string		  		`bson:"name,omitempty" json:"name"`
	Status string	  		`bson:"status,omitempty" json:"status"`
	RunAt time.Time   		`bson:"runAt,omitempty" json:"runAt"`
	CompletedAt time.Time	`bson:"completedAt,omitempty" json:"completedAt"`
	QueueIndex int			`bson:"queueIndex,omitempty"`
	Data interface{}  		`bson:"data,omitempty" json:"data"`
	Error interface{} 		`bson:"error,omitempty" json:"error"`
}