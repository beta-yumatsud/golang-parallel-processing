package main

import "fmt"

func ExampleHello() {
	fmt.Println("Hello")
	// Output: Hello
}

func ExampleShuffle() {
	x := map[string]int{"a": 1, "b": 2, "c": 3}
	for k, v := range x {
		fmt.Printf("k=%s, v=%d", k, v)
	}
	// Unordered output:
	// k=a v=1
	// k=b v=2
	// k=c v=3
}