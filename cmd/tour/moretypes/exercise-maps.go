package main

import (
	"strings"

	"golang.org/x/tour/wc"
)

func WordCount(s string) map[string]int {
	wc := make(map[string]int)
	for _, w := range strings.Fields(s) {
		wc[w]++
	}
	return wc
}

func main() {
	wc.Test(WordCount)
}