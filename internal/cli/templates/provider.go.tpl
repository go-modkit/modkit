package {{.Package}}

// {{.Identifier}}Service is the {{.Name}} service.
type {{.Identifier}}Service struct{}

func New{{.Identifier}}Service() *{{.Identifier}}Service {
	return &{{.Identifier}}Service{}
}
