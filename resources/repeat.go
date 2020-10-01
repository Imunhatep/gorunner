package main

import "fmt"
import "time"

func main() {
	for {
		time.Sleep(1 * time.Second)
		fmt.Println("next2")
	}
}
