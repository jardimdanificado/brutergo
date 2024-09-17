package main

import "fmt"

type VirtualMachine struct {

}

func main() {
	var teste string = "Hello, World!"
	fmt.Println(teste)
	// var i int = 1
	for i := 0;i <= 3;i++ {
        fmt.Println(i)
        i = i + 1
    }
}