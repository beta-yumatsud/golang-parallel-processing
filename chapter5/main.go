package main

import (
	"context"
	"fmt"
	"golang.org/x/time/rate"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"runtime/debug"
	"sort"
	"sync"
	"time"
)

func main() {
	log.SetOutput(os.Stdout)
	log.SetFlags(log.Ltime|log.LUTC)

	err := runJob("1")
	if err != nil {
		msg := "There was an unexpected issue; please report this as a bug."
		if _, ok := err.(IntermediateErr); ok {
			msg = err.Error()
		}
		handleError(1, err, msg)
	}

	sample1()
	sample2()
	sample3()
	sample4()
	sample5()
}

type MyError struct {
	Inner error
	Message string
	StackTrace string
	Misc map[string]interface{}
}

func wrapError(err error, messagef string, msgArgs ...interface{}) MyError {
	return MyError{
		Inner:   err,
		Message: fmt.Sprintf(messagef, msgArgs...),
		StackTrace: string(debug.Stack()),
		Misc: make(map[string]interface{}),
	}
}

func (err MyError) Error() string {
	return err.Message
}

type LowLevelErr struct {
	error
}

func isGloballyExec(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		return false, LowLevelErr{wrapError(err, err.Error())}
	}
	return info.Mode().Perm()&0100 == 0100, nil
}

type IntermediateErr struct {
	error
}

func runJob(id string) error {
	const jobBinPath = "/bad/job/binary"
	isExecutable, err := isGloballyExec(jobBinPath)
	if err != nil {
		return IntermediateErr{wrapError(
			err,
			"cannot run job %q: requisite binaries not availabl",
			id,
		)}
	} else  if isExecutable == false {
		return wrapError(nil, "job binary is not executable")
	}

	return exec.Command(jobBinPath, "--id="+id).Run()
}

func handleError(key int, err error, message string) {
	log.SetPrefix(fmt.Sprintf("[logID: %v]: ", key))
	log.Printf("%#v", err)
	fmt.Printf("[%v] %v", key, message)
}

func sample1() {
	doWork := func(
		done <-chan interface{},
		pulseInterval time.Duration,
	) (<-chan interface{}, <-chan time.Time) {
		heartBeat := make(chan interface{})
		results := make(chan time.Time)

		go func() {
			defer close(heartBeat)
			defer close(results)

			pulse := time.Tick(pulseInterval)
			workGen := time.Tick(2 * pulseInterval)

			sendPulse := func() {
				select {
				case heartBeat <- struct {}{}:
				default:
				}
			}

			sendResult := func(r time.Time) {
				for {
					select {
					case <-done:
						return
					case <-pulse:
						sendPulse()
					case results <- r:
						return
					}
				}
			}

			for {
				select {
				case <-done:
					return
				case <-pulse:
					sendPulse()
				case r := <-workGen:
					sendResult(r)
				}
			}
		}()
		return heartBeat, results
	}

	done := make(chan interface{})
	time.AfterFunc(10 * time.Second, func() {
		close(done)
	})

	const timeout = 2 * time.Second

	heartBeat, result := doWork(done, timeout/2)
	for {
		select {
		case _, ok := <-heartBeat:
			if !ok {
				return
			}
			fmt.Println("pulse")
		case r, ok := <-result:
			if ok == false {
				return
			}
			fmt.Printf("results %v\n", r.Second())
		case <-time.After(timeout):
			return
		}
	}

}

func sample2() {
	doWork := func(done <-chan interface{}) (<-chan interface{}, <-chan int) {
		heartbeatStream := make(chan interface{}, 1)
		workStream := make(chan int)

		go func() {
			defer close(heartbeatStream)
			defer close(workStream)

			for i := 0; i < 10; i++ {
				select {
				case heartbeatStream <- struct {}{}:
				default:
				}

				select {
				case <-done:
					return
				case workStream <- rand.Intn(10):
				}
			}
		}()

		return heartbeatStream, workStream
	}

	done := make(chan interface{})
	defer close(done)

	heartbeat, results := doWork(done)
	for {
		select {
		case _, ok := <-heartbeat:
			if ok {
				fmt.Println("pulse")
			} else {
				return
			}
		case r, ok := <-results:
			if ok {
				fmt.Printf("results %v\n", r)
			} else {
				return
			}
		}
	}
}

