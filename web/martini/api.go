package main

import (
	"time"
)

func index() (int, string) {
	time.Sleep(200 * time.Millisecond)
	return 200, "This is index."
}
