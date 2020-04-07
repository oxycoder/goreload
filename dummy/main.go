package main

import (
	"fmt"
	"time"
)

func main() {
	fmt.Println("Hello world, I will be changed")
	time.AfterFunc(3000, func() {
		fmt.Println("Yah")
	})
}
