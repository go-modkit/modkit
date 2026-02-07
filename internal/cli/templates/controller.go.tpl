package {{.Package}}

import (
	mkhttp "github.com/go-modkit/modkit/modkit/http"
)

// {{.Identifier}}Controller handles {{.Name}} related requests.
type {{.Identifier}}Controller struct{}

func New{{.Identifier}}Controller() *{{.Identifier}}Controller {
	return &{{.Identifier}}Controller{}
}

func (c *{{.Identifier}}Controller) RegisterRoutes(r mkhttp.Router) {
	// TODO: Register routes
	// r.Handle("GET", "/{{.Name}}", ...)
}
