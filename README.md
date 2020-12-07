# Tickle

Golang ticker based task runner. Comes with at time, reschedule, timezone, recovery and cleanup support. Can use user customer interface for wider support.

## How to use

```go
// main.go

// init new tickle
ss := tickle.New(
    "CheckBalance", // task name
    180,            // run every 180 second
    checkBalance,   // run the checkBalance() function
)

// create custom struct for execution to use
mtat := tickle.MeterTransferApprovalTask{
    Conn: db,
    ...
}

// register task to tickle
ss.TaskRegister(mtat.Runner)

// start tickle
ss.Start()
```

## Explain

```go
ss.TaskRegister(...) accepts function which will be executed
```

## Advanced


### At time support

Starts at 2:59 PM at system time.

```go
err = tk.SetIntervalAt(time.Second*10, 14, 59)
if err != nil {
    return err
}
```


### Timezone support

Starts at 2:59 PM at Perth Australia Time.

```go
loc, err := time.LoadLocation("Australia/Perth")
if err != nil {
    return err
}

err = tk.SetIntervalAtTimezone(time.Second*10, 14, 59, loc)
if err != nil {
    return err
}
```


### Recovery and Cleanup support

```go
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
    tk := tickle.New(
        "sayMoo", // task name
        10,       // run every 10 second
        sayMoo,   // run the sayMoo() function
    )

    // do when sayMoo() returns error
    tk.FuncClean = clean

    // do when sayMoo() panics
    tk.FuncRecovery = recovery

    tk.Start()
}
```

### Change interval

Change interval to 3 minutes.

```go
err = tk.SetInterval(time.Second*180)
if err != nil {
    return err
}
```
