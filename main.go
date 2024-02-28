package main

import (
	"fmt"
	"os"

	"github.com/leg100/pug/internal/app"
)

func main() {
	if err := app.Start(); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}