func DoWork(
	done <-chan interface{},
	pulseInterval time.Duration,
	nums ...int,
) (<-chan interface{}, <-chan int) {
	heartbeat := make(chan interface{}, 1)
	intStream := make(chan int)
	go func() {
		defer close(heartbeat)
		defer close(intStream)
		time.Sleep(2*time.Second)

		pulse := time.Tick(pulseInterval)

		numLoop:
		for _, n := range nums {
			for {
				select {
				case <-done:
					return
				case <-pulse:
					select {
					case heartbeat <- struct{}{}:
					default:
					}
				case intStream <- n:
					continue numLoop
				}
			}
		}
	}()
	return heartbeat, intStream
}

func sample3() {
	doWork := func(
		done <-chan interface{},
		id int,
		wg *sync.WaitGroup,
		result chan<- int,
	){
		started := time.Now()
		defer wg.Done()

		// Simulate random load
		simulatedLoadTime := time.Duration(1+rand.Intn(5))*time.Second

		select {
		case <-done:
		case <-time.After(simulatedLoadTime):
		}


		select {
		case <-done:
		case result <- id:
		}

		took := time.Since(started)
		// Display how long handlers would have taken
		if took < simulatedLoadTime {
			took = simulatedLoadTime
		}
		fmt.Printf("%v took %v\n", id, took)
	}

	done := make(chan interface{})
	result := make(chan int)

	var wg sync.WaitGroup
	wg.Add(10)
	for i:=0; i < 10; i++ {
		go doWork(done, i, &wg, result)
	}

	firstReturned := <-result
	close(done)
	wg.Wait()

	fmt.Printf("Received an answer from #%v\n", firstReturned)
}

type APIConnection struct {
	//rateLimiter *rate.Limiter
	rateLimiter RateLimiter
	// これをnetwork、disk、apiなどに分けて設定することま可能
}

func Open() *APIConnection {
	secondLimit := rate.NewLimiter(Per(2, time.Second), 1)
	minuteLimit := rate.NewLimiter(Per(10, time.Minute), 10)
	return &APIConnection{
		// 下記だと1秒に一回
		//rateLimiter: rate.NewLimiter(rate.Limit(1), 1),
		rateLimiter: MultiLimiter(secondLimit, minuteLimit),
	}
}

func (a *APIConnection) ReadFile(ctx context.Context) error {
	if err := a.rateLimiter.Wait(ctx); err != nil {
		return err
	}
	return nil
}

func (a *APIConnection) ResolveAddress(ctx context.Context) error {
	if err := a.rateLimiter.Wait(ctx); err != nil {
		return err
	}
	return nil
}

type RateLimiter interface {
	Wait(context.Context) error
	Limit() rate.Limit
}

func MultiLimiter(limiters ...RateLimiter) *multiLimiter {
	byLimit := func(i, j int) bool {
		return limiters[i].Limit() < limiters[j].Limit()
	}
	sort.Slice(limiters, byLimit)
	return &multiLimiter{limiters: limiters}
}

type multiLimiter struct {
	limiters []RateLimiter
}

