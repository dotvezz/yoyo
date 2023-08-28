package repository

import (
	"fmt"
	"io"
	goTemplate "text/template"

	"github.com/yoyo-project/yoyo/internal/repository/template"
	"github.com/yoyo-project/yoyo/internal/schema"
)

type EntityFileParams struct {
	EntityFields    []string
	Fields          []string
	Imports         []string
	ReferenceFields []string
	EntityName      string
	PackageName     string
}

func NewEntityGenerator(packageName string, db schema.Database, packagePath Finder, reposPath string) EntityGenerator {
	return func(t schema.Table, w io.Writer) error {
		ps := EntityFileParams{
			PackageName: packageName,
			EntityName:  t.ExportedGoName(),
		}
		nullPackagePath, err := packagePath(reposPath + "/nullable")
		if err != nil {
			return fmt.Errorf("couldn't generate entity file: %w", err)
		}
		for _, c := range t.Columns {
			ps.EntityFields = append(ps.EntityFields, fmt.Sprintf("%s %s", c.ExportedGoName(), c.GoTypeString()))
			ps.Fields = append(ps.Fields, c.ExportedGoName())
			if imp := c.RequiredImport(nullPackagePath); imp != "" {
				ps.Imports = append(ps.Imports, imp)
			}
		}

		for _, r := range t.References {
			if r.HasOne {
				ft, _ := db.GetTable(r.TableName)
				for _, cn := range ft.PKColNames() {
					c, _ := ft.GetColumn(cn)

					goName := fmt.Sprintf("%s%s", ft.ExportedGoName(), c.ExportedGoName())
					ps.Fields = append(ps.Fields, goName)
					ps.ReferenceFields = append(ps.ReferenceFields, fmt.Sprintf("%s %s", goName, c.GoTypeString()))
				}
			}
		}

		for _, t2 := range db.Tables {
			for _, r := range t2.References {
				if r.HasMany && r.TableName == t.Name {
					for _, c := range t2.PKColumns() {
						ps.Fields = append(ps.Fields, t2.ExportedGoName()+c.ExportedGoName())
						ps.ReferenceFields = append(ps.ReferenceFields, fmt.Sprintf("%s %s", t2.ExportedGoName()+c.ExportedGoName(), c.GoTypeString()))
					}
				}
			}
		}

		ps.Imports = sortedUnique(ps.Imports)

		tpl := goTemplate.Must(goTemplate.New("EntityFile").Parse(template.EntityFile))

		err = tpl.Execute(w, ps)

		return err
	}
}
