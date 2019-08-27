package options

import (
	"go.mongodb.org/mongo-driver/mongo"
	"time"
)

// represents the initial options when setting up our
type Options struct {
	WorkerCount int
	DateFormat string
	RunFrequency time.Duration
	Client *mongo.Client
	Database string
	Collection string
}
