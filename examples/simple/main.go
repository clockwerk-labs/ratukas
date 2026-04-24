package main

import (
	"context"
	"log"
	"log/slog"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/clockwerk-labs/ratukas"
)

func main() {
	log.Println("Running example simulation")

	expiry := make(chan *ratukas.Bucket)
	defer close(expiry)

	engine := ratukas.NewEngine(
		ratukas.NewTimingWheel(time.Now(), 100*time.Millisecond, 16192, expiry),
		ratukas.NewRegistry(1024),
		slog.New(slog.NewJSONHandler(os.Stdout, nil)),
		expiry,
	)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	go engine.Run(ctx)

	for i := 0; i < 100_000_000; i++ {
		now := time.Now()
		expiration := now.Add(time.Duration(rand.Intn(91)+10) * time.Second)

		engine.AddTask(uint64(i)+1, ratukas.NewTask(expiration, func() {
			log.Printf("Running task %d, expiration %d, delay: %d", i+1, expiration.UnixMilli(), time.Now().Sub(now).Milliseconds())
		}))
	}

	log.Println("All tasks added")

	<-ctx.Done()
}
