package main

import (
	"fmt"

	"github.com/tormath1/gotube/lib"
)

func main() {
	if err := lib.ImportVideo(
		"https://www.youtube.com/watch?v=123456789",
		"/tmp/videos",
	); err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("audio downloaded")
}
