package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/ninja-software/log_helpers"
	"github.com/ninja-software/tickle"
	"github.com/rs/zerolog"
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

	fmt.Println("moo", count)
	return count, nil
}

var recovery tickle.Recovery = func(err error) {
	fmt.Printf("Instruction: keep calm and carry on.\nError: %s\n", err.Error())
	// happen when count equals 5, 10, 15, ...
}

var clean tickle.Clean = func(dat interface{}, err error) {
	fmt.Printf("moo %d is noooooo good\n", dat.(int))
	// happen when count equals 3, 6, 9, ...
}

func main() {
	log := log_helpers.LoggerInitZero("development")
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGKILL)
	defer stop()

	taskname := "sayMoo"
	tickle.MinDurationOverride = true
	tk := tickle.New(
		taskname, // task name
		1,        // run every 10 second
		sayMoo,   // run the sayMoo() function
	)

	// Override the standard logger
	tkLogger := log_helpers.NamedLogger(log, "tickle").With().Str("task", taskname).Logger().Level(zerolog.DebugLevel)
	tk.Log = &tkLogger
	tk.LogVerboseMode = true

	// do when sayMoo() returns error
	tk.FuncClean = clean

	// do when sayMoo() panics
	tk.FuncRecovery = recovery

	tk.Start()
	<-ctx.Done()
}
