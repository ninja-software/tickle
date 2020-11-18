package tickle

import (
	"fmt"
	"runtime/debug"
	"time"

	"github.com/ninja-software/terror"
	"github.com/prometheus/common/log"
)

// Tickle contain the information that the tickle inner settings
type Tickle struct {
	Name           string // name of the scheduled task
	intervalSecond int    // what interval the task will run at

	ticker *time.Ticker // internal ticker

	FuncTask     Task     // function to be executed on regular interval
	FuncClean    Clean    // function to run when error occured (optional)
	FuncRecovery Recovery // function to run when panic occured (optional)

	StartZero    bool // start the task immediately when tickle starts
	Count        int  // number of times this been triggered
	CountFail    int  // number of failed trigger
	CountSuccess int  // number of successful trigger

	LastError *error    // what is the task's last error
	LastTick  time.Time // when is the task last ran

	TimeOut time.Duration // how long the task should wait before give up, does not affect the next task

	// time allowed to run in a range (inclusive)   -----[      ]-----    [ = open      ] = close
	TimeRangeOpen  time.Time // when the task is allowed to run after
	TimeRangeClose time.Time // when the task is allowed to run before

	StopMaxInterval int // stops when maximum number of interval reached
	StopMaxError    int // stops when maximum number of consecutive error reached
}

// Task uses user supplied function to run on interval
// It will returns the number of action/change/touch/created/update/delete performed and error status
type Task func() (int, error)

// Clean uses user supplied function to run clean from error
type Clean func(interface{}, error)

// Recovery uses user supplied function to run when panic occured
type Recovery func(error)

// Start will begin the tickle
func (sc *Tickle) Start() {
	if sc.FuncTask == nil {
		log.Errorf(" Err: Tickle have task registered (%s)\n", sc.Name)
		return
	}

	log.Infof("Start tickle (%s)\n", sc.Name)

	// run first time, then ticker next
	if sc.StartZero {
		sc.TaskRun()
	}

	var duration time.Duration = time.Second * time.Duration(sc.intervalSecond)

	sc.ticker = time.NewTicker(duration)
	done := make(chan bool, 1)
	go func(t *time.Ticker) {
		for {
			select {
			case <-t.C:
				sc.TaskRun()
			case <-done:
				log.Infof("Tickle ticker done. (%s)\n", sc.Name)
				return
			}
		}
	}(sc.ticker)
}

// TaskRun execute the function (task) it been assigned to
func (sc *Tickle) TaskRun() {
	// sanity check
	// too early
	if !sc.TimeRangeOpen.IsZero() && time.Now().Before(sc.TimeRangeOpen) {
		return
	}
	// too late
	if !sc.TimeRangeClose.IsZero() && time.Now().After(sc.TimeRangeClose) {
		return
	}
	// too many
	if sc.StopMaxInterval > 0 && sc.Count > sc.StopMaxInterval {
		return
	}
	// too error
	if sc.StopMaxError > 0 && sc.CountFail > sc.StopMaxError {
		return
	}

	// clear start zero
	sc.StartZero = false
	// remember
	sc.LastTick = time.Now()

	// recover from panic from FuncRecovery
	defer func() {
		if rec := recover(); rec != nil {
			message := "Tickle task panicked-panicked (" + sc.Name + ")"
			log.Errorln(message)
			strStack := string(debug.Stack())

			var err error
			switch v := rec.(type) {
			case error:
				err = v
			default:
				err = fmt.Errorf(message)
			}
			sc.LastError = &err

			log.Errorln("Tickle panic-panic recovered ("+sc.Name+"): ", err, "\n", strStack)
		}
	}()
	// recover from panic
	defer func() {
		if rec := recover(); rec != nil {
			message := "Tickle task panicked (" + sc.Name + ")"
			log.Errorln(message)
			strStack := string(debug.Stack())

			var err error
			switch v := rec.(type) {
			case error:
				err = v
			default:
				err = fmt.Errorf(message)
			}
			sc.LastError = &err

			log.Errorln("Tickle panic recovered ("+sc.Name+"): ", err, "\n", strStack)

			if sc.FuncRecovery != nil {
				sc.FuncRecovery(err)
			}
		}
	}()

	log.Infof("Tickle task run (%s)\n", sc.Name)
	defer func() {
		log.Infof("Tickle task exit (%s)\n", sc.Name)
	}()

	if sc.FuncTask == nil {
		err := fmt.Errorf("Tickle func is nil")
		log.Errorf("Tickle task failed (%s)\n", sc.Name)
		terror.Echo(err)
		sc.LastError = &err
		sc.CountFail++
		sc.Count++
		return
	}

	dat, err := sc.FuncTask()
	if err != nil {
		log.Errorf("Tickle task failed (%s)\n", sc.Name)
		terror.Echo(err)
		sc.LastError = &err
		sc.CountFail++

		if sc.FuncClean != nil {
			sc.FuncClean(dat, err)
		}
	} else {
		sc.CountSuccess++
		sc.LastError = nil
	}
	// inc by 1
	sc.Count++
}

// SetInterval change the ticker reoccuring time rate
func (sc *Tickle) SetInterval(interval time.Duration) error {
	if interval.Seconds() < 10 {
		return terror.New(fmt.Errorf("duration must be 10 seconds or above"), "")
	}

	sc.intervalSecond = int(interval.Seconds())

	sc.Stop()
	sc.Start()

	return nil
}

// SetTimeOpen change the time range that task would run
func (sc *Tickle) SetTimeOpen(y, m, d, h, min, s int) error {
	if m < 1 || m > 12 {
		return terror.New(fmt.Errorf("wrong month number %d", m), "")
	}

	mth := time.Month(m)
	sc.TimeRangeOpen = time.Date(y, mth, d, h, min, s, 0, time.Local)

	sc.Stop()
	sc.Start()

	return nil
}

// SetTimeClose change the time range that task would not run
func (sc *Tickle) SetTimeClose(y, m, d, h, min, s int) error {
	if m < 1 || m > 12 {
		return terror.New(fmt.Errorf("wrong month number %d", m), "")
	}

	mth := time.Month(m)
	sc.TimeRangeClose = time.Date(y, mth, d, h, min, s, 0, time.Local)

	sc.Stop()
	sc.Start()

	return nil
}

// CounterReset reset all counters to zero
func (sc *Tickle) CounterReset() {
	sc.Count = 0
	sc.CountFail = 0
	sc.CountSuccess = 0
}

// Stop will halt the tickle
func (sc *Tickle) Stop() {
	log.Info("Stop tickle")

	sc.ticker.Stop()
}

// New makes creating tickle easily
func New(
	taskName string, // name of the task to identify, please make it unique
	timeSecond int, // interval in seconds
	funcTask Task, //  function for task to execute
) *Tickle {
	// TODO enable me after test
	// if timeSecond < 10 {
	// 	panic("cannot be less than 10 seconds for interval")
	// }
	if funcTask == nil {
		panic("task must be given")
	}

	tk := &Tickle{
		Name:            taskName,
		FuncTask:        funcTask,
		intervalSecond:  timeSecond,
		StopMaxInterval: 2147483647, // ~68 years if triggger every second
	}

	return tk
}
