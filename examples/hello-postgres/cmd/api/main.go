package main

import (
	"log"

	_ "github.com/lib/pq"

	"github.com/go-modkit/modkit/examples/hello-postgres/internal/httpserver"
	mkhttp "github.com/go-modkit/modkit/modkit/http"
)

func main() {
	handler, err := httpserver.BuildHandler()
	if err != nil {
		log.Fatalf("build handler: %v", err)
	}

	log.Println("Server starting on http://localhost:8080")
	log.Println("Try: curl http://localhost:8080/api/v1/health")
	if err := mkhttp.Serve(":8080", handler); err != nil {
		log.Fatalf("serve: %v", err)
	}
}
