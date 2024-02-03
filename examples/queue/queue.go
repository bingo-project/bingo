package main

import (
	"log"

	"github.com/duke-git/lancet/v2/convertor"
	"github.com/hibiken/asynq"

	"bingo/internal/apiserver/job"
)

var (
	Queue *asynq.Client
)

func init() {
	Queue = asynq.NewClient(asynq.RedisClientOpt{Addr: ":6379", DB: 1})
}

func main() {
	// Enqueue task
	payload := &job.EmailTaskPayload{
		Username: "Peter",
		Email:    "peter@gmail.com",
	}

	t := asynq.NewTask("demo:task", []byte(convertor.ToString(payload)))
	info, err := Queue.Enqueue(t)
	if err != nil {
		log.Println(err)

		return
	}

	log.Println("demo:task enqueued", convertor.ToString(info))
}
