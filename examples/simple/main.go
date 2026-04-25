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

type Callback func() error

func main() {
	log.Println("Running example simulation")

	out := make(chan Callback)
	defer close(out)

	go func() {
		for o := range out {
			if err := o(); err != nil {
				log.Println(err)
			}
		}
	}()

	engine := ratukas.NewEngine(
		time.Now(),
		100*time.Millisecond,
		16192,
		ratukas.NewRegistry[Callback](1024),
		slog.New(slog.NewJSONHandler(os.Stdout, nil)),
		out,
	)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	go engine.Run(ctx)

	for i := 0; i < 1_000_000; i++ {
		now := time.Now()
		expiration := now.Add(time.Duration(rand.Intn(31)+10) * time.Second)

		engine.AddTask(uint64(i)+1, ratukas.NewTask[Callback](expiration, func() error {
			log.Printf("Running task %d, expiration %s, delay: %d", i+1, expiration.String(), time.Now().Sub(now).Milliseconds())

			return nil
		}))
	}

	<-ctx.Done()
}
