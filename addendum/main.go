package main

import (
	"context"
	"log"
	"os"
	"runtime/trace"
	"time"
)

// 下記を実行して、「go tool trace trace.out 」で確認できる。便利〜。
func main() {
	f, err := os.Create("trace.out")
	if err != nil {
		log.Fatalf("failed to create trace output file: %v", err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			log.Fatalf("failed to close trace output file: %v", err)
		}
	}()

	if err := trace.Start(f); err != nil {
		panic(err)
	}
	defer trace.Stop()

	ctx := context.Background()
	ctx, task := trace.NewTask(ctx, "makeCoffee")
	defer task.End()
	trace.Log(ctx, "orderID", "1")

	coffee := make(chan bool)

	go func() {
		trace.WithRegion(ctx, "extractCoffee", func() {
			time.Sleep(5 * time.Second)
			coffee <- true
		})
	}()
	<-coffee
}
