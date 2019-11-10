package main

import (
	"context"
	"fmt"
	"github.com/dustin/go-humanize"
	"os"
	"os/signal"
	"syscall"
)

func printHumanize() {
	name := os.Args[1]
	s, _ := os.Stat(name)
	fmt.Printf("%s: %s\n", name, humanize.Bytes(uint64(s.Size())))
}

func mainStudy() {
	defer fmt.Println("done")
	trapSignals := []os.Signal{
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	}
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, trapSignals...)

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		sig := <-sigCh
		fmt.Println("Got signal", sig)
		cancel()
	}()
	doSomething(ctx)
}

func doSomething(context context.Context) {
	defer fmt.Println("done doSomething")
	for {
		select {
		case <-context.Done():
			return
		default:
		}
		printHumanize()
	}
}

// CLIを作る時のサードパーティ製ツール
// https://github.com/mitchellh/cli
// https://github.com/urfave/cli