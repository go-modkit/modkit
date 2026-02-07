package {{.Package}}

import (
	"github.com/go-modkit/modkit/modkit/module"
)

// {{.Name | Title}}Service is the {{.Name}} service.
type {{.Name | Title}}Service struct{}

func New{{.Name | Title}}Service() *{{.Name | Title}}Service {
	return &{{.Name | Title}}Service{}
}
