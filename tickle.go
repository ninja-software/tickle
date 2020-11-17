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
	Name         string        // name of the scheduled task
	timeInterval time.Duration // what interval it will trigger
	ticker       *time.Ticker  // internal ticker

	funcTask  Task  // task function to be executed
	funcClean Clean // function to run when error occured (optional)

	StartZero    bool // start the task immediately when tickle starts
	Count        int  // number of times this been triggered
	CountFail    int  // number of failed trigger
	CountSuccess int  // number of successful trigger

	LastError *error    // what is the task's last error
	LastTick  time.Time // when is the task last ran

	TimeOut   time.Duration // how long the task should wait before give up, does not affect the next task
	TimeStart time.Time     // when the task should start running
	TimeStop  time.Time     // when the task should stop running

	StopMaxInterval int // stops when maximum number of interval reached
	StopMaxError    int // stops when maximum number of consecutive error reached
}

// Task uses user supplied function to run on interval
// It will returns the number of action/change/touch/created/update/delete performed and error status
type Task func() (int, error)

// Clean uses user supplied function to run clean from error
type Clean func(interface{}, error)

// Start will begin the tickle
func (sc *Tickle) Start() {
	if sc.funcTask == nil {
		log.Errorf(" Err: Tickle no task registered (%s)\n", sc.Name)
		return
	}

	log.Infof("Start tickle (%s)\n", sc.Name)

	// run first time, then ticker next
	if sc.StartZero {
		sc.TaskRun()
	}

	sc.ticker = time.NewTicker(sc.timeInterval)
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
	if !sc.TimeStart.IsZero() && time.Now().Before(sc.TimeStart) {
		return
	}
	// too late
	if !sc.TimeStop.IsZero() && time.Now().After(sc.TimeStop) {
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
		}

		// ctx := context.Background()
		// helpers.SentrySend(ctx, nil, nil, err, strStack)
	}()

	log.Infof("Tickle task run (%s)\n", sc.Name)
	defer func() {
		log.Infof("Tickle task exit (%s)\n", sc.Name)
	}()

	dat, err := sc.funcTask()
	if err != nil {
		log.Errorf("Tickle task failed (%s)\n", sc.Name)
		terror.Echo(err)
		sc.LastError = &err
		sc.CountFail++

		if sc.funcClean != nil {
			sc.funcClean(dat, err)
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

	sc.timeInterval = interval

	sc.Stop()
	sc.Start()

	return nil
}

// SetTimeStart change the time that task would run
func (sc *Tickle) SetTimeStart(y, m, d, h, min, s int) error {
	if m < 1 || m > 12 {
		return terror.New(fmt.Errorf("wrong month number %d", m), "")
	}

	mth := time.Month(m)
	sc.TimeStart = time.Date(y, mth, d, h, min, s, 0, time.Local)

	sc.Stop()
	sc.Start()

	return nil
}

// SetTimeStop change the time that task would not run
func (sc *Tickle) SetTimeStop(y, m, d, h, min, s int) error {
	if m < 1 || m > 12 {
		return terror.New(fmt.Errorf("wrong month number %d", m), "")
	}

	mth := time.Month(m)
	sc.TimeStop = time.Date(y, mth, d, h, min, s, 0, time.Local)

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

// SetTask set the function (task) for running when called
func (sc *Tickle) SetTask(task Task) {
	sc.funcTask = task
}

// SetTaskAndStart adds the function (task) and start immediately
func (sc *Tickle) SetTaskAndStart(task Task) {
	sc.funcTask = task
	sc.Start()
}

// Stop will halt the tickle
func (sc *Tickle) Stop() {
	log.Info("Stop tickle")

	sc.ticker.Stop()
}

// New will return a new tickle
func New(
	taskName string,
	timeSecond int,
	startZero bool,
	funcTask Task,
	funcClean Clean, // optional, can use nil
) *Tickle {
	// TODO enable me after test
	// if timeSecond < 10 {
	// 	panic("cannot be less than 10 seconds for interval")
	// }
	if funcTask == nil {
		panic("task must be given")
	}

	var ti time.Duration
	ti = time.Second * time.Duration(timeSecond)

	tk := &Tickle{
		Name:            taskName,
		funcTask:        funcTask,
		timeInterval:    ti,
		StartZero:       startZero,
		StopMaxInterval: 2147483647, // ~68 years if triggger every second
	}

	if funcClean != nil {
		tk.funcClean = funcClean
	}

	return tk
}
