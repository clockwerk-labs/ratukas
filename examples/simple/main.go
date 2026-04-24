package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"
	"time"

	"github.com/clockwerk-labs/ratukas"
)

func main() {
	log.Println("Running example simulation")

	task1 := ratukas.NewTask(1, time.Now().Add(5*time.Second), func() {
		log.Println("Running task 1")
	})
	task2 := ratukas.NewTask(2, time.Now().Add(15*time.Second), func() {
		log.Println("Running task 2")
	})

	registry := make(map[uint64]*ratukas.Task)
	registry[task1.ID()] = task1
	registry[task2.ID()] = task2

	expiry := make(chan *ratukas.Bucket)
	defer close(expiry)

	wheel := ratukas.NewTimingWheel(time.Now(), 100*time.Millisecond, 60, expiry)

	execution := make(chan uint64)
	defer close(execution)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	engine := ratukas.NewEngine(wheel, expiry, execution)
	go engine.Start(ctx)

	wheel.Add(task1)
	wheel.Add(task2)

	go func() {
		for id := range execution {
			if task, ok := registry[id]; ok {
				task.Execute()
			}
		}
	}()

	<-ctx.Done()
}
