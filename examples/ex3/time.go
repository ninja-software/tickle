package main

import (
	"fmt"
	"time"
)

func main() {
	// min diff between local and utc

	now := time.Now()
	fmt.Println("now       ", now)

	utc := now.UTC()
	fmt.Println("now.utc() ", utc)

	diff := utc.Sub(now)
	fmt.Println("diff      ", diff)

	name, offset := now.Zone()
	fmt.Println("Zone      ", name)
	fmt.Println("offset    ", offset)

	ny, err := time.LoadLocation("America/New_York")
	if err != nil {
		panic(err)
	}
	fmt.Println("loc NY    ", ny)
	per, err := time.LoadLocation("Australia/Perth")
	if err != nil {
		panic(err)
	}
	fmt.Println("loc Perth ", per)

	nowtk := now.Truncate(time.Hour * 24)
	fmt.Println("now() trun", nowtk)

	nyt := now.In(ny)
	fmt.Println("NY.now()  ", nyt)

	nyttk := nyt.Truncate(time.Hour * 24)
	fmt.Println("loc NY tru", nyttk)

	pert := now.In(per)
	pertk := pert.Truncate(time.Hour * 24)
	fmt.Println("loc Pr tru", pertk)

	fmt.Println("loc NY utc", nyttk.In(time.UTC))
	fmt.Println("loc NY per", nyttk.In(per))
}
