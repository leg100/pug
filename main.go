package main

import (
	"fmt"
	"os"

	"github.com/leg100/pug/internal/app"
)

func main() {
	if err := app.Start(os.Stdout, os.Stderr, os.Args[1:]); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}
