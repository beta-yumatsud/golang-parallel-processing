package main

import (
	"bytes"
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"sync"
	"time"
)

func main() {
	sample1()
	sample2()
	sample3()
	sample4()
	sample5()
	sample6()
	sample7()
	sample8()
	sample11()
	sample12()
	sample13()
	sample14()
}

func sample1() {
	printData := func(wg *sync.WaitGroup, data []byte) {
		defer wg.Done()

		var buff bytes.Buffer
		for _, b := range data {
			fmt.Fprintf(&buff, "%c", b)
		}
		fmt.Println(buff.String())
	}

	var wg sync.WaitGroup
	wg.Add(2)
	data := []byte("golang")
	go printData(&wg, data[:3])
	go printData(&wg, data[3:])

	wg.Wait()
}

func sample2() {
	doWork := func(
		done <-chan interface{},
		strings <-chan string,
	) <-chan interface{} {
		terminated := make(chan interface{})
		go func() {
			defer fmt.Println("doWork exited...")
			defer fmt.Println("doWork exited..")
			defer fmt.Println("doWork exited.")
			defer close(terminated)

			for {
				select {
				case <-done:
					return
				case s := <-strings:
					fmt.Printf("hahahahaha: %s\n", s)
				}
			}
		}()
		return terminated
	}

	done := make(chan interface{})
	terminated := doWork(done, nil)

	go func() {
		time.Sleep(3 * time.Second)
		fmt.Println("Canceling doWork goroutine...")
		close(done)
	}()

	<-terminated
	fmt.Println("Done.")
}

func sample3() {
	newRandStream := func(done <-chan interface{}) <-chan int {
		randStream := make(chan int)
		go func() {
			defer fmt.Println("newRandStream closure exited.")
			defer close(randStream)

			for {
				select {
				case <-done:
					return
				case randStream <- rand.Int():
				}
			}
		}()
		return randStream
	}

	done := make(chan interface{})
	randStream := newRandStream(done)
	fmt.Println("3 random ints:")
	for i := 0; i < 3; i++ {
		fmt.Printf("%d: %d\n", i, <-randStream)
	}
	close(done)
	time.Sleep(3 * time.Second)
}

