module example

go 1.17

replace github.com/ninja-software/tickle => ../

require (
	github.com/getsentry/sentry-go v0.11.0
	github.com/ninja-software/tickle v1.2.1
	github.com/rs/zerolog v1.26.1
)
