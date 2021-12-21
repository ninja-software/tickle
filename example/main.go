package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/getsentry/sentry-go"
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
	fmt.Printf("moo %d is no good\n", dat.(int))
	// happen when count equals 3, 6, 9, ...
}

func main() {
	// Start context to block on
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGKILL)
	defer stop()

	// Initialise sentry
	err := sentry.Init(sentry.ClientOptions{})
	if err != nil {
		panic(err)
	}

	// Initialise tickle
	taskname := "sayMoo"
	tickle.MinDurationOverride = true
	tk := tickle.New(
		taskname, // task name
		1,        // run every 10 second
		sayMoo,   // run the sayMoo() function
	)

	// Override tickle standard library logger
	log := zerolog.New(zerolog.NewConsoleWriter()).With().Timestamp().Logger()
	tkLogger := log.With().Str("name", "tickle").Str("task", taskname).Logger().Level(zerolog.DebugLevel)
	tk.Log = &tkLogger
	tk.LogVerboseMode = true

	// Add sentry as the tracer
	tk.Tracer = NewSentryTracer()
	tk.TracerPerentCtx = ctx

	// Run clean on error
	tk.FuncClean = clean

	// Run recover on panics
	tk.FuncRecovery = recovery

	// Start tickle cycle
	tk.Start()

	// Block for the example
	<-ctx.Done()
}

type SentryTracer struct{}

func NewSentryTracer() *SentryTracer {
	return &SentryTracer{}
}

func (t *SentryTracer) OnTaskStart(ctx context.Context, log tickle.Logger, operation string, taskName string) context.Context {
	// Setup tracing
	perentSpan := sentry.TransactionFromContext(ctx)
	if perentSpan == nil {
		perentSpan = sentry.StartSpan(ctx, operation)
	}
	span := sentry.StartSpan(ctx, operation)

	ctx = span.Context() // This allows this span to be accessed by TracerStop and to be used as a perent
	return ctx
}

func (t *SentryTracer) OnTaskStop(ctx context.Context, log tickle.Logger, taskName string) {
	span := sentry.TransactionFromContext(ctx)
	span.Finish()

	log.Printf("%s trace | %s Call took %s", span.Op, span.EndTime.Sub(span.StartTime))
}
