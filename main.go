package main

import (
	"ahmedash95/pedis/pedis"
	"fmt"
)

func main() {
	fmt.Println("Listening on port 6379")

	server := pedis.NewServer()
	server.ListenAndServe("0.0.0.0:6379")
}
