package repository

import (
	"fmt"
	"io"
	"strings"
	goTemplate "text/template"

	"github.com/yoyo-project/yoyo/internal/repository/template"
	"github.com/yoyo-project/yoyo/internal/schema"
)

type RepositoryParams struct {
	ExportedGoName   string
	QueryPackageName string
	Table            schema.Table

	PackageName   string
	PKNames       []string
	InsertColumns []string
	SelectColumns []string
	ScanFields    []string
	InFields      []string
	PKFields      []string

	QueryImportPath   string
	ColumnAssignments []string

	PKCapture string
	PKQuery   string

	StatementPlaceholders []string
}

func NewEntityRepositoryGenerator(packageName string, adapter Adapter, reposPath string, packagePath Finder, db schema.Database) EntityGenerator {
	return func(t schema.Table, w io.Writer) (err error) {
		ps := RepositoryParams{
			ExportedGoName:   t.ExportedGoName(),
			QueryPackageName: t.QueryPackageName(),
			Table:            t,
			PackageName: packageName,
		}

		for _, col := range t.Columns {
			if col.PrimaryKey {
				ps.PKFields = append(ps.PKFields, strings.ReplaceAll(template.PKFieldTemplate, template.FieldName, col.ExportedGoName()))
				ps.PKNames = append(ps.PKNames, col.Name)
			}
			if !col.AutoIncrement {
				ps.InsertColumns = append(ps.InsertColumns, col.Name)
			}
			ps.SelectColumns = append(ps.SelectColumns, col.Name)
			ps.ScanFields = append(ps.ScanFields, fmt.Sprintf("&ent.%s", col.ExportedGoName()))
			ps.InFields = append(ps.InFields, fmt.Sprintf("in.%s", col.ExportedGoName()))
		}

		for _, r := range t.References {
			if r.HasOne {
				ft, _ := db.GetTable(r.TableName)
				for _, cn := range r.ColNames(ft) {
					ps.SelectColumns = append(ps.SelectColumns, cn)
					ps.InsertColumns = append(ps.InsertColumns, cn)
				}
				for _, cn := range ft.PKColNames() {
					c, _ := ft.GetColumn(cn)
					goName := fmt.Sprintf("%s%s", ft.ExportedGoName(), c.ExportedGoName())
					ps.ScanFields = append(ps.ScanFields, fmt.Sprintf("&ent.%s", goName))
					ps.InFields = append(ps.InFields, fmt.Sprintf("in.%s", goName))
				}
			}
		}

		for _, t2 := range db.Tables {
			for _, r := range t2.References {
				if r.HasMany && r.TableName == t.Name {
					for _, col := range t2.PKColumns() {
						ps.SelectColumns = append(ps.SelectColumns, col.Name)
						ps.InsertColumns = append(ps.InsertColumns, col.Name)
						ps.ScanFields = append(ps.ScanFields, fmt.Sprintf("&ent.%s", t2.ExportedGoName()+col.ExportedGoName()))
						ps.InFields = append(ps.InFields, fmt.Sprintf("in.%s", col.ExportedGoName()))
					}
				}
			}
		}

		ps.QueryImportPath, err = packagePath(fmt.Sprintf("%s/query/%s", reposPath, t.QueryPackageName()))
		if err != nil {
			return fmt.Errorf("unable to generate repository: %w", err)
		}

		var pkCapTemplate string
		pkReplacer := strings.NewReplacer()

		switch len(t.PKColumns()) {
		case 0:
			// Do nothing
		case 1:
			col := t.PKColumns()[0]
			switch col.AutoIncrement {
			case true:
				pkCapTemplate = template.SinglePKCaptureTemplate
			case false:
				pkCapTemplate = template.NoPKCapture
			}
			pkReplacer = strings.NewReplacer(
				template.FieldName,
				col.ExportedGoName(),
				template.Type,
				col.GoTypeString(),
			)
		default:
			pkCapTemplate = template.MultiPKCaptureTemplate
			pkReplacer = strings.NewReplacer()
		}

		ps.PKCapture = pkReplacer.Replace(pkCapTemplate)

		pkQueryReplacer := strings.NewReplacer(
			template.QueryPackageName,
			t.QueryPackageName(),
			template.PKFields,
			strings.Join(ps.PKFields, "\n		"),
		)

		ps.PKQuery = pkQueryReplacer.Replace(template.PKQueryTemplate)

		ps.StatementPlaceholders = adapter.PreparedStatementPlaceholders(len(ps.SelectColumns))
		for i, colName := range ps.SelectColumns {
			ps.ColumnAssignments = append(ps.ColumnAssignments, fmt.Sprintf("%s = %s", colName, ps.StatementPlaceholders[i]))
		}

		tpl := goTemplate.Must(
			goTemplate.New("RepositoryFile").
				Funcs(goTemplate.FuncMap{"join": Join}).
				Parse(template.RepositoryFile),
		)
		err = tpl.Execute(w, ps)

		return err
	}
}

func Join(d string, ss []string) string {
	return strings.Join(ss, d)
}
