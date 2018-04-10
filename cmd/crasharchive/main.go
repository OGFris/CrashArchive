package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/pmmp/CrashArchive/app"
	"github.com/pmmp/CrashArchive/app/database"
	"github.com/pmmp/CrashArchive/app/handler"
	"github.com/pmmp/CrashArchive/app/router"
	"github.com/pmmp/CrashArchive/app/view"
	"github.com/pmmp/CrashArchive/app/webhook"
)

const dbRetry = 5

func main() {
	log.SetFlags(log.Lshortfile)

	configPath := flag.String("c", "config.json", "path to `config.json`")
	flag.Parse()

	var err error
	config, err := app.LoadConfig(*configPath)
	if err != nil {
		log.Printf("unable to load config: %v", err)
		os.Exit(1)
	}

	view, err := view.Load(config.Templates)
	if err != nil {
		log.Fatal(err)
	}

	db, err := database.New(config.Database)
	if err != nil {
		log.Fatal(fmt.Errorf("database error: %v", err))
	}

	c := &handler.Common{
		DB:      db,
		WH:      webhook.New(config.SlackURL),
		View:    view,
		Assets:  config.Assets,
		Reports: config.Reports,
	}

	r := router.New(c)
	log.Printf("listening on: %s\n", config.ListenAddress)
	if err = http.ListenAndServe(config.ListenAddress, r); err != nil {
		log.Fatal(err)
	}

}
