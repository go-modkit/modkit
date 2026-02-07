package {{.Package}}

import (
	"net/http"

	mkhttp "github.com/go-modkit/modkit/modkit/http"
)

// {{.Name | Title}}Controller handles {{.Name}} related requests.
type {{.Name | Title}}Controller struct{}

func New{{.Name | Title}}Controller() *{{.Name | Title}}Controller {
	return &{{.Name | Title}}Controller{}
}

func (c *{{.Name | Title}}Controller) RegisterRoutes(r mkhttp.Router) {
	// TODO: Register routes
	// r.Handle(http.MethodGet, "/{{.Name}}", http.HandlerFunc(c.HandleGet))
}
