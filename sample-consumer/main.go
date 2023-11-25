package main

import (
	consumer "github.com/ducthanh98/server-kit/kit/consumer"
	"github.com/ducthanh98/server-kit/sample-consumer/handlers"
)

// ALL HANDLER MUST BE REGISTERED HERE
func initHandler(w *consumer.Worker) {
	consumer.Bootstrapper.AddConsumerTask(&handlers.SampleHandler{Key: "sample"})
}

func main() {
	// For Bootstrapping consumer
	consumer.Bootstrapping("", initHandler)
}
