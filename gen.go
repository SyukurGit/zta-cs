package main

import (
	"fmt"
	"github.com/syukurgit/zta/pkg/utils"
)

func main() {
	hash, _ := utils.HashPassword("password123")
	fmt.Println(hash)
}
