package main

import (
	"log"
	"testing"
)

func TestCounter(t *testing.T) {
	counter := newTreeTopicViewCounter()
	counter.Bump(1)
	counter.Bump(57)
	counter.Bump(58)
	counter.Bump(59)
	counter.Bump(9)
}

func TestScope(t *testing.T) {
	var outVar int
	closureHolder := func() {
		outVar = 2
	}
	closureHolder()
	log.Print("outVar: ", outVar)
}
