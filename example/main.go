package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

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

var recovery tickle.Recovery = func(err error) {
	fmt.Printf("Instruction: keep calm and carry on.\nError: %s\n", err.Error())
	// happen when count equals 5, 10, 15, ...
}

var clean tickle.Clean = func(dat interface{}, err error) {
	log.Printf("moo %d is noooooo good\n", dat.(int))
	// happen when count equals 3, 6, 9, ...
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGKILL)
	defer stop()

	tickle.MinDurationOverride = true
	tk := tickle.New(
		"sayMoo", // task name
		1,        // run every 10 second
		sayMoo,   // run the sayMoo() function
	)

	// do when sayMoo() returns error
	tk.FuncClean = clean

	// do when sayMoo() panics
	tk.FuncRecovery = recovery

	tk.Start()
	<-ctx.Done()
}
