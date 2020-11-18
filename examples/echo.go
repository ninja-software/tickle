package main

import (
	"fmt"
	"log"
	"time"

	"github.com/ninja-software/tickle"
)

var count int = 0

var sayMoo tickle.Task = func() (int, error) {
	count++

	if count%5 == 0 {
		// error
		return count, fmt.Errorf("multiple of 5 is bad")
	}

	if count%3 == 0 {
		// panic
		arr := []int{}
		fmt.Println(arr[3])
	}

	log.Println("moo", count)
	return count, nil
}

var clean tickle.Clean = func(dat interface{}, err error) {
	log.Printf("moo %d is noooooo good\n", dat.(int))

	// panic 1
	// log.Printf("moo %s is noooooo good\n", dat.(string))

	// panic 2
	// arr := []int{}
	// fmt.Println(arr[555])
}

func main() {
	// tickle.Cow()

	tk := tickle.New(
		"sayMoo", // task name
		3,        // run every 180 second
		sayMoo,   // run the sayMoo() function
	)

	// do when sayMoo() returns error
	tk.FuncClean = clean

	tk.Start()

	// so program dont exit
	for {
		time.Sleep(time.Second)
	}
}
