package tickle

import (
	"fmt"
	"runtime/debug"
	"time"

	"log"

	"github.com/ninja-software/terror"
)

const Version = "v1.2.0"

// MinDurationOverride override mininum 10 seconds per loop
var MinDurationOverride = false

// Tickle contain the information that the tickle inner settings
type Tickle struct {
	Name string // name of the scheduled task

	FuncTask     Task     // function to be executed on regular interval
	FuncClean    Clean    // function to run when error occurred (optional)
	FuncRecovery Recovery // function to run when panic occurred (optional)

	Count        int // number of times this been triggered
	CountFail    int // number of failed trigger
	CountSuccess int // number of successful trigger

	LastError *error     // what is the task's last error
	LastTick  *time.Time // when is the task last ran (Note: changing will not affect the ticker)
	NextTick  *time.Time // when the next TaskRun() will be triggered (Note: changing will not affect the ticker)
	StartedAt *time.Time // when the tick was started (Note: changing will not affect the ticker)

	// time allowed to run in a range (inclusive)   -----[      ]-----    [ = open      ] = close
	TimeRangeOpen  time.Time // when the task is allowed to run after
	TimeRangeClose time.Time // when the task is allowed to run before

	StopMaxInterval int // stops when maximum number of interval reached
	StopMaxError    int // stops when maximum number of consecutive error reached

	// internal
	intervalSecond float64      // how many seconds each interval the task will run at
	ticker         *time.Ticker // internal ticker
	timerTicker    *time.Timer  // timer that starts the ticker above
}

// Task uses user supplied function to run on interval
// It will returns the number of action/change/touch/created/update/delete performed and error status
type Task func() (int, error)

// Clean uses user supplied function to run clean from error
type Clean func(interface{}, error)

// Recovery uses user supplied function to run when panic occurred
type Recovery func(error)

