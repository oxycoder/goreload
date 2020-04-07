package main

import (
	"fmt"
	"time"
)

func main() {
	fmt.Println("Hello world, I will be changed")
	ch := make(chan int)
	time.AfterFunc(5*time.Second, func() {
		fmt.Println("5 seconds passed")
		ch <- 10
	})
	for {
		select {
		case i := <-ch:
			fmt.Println(i, " is coming, end.")
			return
		default:
			fmt.Println("wait")
			time.Sleep(1 * time.Second)
		}
	}
}
