package main

import (
	"log"

	mkhttp "github.com/go-modkit/modkit/modkit/http"
	"github.com/go-modkit/modkit/modkit/kernel"

	"{{.Name}}/internal/modules/app"
)

func main() {
	// Bootstrap the app
	appInstance, err := kernel.Bootstrap(&app.AppModule{})
	if err != nil {
		log.Fatal(err)
	}

	// Create router and register controllers
	router := mkhttp.NewRouter()
	if err := mkhttp.RegisterRoutes(mkhttp.AsRouter(router), appInstance.Controllers); err != nil {
		log.Fatal(err)
	}

	// Start server
	log.Println("Listening on :8080")
	if err := mkhttp.Serve(":8080", router); err != nil {
		log.Fatal(err)
	}
}