// Start will begin the tickle
func (sc *Tickle) Start() {
	log.Printf("Start tickle (%s)\n", sc.Name)

	var duration time.Duration = time.Duration(float64(time.Second) * sc.intervalSecond)

	// remember
	now := time.Now()
	sc.StartedAt = &now
	next := now.Add(time.Duration(float64(time.Second) * sc.intervalSecond))
	sc.NextTick = &next

	sc.ticker = time.NewTicker(duration)
	done := make(chan bool, 1)
	go func(t *time.Ticker) {
		for {
			select {
			case <-t.C:
				sc.TaskRun()
			case <-done:
				log.Printf("Tickle ticker done. (%s)\n", sc.Name)
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

	// remember
	now := time.Now()
	sc.LastTick = &now
	next := now.Add(time.Duration(sc.intervalSecond) * time.Second)
	sc.NextTick = &next

	// recover from panic from FuncRecovery
	defer func() {
		if rec := recover(); rec != nil {
			message := "Tickle task panicked-panicked (" + sc.Name + ")"
			log.Println(message)
			strStack := string(debug.Stack())

			var err error
			switch v := rec.(type) {
			case error:
				err = v
			default:
				err = fmt.Errorf(message)
			}
			sc.LastError = &err

			log.Println("Tickle panic-panic recovered ("+sc.Name+"): ", err, "\n", strStack)
		}
	}()
	// recover from panic
	defer func() {
		if rec := recover(); rec != nil {
			message := "Tickle task panicked (" + sc.Name + ")"
			log.Println(message)
			strStack := string(debug.Stack())

			var err error
			switch v := rec.(type) {
			case error:
				err = v
			default:
				err = fmt.Errorf(message)
			}
			sc.LastError = &err

			log.Println("Tickle panic recovered ("+sc.Name+"): ", err, "\n", strStack)

			if sc.FuncRecovery != nil {
				sc.FuncRecovery(err)
			}
		}
	}()

	log.Printf("Tickle task run (%s)\n", sc.Name)
	defer func() {
		log.Printf("Tickle task exit (%s)\n", sc.Name)
	}()

	if sc.FuncTask == nil {
		err := fmt.Errorf("Tickle func is nil")
		log.Printf("Tickle task failed (%s)\n", sc.Name)
		terror.Echo(err)
		sc.LastError = &err
		sc.CountFail++
		sc.Count++
		return
	}

	dat, err := sc.FuncTask()
	if err != nil {
		log.Printf("Tickle task failed (%s)\n", sc.Name)
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
	if !MinDurationOverride && interval.Seconds() < 10 {
		return terror.New(fmt.Errorf("duration must be 10 seconds or above"), "")
	}

	sc.intervalSecond = float64(interval.Seconds())

	sc.Stop()
	sc.Start()

	return nil
}

// SetIntervalAt change the ticker interval and start at specified hour and minute of the day, using local timezone
func (sc *Tickle) SetIntervalAt(interval time.Duration, startHour, startMinute int) error {
	// _, offsetSecond := time.Now().Zone()
	loc := time.Local

	return sc.SetIntervalAtTimezone(interval, startHour, startMinute, loc)
}

// SetIntervalAtTimezone change the ticker interval and start at specified hour and minute of the day, with target timezone offset in minutes (Note: will auto stop and auto start after set)
func (sc *Tickle) SetIntervalAtTimezone(interval time.Duration, startHour, startMinute int, loc *time.Location) error {
	if !MinDurationOverride && interval.Seconds() < 10 {
		return terror.New(fmt.Errorf("duration must be 10 seconds or above"), "")
	}
	if startHour < -1 || startHour > 23 {
		return terror.New(fmt.Errorf("startHour must be range of -1..23"), "")
	}
	if startMinute < -1 || startMinute > 59 {
		return terror.New(fmt.Errorf("startMinute must be range of -1..59"), "")
	}

	if sc.ticker != nil {
		sc.Stop()
	}
	if sc.timerTicker != nil {
		sc.timerTicker.Stop()
	}

	now := time.Now()

	// start time, st
	st := now.UTC()

	if startHour == -1 && startMinute == -1 {
		fmt.Println("course 1")
		// start next minute
		st = st.Truncate(time.Minute)

		// if now is after next start time, make it next minute
		if now.After(st) {
			st = st.Add(time.Minute)
		}

	} else if startHour == -1 && startMinute > -1 {
		fmt.Println("course 2")
		// start beginning of next hour at matching minute
		st = st.Truncate(time.Hour)
		st = st.Add(time.Minute * time.Duration(startMinute))

		// if now is after next start time, make it next hour
		if now.After(st) {
			st = st.Add(time.Hour)
		}

	} else if startHour > -1 {
		fmt.Println("course 3")
		// start beginning of next matching hour at matching minute
		// or if startMinute == -1, then minute at 0
		if startMinute == -1 {
			startMinute = 0
		}
		st = time.Date(st.Year(), st.Month(), st.Day(), startHour, startMinute, 0, 0, loc)

		// if now is after next start time, make it next day
		if now.After(st) {
			st = st.Add(time.Hour * 24)
			fmt.Println(21, st)
		}

	} else {
		fmt.Println("course 5")
		// it shouldn't reach here
		return terror.New(fmt.Errorf("unknown condition"), "")
	}

	sc.StartedAt = &st
	sc.NextTick = &st
	sc.intervalSecond = float64(interval.Seconds())
	startInDuration := st.Sub(now)

	log.Printf("Set tickle (%s). Starts at %s (interval %s)\n", sc.Name, startInDuration.String(), interval.String())

	a := time.AfterFunc(startInDuration, func() {
		sc.TaskRun()
		sc.Start()
	})
	sc.timerTicker = a

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
	log.Println("Stop tickle")

	// reset the time info
	sc.StartedAt = nil
	sc.NextTick = nil

	sc.ticker.Stop()
}

// New makes creating tickle easily
func New(
	taskName string, // name of the task to identify, please make it unique
	timeSecond float64, // interval in seconds
	funcTask Task, //  function for task to execute
) *Tickle {
	if !MinDurationOverride && timeSecond < 10 {
		panic("cannot be less than 10 seconds for interval")
	}
	if timeSecond < 0 {
		panic("must be larger than 0 second for interval")
	}
	if funcTask == nil {
		panic("task must be given")
	}

	tk := &Tickle{
		Name:            taskName,
		FuncTask:        funcTask,
		intervalSecond:  timeSecond,
		StopMaxInterval: 2147483647, // ~68 years if triggered every second
	}

	return tk
}
