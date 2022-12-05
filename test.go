package main

import (
	"fmt"
)

type session struct {
	Id   string
	Name string
}

func main() {
	var t map[string]session
	if t["2"] == (session{}) {
		fmt.Println(t["1"])
	}
}
