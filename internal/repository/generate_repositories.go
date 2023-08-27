package repository

import (
	"io"
	goTemplate "text/template"

	"github.com/yoyo-project/yoyo/internal/repository/template"

	"github.com/yoyo-project/yoyo/internal/schema"
)

type RepositoriesFileParams struct {
	schema.Database
	PackageName string
}

func NewRepositoriesGenerator(packageName string) WriteGenerator {
	return func(db schema.Database, w io.Writer) (err error) {
		ps := RepositoriesFileParams{
			Database:    db,
			PackageName: packageName,
		}
		tpl := goTemplate.Must(goTemplate.New("RepositoriesFile").Parse(template.RepositoriesFile))
		err = tpl.Execute(w, ps)

		return err
	}
}
