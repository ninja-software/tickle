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
	// // panic 3
	// arr := []int{}
	// fmt.Println(arr[999])

	arrow := ">>>>>>>>>>>>>>>>>>>>>>>"
	fmt.Printf("%s Instruction: keep calm and carry on.\n%s Error: %s\n", arrow, arrow, err.Error())
}

func main() {
	tk := tickle.New(
		"sayMoo", // task name
		10,       // run every 3 second
		sayMoo,   // run the sayMoo() function
	)

	// do when sayMoo() returns error
	tk.FuncClean = clean
	// do when sayMoo() panics
	tk.FuncRecovery = recovery
	// tk.TaskRun()
	// tk.Start()
	// err := tk.SetIntervalAt(time.Second*3, 13, 00)

	loc, err := time.LoadLocation("Australia/Sydney")
	if err != nil {
		log.Fatal(err)
	}

	err = tk.SetIntervalAtTimezone(time.Second*3, 19, 0, loc)
	if err != nil {
		log.Fatal(err)
	}

	// so program dont exit
	for {
		time.Sleep(time.Second)
	}
}
