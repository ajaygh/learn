package main

import (
	"fmt"
	"math/rand"
	"time"
)

func f(n int) {
	for i := 0; i < 10; i++ {
		fmt.Println(n, " : ", i)
		amt := time.Duration(rand.Intn(300))
		time.Sleep(time.Millisecond * amt)
	}
}

func main() {
	for i := 0; i < 10; i++ {
		go f(i)
	}
	var s string
	fmt.Scanln(&s)
	fmt.Println(&s)
}
