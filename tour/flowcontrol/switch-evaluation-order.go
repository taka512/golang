package main

import (
	"fmt"
	"time"
)

func main() {
	fmt.Println("When Saturday?")
	today := time.Now().Weekday()
	switch time.Saturday {
	case today + 0:
		fmt.Println("Today.")
	case today + 1:
		fmt.Println("Tommorrow.")
	case today + 2:
		fmt.Println("IN two days.")
	default:
		fmt.Println("Too far away.")
	}
}