package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/adjust/rmq"
	"github.com/golang/glog"
	"gopkg.in/redis.v3"
)

var (
	redisURL, redisPort, redisPassword, redisTopic string
	redisProtocol                                  = "tcp"
	redisNetworkTag                                = "transcode_task_consume"

	// TODO: Make below variables as a CLI argument
	queueFetchInterval            = 10
	queueFetchIntervalMeasurement = time.Second
)

func main() {
	loadEnvironmentVariables()

	taskQueue := openTaskQueue()
	glog.Infof("Queue accessed: %s\n", redisTopic)

	taskQueue.StartConsuming(queueFetchInterval, queueFetchIntervalMeasurement)
	glog.Infoln("Queue consumption started...")

	taskConsumer := &TaskConsumer{}
	taskQueue.AddConsumer("Task consumer 1", taskConsumer)

	select {}
}

// TaskConsumer represents the Redis topic consumer
type TaskConsumer struct {
	name         string
	count        int
	lastAccessed time.Time
}

// Consume method implements TaskConsumer struct
// to be registered on Queue.
// It handles actual data handling from Queue
func (tc *TaskConsumer) Consume(delivery rmq.Delivery) {
	var task Task

	tc.count++

	if err := json.Unmarshal([]byte(delivery.Payload()), &task); err != nil {
		glog.Errorf("Failed to read task message: %s\n", err)
		delivery.Reject()
		return
	}

	glog.Infof("Processed task message: Transcoding %s\n", task.FilePath)
	// TODO: Call Go subroutine to call go binding of ffmpeg
	delivery.Ack()
}

// loadEnvironmentVariables loads Redis
// information from environment variables
func loadEnvironmentVariables() {
	redisURL = os.Getenv("REDIS_URL")
	if len(redisURL) == 0 {
		panic("No REDIS_URL environment variable")
	}

	redisPort = os.Getenv("REDIS_PORT")
	if len(redisPort) == 0 {
		panic("No REDIS_PORT environment variable")
	}

	redisTopic = os.Getenv("REDIS_TOPIC")
	if len(redisTopic) == 0 {
		panic("No REDIS_TOPIC environment variable")
	}

	redisPassword = os.Getenv("REDIS_PASSWORD")
	if len(redisPassword) == 0 {
		panic("No REDIS_PASSWORD environment variable")
	}

	redisTopic = os.Getenv("REDIS_TOPIC")
	if len(redisTopic) == 0 {
		panic("No REDIS_TOPIC environment variable")
	}
}

// openTaskQueue connects to redis and return a Queue interface
func openTaskQueue() rmq.Queue {
	redisClient := redis.NewClient(&redis.Options{
		Network:  redisProtocol,
		Addr:     fmt.Sprintf("%s:%s", redisURL, redisPort),
		DB:       int64(1),
		Password: redisPassword,
	})

	connection := rmq.OpenConnectionWithRedisClient(redisNetworkTag, redisClient)

	glog.Infof("Connected to Redis task queue: %s\n", connection.Name)

	return connection.OpenQueue(redisTopic)
}
