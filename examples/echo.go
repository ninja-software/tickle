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

	// // panic 1
	// log.Printf("moo %s is noooooo good\n", dat.(string))

	// // panic 2
	// arr := []int{}
	// fmt.Println(arr[555])
}

var recovery tickle.Recovery = func(err error) {
	// panic 3
	arr := []int{}
	fmt.Println(arr[999])

	arrow := ">>>>>>>>>>>>>>>>>>>>>>>"
	fmt.Printf("%s Instruction: keep calm and carry on.\n%s Error: %s\n", arrow, arrow, err.Error())
}

func main() {
	tk := tickle.New(
		"sayMoo", // task name
		3,        // run every 180 second
		sayMoo,   // run the sayMoo() function
	)

	// do when sayMoo() returns error
	tk.FuncClean = clean
	tk.FuncRecovery = recovery
	tk.Start()

	// so program dont exit
	for {
		time.Sleep(time.Second)
	}
}
