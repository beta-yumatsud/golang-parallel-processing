package main

import (
	"bytes"
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