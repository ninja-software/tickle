# Tickle time (Incomplete)

Golang ticker based task runner.

## How to use

```go
main.go

// init new tickle
ss := tickle.New(
    "CheckBalance", // task name
    180,            // run every 180 second
    true,           // run immediately when initiated
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
