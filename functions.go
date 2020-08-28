package main

import "fmt"

func add(x int, y int) int {
	return x + y
}

func name(n string) string {
	return "Your name is:" + n
}

func main() {
	fmt.Println(name("Ngoan"))
	fmt.Println(add(42, 13))
}

