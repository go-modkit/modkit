module github.com/go-modkit/modkit/examples/hello-sqlite

go 1.25.7

require (
	github.com/go-modkit/modkit v0.0.0
	github.com/mattn/go-sqlite3 v1.14.22
)

require github.com/go-chi/chi/v5 v5.2.4 // indirect

replace github.com/go-modkit/modkit => ../..
