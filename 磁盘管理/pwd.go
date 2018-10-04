package main

import (
	"fmt"
	"os"
)

func main() {
	path, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(path)
}