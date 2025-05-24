package main

import (
	"log"

	"gitlab.com/Hamed1984/stats/pkg/app"
	"gitlab.com/Hamed1984/stats/pkg/conf"
)

func main() {
	app, err := app.NewApplication(conf.GetConffiguration())
	if err != nil {
		log.Fatal(err)
	}

	err = app.Start()
	if err!=nil {
		log.Fatal(err)
	}
}
