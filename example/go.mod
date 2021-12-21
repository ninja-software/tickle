module example

go 1.17

replace github.com/ninja-software/tickle => ../

require (
	github.com/ninja-software/tickle v1.2.1
	github.com/rs/zerolog v1.26.1
)

require github.com/ninja-software/terror v0.0.7 // indirect
