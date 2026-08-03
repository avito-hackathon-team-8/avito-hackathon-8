package main

import (
	"log"
	"net/http"
	"time"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
)

func main() {
	app := pocketbase.New()

	app.OnServe().BindFunc(func(event *core.ServeEvent) error {
		event.Router.GET("/api/app/status", func(request *core.RequestEvent) error {
			return request.JSON(http.StatusOK, map[string]string{
				"service": "pocketbase-go",
				"status":  "ready",
				"time":    time.Now().UTC().Format(time.RFC3339),
			})
		})

		return event.Next()
	})

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
}
