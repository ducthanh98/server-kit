package utils

import (
	"fmt"
	"sync"
)

type Worker struct {
	messages    chan Job
	waitGroup   *sync.WaitGroup
	workerCount int
}

type Job struct {
	Message interface{}
	Handler func(interface{})
}

func NewWorker(workerCount int, q chan Job, wg *sync.WaitGroup) *Worker {
	return &Worker{
		messages:    q,
		waitGroup:   wg,
		workerCount: workerCount,
	}
}

func (w *Worker) Init() {
	for i := 0; i < w.workerCount; i++ {
		go w.startWorker()
	}
}

func (w *Worker) QueueLength() int {
	return len(w.messages)
}

func (w *Worker) startWorker() {
	for {
		select {
		case param, ok := <-w.messages:
			{
				if !ok {
					fmt.Println("Batch worker is stopping...")
					return
				}
				param.Handler(param.Message)
			}

		}
	}
}

func (w *Worker) Process(job Job) {
	//comment out to reduce log noise
	//fmt.Printf("Producing with parameter: %d\n", param)
	w.waitGroup.Add(1)
	w.messages <- job
}

func (w *Worker) Stop() {
	close(w.messages)
}
