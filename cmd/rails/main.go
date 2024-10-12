package main

import (
	"fmt"
	"os"
	"rails/internal/app"
	"rails/internal/utils"
)

func main() {
	if utils.ContainsAny([]string{"-v", "-version"}, os.Args) {
		fmt.Println("Rails:", utils.Version)
		return
	}
	var server app.Server
	server.Start()
}