func sample4() {
	var or func(channels ...<-chan interface{}) <-chan interface{}
	or = func(channels ...<-chan interface{}) <-chan interface{} {
		switch len(channels) {
		case 0:
			return nil
		case 1:
			return channels[0]
		}

		orDone := make(chan interface{})
		go func() {
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

	sig := func(after time.Duration) <-chan interface{} {
		c := make(chan interface{})
		go func() {
			defer close(c)
			time.Sleep(after)
		}()
		return c
	}

	start := time.Now()
	<-or(
		sig(2*time.Hour),
		sig(5*time.Minute),
		sig(1*time.Second),
		sig(1*time.Hour),
		sig(1*time.Minute),
	)
	fmt.Printf("done after %v\n", time.Since(start))
}

func sample5() {
	type Result struct {
		Error    error
		Response *http.Response
	}

	checkStatus := func(done <-chan interface{}, urls ...string) <-chan Result{
		results := make(chan Result)
		go func() {
			defer close(results)
			for _, url := range urls {
				var result Result
				resp, err := http.Get(url)
				result = Result{Error: err, Response: resp}
				select {
				case <-done:
					return
				case results <- result:
				}
			}
		}()
		return results
	}

	done := make(chan interface{})
	defer close(done)

	urls := []string{"https://www.google.com", "https://www.yahoo.co.jp"}
	for result := range checkStatus(done, urls...) {
		if result.Error != nil {
			fmt.Printf("error: %v", result.Error)
			continue
		}
		fmt.Printf("Response: %v\n", result.Response.Status)
	}
}

func sample6() {
	generator := func(done <-chan interface{}, integers ...int) <-chan int {
		intCh := make(chan int, len(integers))
		go func() {
			defer close(intCh)
			for _, i := range integers {
				select {
				case <-done:
					return
				case intCh <- i:
				}
			}
		}()
		return intCh
	}

	multiply := func(
		done <-chan interface{},
		intCh <-chan int,
		multiplier int,
	) <-chan int{
		multipliedCh := make(chan int)
		go func() {
			defer close(multipliedCh)
			for i := range intCh {
				select {
				case <-done:
					return
				case multipliedCh <- i * multiplier:
				}
			}
		}()
		return multipliedCh
	}

	add := func(
		done <-chan interface{},
		intCh <-chan int,
		additive int,
	) <-chan int{
		addedCh := make(chan int)
		go func() {
			defer close(addedCh)
			for i := range intCh {
				select {
				case <-done:
					return
				case addedCh <- i + additive:
				}
			}
		}()
		return addedCh
	}

	done := make(chan interface{})
	defer close(done)

	intCh := generator(done, 1, 2, 3, 4)
	pipeline := multiply(done, add(done, multiply(done, intCh, 2), 1), 2)

	for v := range pipeline {
		fmt.Println(v)
	}
}

func sample7() {
	repeat := func(
		done <-chan interface{},
		values ...interface{},
	) <-chan interface{} {
		valueCh := make(chan interface{}, len(values))
		go func() {
			defer close(valueCh)
			for {
				for _, v := range values {
					select {
					case <-done:
						return
					case valueCh <- v:
					}
				}
			}
		}()
		return valueCh
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

	done := make(chan interface{})
	defer close(done)

	for num := range take(done, repeat(done, 1, 2, 3), 10) {
		fmt.Printf("%v ", num)
	}
}

func sample8() {
	repeatFn := func(
		done <-chan interface{},
		fn func() interface{},
	) <-chan interface{} {
		valueCh := make(chan interface{})
		go func() {
			defer close(valueCh)
			for {
				select {
				case <-done:
					return
				case valueCh <- fn():
				}
			}
		}()
		return valueCh
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

	done := make(chan interface{})
	defer close(done)

	randFn := func() interface{} { return rand.Int() }

	for num := range take(done, repeatFn(done, randFn), 10) {
		fmt.Printf("num is %v\n", num)
	}
}
/*
func sample9() {
	fanIn := func(
		done <-chan interface{},
		channels ...<-chan interface{},
	) <-chan interface{} {
		var wg sync.WaitGroup
		multiplexedStream := make(chan interface{})

		multiplex := func(c <-chan interface{}) {
			defer wg.Done()
			for i := range c {
				select {
				case <-done:
					return
				case multiplexedStream <- i:
				}
			}
		}

		wg.Add(len(channels))
		for _, c := range channels {
			go multiplex(c)
		}

		go func() {
			wg.Wait()
			close(multiplexedStream)
		}()

		return multiplexedStream
	}

	done := make(chan interface{})
	defer close(done)

	start := time.Now()

	rand2 := func() interface{} { return rand.Intn(50000000) }
	randIntStream := toInt(done, repeatFn(done, rand2))

	numFinders := runtime.NumCPU()
	fmt.Printf("Spinning up %d prime finders.\n", numFinders)
	finders := make([]<-chan interface{}, numFinders)
	fmt.Println("Primes:")
	for i := 0; i < numFinders; i++ {
		finders[i] = primeFinder(done, randIntStream) }
	for prime := range take(done, fanIn(done, finders...), 10) {
		fmt.Printf("\t%d\n", prime)
	}
	fmt.Printf("Search took: %v", time.Since(start))
}
*/
/*
func sample10() {
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

	done := make(chan interface{})
	defer close(done)

	for val := range orDone(done, myChan) {
		// valに対して何かする
	}
}
*/

func sample11() {
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

	genVals := func() <-chan <-chan interface{} {
		chanStream := make(chan (<-chan interface{}))
		go func() {
			defer close(chanStream)
			for i := 0; i < 10; i++ {
				stream := make(chan interface{}, 1)
				stream <- i
				close(stream)
				chanStream <- stream
			}
		}()
		return chanStream
	}
	for v := range bridge(nil, genVals()) {
		fmt.Printf("%v ", v)
	}

}

// Queueとして扱うために、バッファ付きチャネルを使う話。リトルの法則

// Context
func sample12() {
	var wg sync.WaitGroup
	done := make(chan interface{})
	defer close(done)

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := printGreeting(done); err != nil {
			fmt.Printf("%v", err)
			return
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := printFarewell(done); err != nil {
			fmt.Printf("%v", err)
			return }
	}()

	wg.Wait()
}

func printGreeting(done <-chan interface{}) error {
	greeting, err := genGreeting(done)
	if err != nil {
		return err
	}
	fmt.Printf("%s world!\n", greeting)
	return nil
}

func genGreeting(done <-chan interface{}) (string, error) {
	switch locale, err := locale(done); {
	case err != nil:
		return "", err
	case locale == "EN/US":
		return "hello", nil
	}
	return "", fmt.Errorf("unsupported locale")
}

func printFarewell(done <-chan interface{}) error {
	farewell, err := genFarewell(done)
	if err != nil {
		return err
	}
	fmt.Printf("%s world!\n", farewell)
	return nil
}

func genFarewell(done <-chan interface{}) (string, error) {
	switch locale, err := locale(done); {
	case err != nil:
		return "", err
	case locale == "EN/US":
		return "goodbye", nil
	}
	return "", fmt.Errorf("unsupported locale")
}

func locale(done <-chan interface{}) (string, error) {
	select {
	case <-done:
		return "", fmt.Errorf("canceled")
	case <-time.After(1 * time.Second):
	}
	return "EN/US", nil
}

func sample13() {
	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	wg.Add(1)
	go func() {
		defer wg.Done()

		if err := printGreetingWithContext(ctx); err != nil {
			fmt.Printf("cannot print greeting: %v\n", err)
			cancel()
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		if err := printFarewellWithContext(ctx); err != nil {
			fmt.Printf("cannot print farewell: %v\n", err)
		}
	}()

	wg.Wait()
}

func printGreetingWithContext(ctx context.Context) error {
	greeting, err := genGreetingWithContext(ctx)
	if err != nil {
		return err
	}
	fmt.Printf("%s world!\n", greeting)
	return nil
}

func printFarewellWithContext(ctx context.Context) error {
	farewell, err := genFarewellWithContext(ctx)
	if err != nil {
		return err
	}
	fmt.Printf("%s world!\n", farewell)
	return nil
}

func genGreetingWithContext(ctx context.Context) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	switch locale, err := localeWithContext(ctx); {
	case err != nil:
		return "", err
	case locale == "EN/US":
		return "hello", nil
	}
	return "", fmt.Errorf("unsupported locale")
}

func genFarewellWithContext(ctx context.Context) (string, error) {
	switch locale, err := localeWithContext(ctx); {
	case err != nil:
		return "", err
	case locale == "EN/US":
		return "goodbye", nil
	}
	return "", fmt.Errorf("unsupported locale")
}

func localeWithContext(ctx context.Context) (string, error) {
	if deadline, ok := ctx.Deadline(); ok {
		if deadline.Sub(time.Now().Add(1 * time.Minute)) <= 0 {
			return "", context.DeadlineExceeded
		}
	}

	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case <-time.After(1 * time.Minute):
	}
	return "EN/US", nil
}

type ctxKey int

const (
	ctxUserID    ctxKey = iota
	ctxAuthToken
)

func UserID(c context.Context) string {
	return c.Value(ctxUserID).(string)
}

func AuthToken(c context.Context) string {
	return c.Value(ctxAuthToken).(string)
}

func sample14() {
	ProcessRequest("jane", "abc123")
}

func ProcessRequest(userID, authToken string) {
	ctx := context.WithValue(context.Background(), ctxUserID, userID)
	ctx = context.WithValue(ctx, ctxAuthToken, authToken)
	HandleResponse(ctx)
}

func HandleResponse(ctx context.Context) {
	fmt.Printf(
		"handling response for %v (%v)\n",
		UserID(ctx),
		AuthToken(ctx),
	)
}