func (l *multiLimiter) Wait(ctx context.Context) error {
	for _ ,l := range l.limiters {
		if err := l.Wait(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (l *multiLimiter) Limit() rate.Limit {
	return l.limiters[0].Limit()
}

func Per(eventCount int, duration time.Duration) rate.Limit {
	return rate.Every(duration/time.Duration(eventCount))
}

func sample4() {
	defer fmt.Println("Done.")
	log.SetOutput(os.Stdout)
	log.SetFlags(log.Ltime | log.LUTC)

	apiConnection := Open()
	var wg sync.WaitGroup
	wg.Add(20)

	for i := 0; i < 10; i++ {
		go func() {
			defer wg.Done()
			err := apiConnection.ReadFile(context.Background())
			if err != nil {
				log.Printf("cannot ReadFile: %v", err)
			}
			log.Printf("ReadFile")
		}()
	}

	for i := 0; i < 10; i++ {
		go func() {
			defer wg.Done()
			err := apiConnection.ResolveAddress(context.Background())
			if err != nil {
				log.Printf("cannot ResolveAddress: %v", err)
			}
			log.Printf("ResolveAddress")
		}()
	}

	wg.Wait()
}

func sample5() {
	var or func(channels ...<-chan interface{}) <-chan interface{}
	or = func(channels ...<-chan interface{}) <-chan interface{} {
		switch len(channels) {
		case 0:
			return nil
		case 1:
			return channels[0]
		}
		orDone := make(chan interface{})
		go func() { // ❹
			defer close(orDone)
			switch len(channels) {
			case 2:
				select {
				case <-channels[0]:
				case <-channels[1]:
				}
			default:
				select {
				case <-channels[0]:
				case <-channels[1]:
				case <-channels[2]:
				case <-or(append(channels[3:], orDone)...):
				}
			}
		}()
		return orDone
	}
	type startGoroutineFn func(
		done <-chan interface{},
		pulseInterval time.Duration,
	) (heartbeat <-chan interface{})

	newSteward := func(
		timeout time.Duration,
		startGoroutine startGoroutineFn,
	) startGoroutineFn {
		return func(done <-chan interface{}, pulseInterval time.Duration) (<-chan interface{}) {
			heartbeat := make(chan interface{})
			go func() {
				defer close(heartbeat)

				var wardDone chan interface{}
				var wardHeartbeat <-chan interface{}
				startWard := func() {
					wardDone = make(chan interface{})
					wardHeartbeat = startGoroutine(or(wardDone, done), timeout/2)
				}
				startWard()
				pulse := time.Tick(pulseInterval)

			monitorLoop:
				for {
					timeoutSignal := time.After(timeout)

					for {
						select {
						case <-pulse:
							select {
							case heartbeat <- struct {}{}:
							default:
							}
						case <-wardHeartbeat:
							continue monitorLoop
						case <-timeoutSignal:
							log.Println("steward: ward unhealthy; restarting")
							close(wardDone)
							startWard()
							continue monitorLoop
						case <-done:
							return
						}
					}
				}
			}()
			return heartbeat
		}
	}

	log.SetOutput(os.Stdout)
	log.SetFlags(log.Ltime | log.LUTC)

	doWork := func(done <-chan interface{}, _ time.Duration) <-chan interface{} {
		log.Println("ward: Hello, I'm irresponsible!")
		go func() {
			<-done
			log.Println("ward: I am halting.")
		}()
		return nil
	}
	doWorkWithSteward := newSteward(4*time.Second, doWork)

	done := make(chan interface{})
	time.AfterFunc(9*time.Second, func() {
		log.Println("main: halting steward and ward.")
		close(done)
	})

	for range doWorkWithSteward(done, 4*time.Second){}
	log.Println("Done")

	orDone := func(done ,c <-chan interface{}) <-chan interface{} {
		valStream := make(chan interface{})
		go func() {
			defer close(valStream)
			for {
				select {
				case <-done:
					return
				case v, ok := <-c:
					if ok == false {
						return
					}
					select {
					case valStream <- v:
					case <-done:
					}
				}
			}
		}()
		return valStream
	}

	bridge := func(
		done <-chan interface{},
		chanStream <-chan <-chan interface{},
	) <-chan interface{} {
		valStream := make(chan interface{})
		go func() {
			defer close(valStream)
			for {
				var stream <-chan interface{}
				select {
				case <-done:
					return
				case maybeStream, ok := <-chanStream:
					if ok == false {
						return
					}
					stream = maybeStream
				}
				for val := range orDone(done, stream) {
					select {
					case valStream <- val:
					case <-done:
					}
				}
			}

		}()
		return valStream
	}

	take := func(
		done <-chan interface{},
		valueCh <-chan interface{},
		num int,
	) <-chan interface{} {
		takeCh := make(chan interface{})
		go func() {
			defer close(takeCh)
			for i := 0; i < num; i++ {
				select {
				case <-done:
					return
				case takeCh <- <- valueCh:
				}
			}
		}()
		return takeCh
	}

	doWorkFn := func(
		done <-chan interface{},
		intList ...int,
	) (startGoroutineFn, <-chan interface{}) {
		intChanStream := make(chan (<-chan interface{}))
		intStream := bridge(done, intChanStream)

		doWork := func(
			done <-chan interface{},
			pulseInterval time.Duration,
		) <-chan interface{} {
			intStream := make(chan interface{})
			heartbeat := make(chan interface{})

			go func() {
				defer close(intStream)
				select {
				case intChanStream <- intStream:
				case <-done:
					return
				}

				pulse := time.Tick(pulseInterval)

				for {
					valueLoop:
					for _, interval := range intList {
						if interval < 0 {
							log.Printf("negative value: %v\n", interval)
							return
						}

						for {
							select {
							case <-pulse:
								select {
								case heartbeat <- struct{}{}:
								default:
								}
							case intStream <- interval:
								continue valueLoop
							case <-done:
								return
							}
						}
					}
				}
			}()
			return heartbeat
		}
		return doWork, intStream
	}

	log.SetFlags(log.Ltime | log.LUTC)
	log.SetOutput(os.Stdout)

	done = make(chan interface{})
	defer close(done)

	doWork, intStream := doWorkFn(done, 1, 2, -1, 3, 4, 5)
	doWorkWithSteward = newSteward(1*time.Millisecond, doWork)
	doWorkWithSteward(done, 1*time.Hour)
	for intVal := range take(done, intStream, 6) {
		fmt.Printf("Received: %v\n", intVal)
	}
}

func sample6() {

}