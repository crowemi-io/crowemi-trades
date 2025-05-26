package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/crowemi-io/crowemi-go-utils/db"
	"github.com/crowemi-io/crowemi-go-utils/log"
	trader "github.com/crowemi-io/crowemi-trades"
	"github.com/crowemi-io/crowemi-trades/api"
)

func main() {

	config, err := trader.Bootstrap()
	if err != nil {
		// TODO: clean logging
		println("error")
	}
	config.Logger.Log("crowemi-trades start", log.INFO, nil, "main.main")

	// TODO: create go routine for trade updates
	go func() {}()
	// TODO: create go routine for market dater
	go func() {}()
	// TODO: get activties
	go func() {
		// read config from gcp
		var token string
		lastActivity, err := db.GetOne[trader.Activities](context.Background(), config.MongoClient, "activities", nil, []db.MongoSort{{Field: "_id", Direction: -1}})
		if err != nil {
			config.Logger.Log(fmt.Sprintf("failed to get last activity: %e", err), log.ERROR, nil, "main.main")
		}

		if lastActivity.ID != "" {
			token = lastActivity.ID
		}

		activities, err := config.AlpacaClient.GetActivities(token)
		if err != nil {
			config.Logger.Log(fmt.Sprintf("failed to get activities: %e", err), log.ERROR, nil, "main.main")
		}

		res, err := db.InsertMany(context.Background(), config.MongoClient, "activities", activities)
		if err != nil {
			config.Logger.Log(fmt.Sprintf("failed to insert activities: %e", err), log.ERROR, nil, "main.main")
		}
		config.Logger.Log(fmt.Sprintf("inserted %d activities", res), log.INFO, nil, "main.main")

		var buffer bytes.Buffer
		for _, activity := range activities {
			dater, err := json.Marshal(activity)
			if err != nil {
				config.Logger.Log(fmt.Sprintf("failed to marshal json activity: %e", err), log.ERROR, activity, "main.main")
			}
			buffer.Write(dater)
			buffer.WriteString("\n")

		}

		_, err = config.GcpClient.Write(fmt.Sprintf("%s/activities/%s.json", config.Crowemi.Env, time.Now().Format("2006-01-02T15:04:05Z07:00")), buffer.Bytes())
		if err != nil {
			config.Logger.Log("failed writing to cloud storage", log.ERROR, nil, "main.main")
		}
		_, err = config.GcpClient.Write(fmt.Sprintf("%s/raw/activities/%s.json", config.Crowemi.Env, time.Now().Format("2006-01-02T15:04:05Z07:00")), buffer.Bytes())
		if err != nil {
			config.Logger.Log("failed writing to cloud storage", log.ERROR, nil, "main.main")
		}

		if config.Crowemi.Debug {
			_ = os.WriteFile("test.json", buffer.Bytes(), 0644)
		}

		config.Logger.Log(fmt.Sprintf("new activities %d", len(activities)), log.INFO, nil, "main.main")
		print(activities)
	}()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.Split(r.RequestURI, "/")
		h := api.HandlerFactory(path[1], config)
		if h == nil {
			// set return code, so something better here
			w.Write([]byte("Invalid path"))
		}
		switch r.Method {
		case http.MethodGet:
			h.Get(w, r)
		case http.MethodDelete:
			h.Delete(w, r)
		case http.MethodPost:
			h.Post(w, r)
		case http.MethodPut:
			h.Put(w, r)
		// add more methods
		default:
			// set return code, so something better here
			w.Write([]byte("Invalid path"))
		}
	})
	http.ListenAndServe(":8004", nil)
}
