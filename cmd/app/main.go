package main

import (
	"log"
	"os"

	"github.com/wanomir/keychron-tg-parser/internal/app"
)

func main() {
	a, err := app.NewApp()
	if err != nil {
		log.Fatal(err)
	}

	os.Exit(a.Run())
}
