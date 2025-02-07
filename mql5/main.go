package main

import (
	"fmt"
	"time"
)

func main() {
	now := time.Now().UTC()
	from := now.Add(-time.Hour * 24 * 7)
	to := now
	events, err := fetch(from, to)
	if err != nil {
		fmt.Printf("Error: %+v\n", err)
		return
	}

	fmt.Printf("%+v\n", events)
}
