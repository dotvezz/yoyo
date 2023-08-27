package repository

import (
	"fmt"
	"io"
	"sort"
	goTemplate "text/template"

	_ "embed"
	"github.com/yoyo-project/yoyo/internal/repository/template"
	"github.com/yoyo-project/yoyo/internal/schema"
)

func NewQueryFileGenerator(reposPath string, findPackagePath Finder, db schema.Database) EntityGenerator {
	return func(t schema.Table, w io.Writer) error {
		// We always need fmt because we use it for Query.SQL()
		imports := []string{`"fmt"`}

		ps := QueryFileParams{}
		for _, c := range t.Columns {
			ops, is := buildOptsAndImports(c)
			ps.Columns = append(ps.Columns, ColumnParams{
				Column:     c,
				Operations: ops,
			})

			imports = append(imports, is...)
		}

		for _, r := range t.References {
			if r.HasMany {
				continue // Skip HasMany references
			}

			ft, ok := db.GetTable(r.TableName)
			if !ok {
				return fmt.Errorf("unable to generate queries for table %s, missing foreign table %s", t.Name, r.TableName)
			}

			for i, n := range r.ColNames(ft) {
				c := ft.PKColumns()[i]
				// Override the GoName in order to generate correct method/function names
				c.GoName = r.ExportedGoName() + c.ExportedGoName()
				// Override the name - use the fk name
				c.Name = n
				ops, is := buildOptsAndImports(c)
				imports = append(imports, is...)

				ps.Columns = append(ps.Columns, ColumnParams{
					Column:     c,
					Operations: ops,
				})
			}
		}

		var err error
		ps.Imports = sortedUnique(imports)
		ps.RepositoriesPackage, err = findPackagePath(reposPath + "/")
		ps.PackageName = t.QueryPackageName()
		if err != nil {
			return fmt.Errorf("unable to generate query file: %w", err)
		}

		tpl := goTemplate.Must(goTemplate.New("QueryFile").Parse(template.QueryFile))
		err = tpl.Execute(w, ps)
		return err
	}
}

const (
	Equals         = "Equals"
	Not            = "Not"
	Contains       = "Contains"
	ContainsNot    = "ContainsNot"
	StartsWith     = "StartsWith"
	StartsWithNot  = "StartsWithNot"
	EndsWith       = "EndsWith"
	EndsWithNot    = "EndsWithNot"
	GreaterThan    = "GreaterThan"
	GreaterOrEqual = "GreaterOrEqual"
	LessThan       = "LessThan"
	LessOrEqual    = "LessOrEqual"
	Before         = "Before"
	BeforeOrEqual  = "BeforeOrEqual"
	After          = "After"
	AfterOrEqual   = "AfterOrEqual"

	IsNull    = "IsNull"
	IsNotNull = "IsNotNull"
)

func buildOptsAndImports(column schema.Column) (operations []Operation, imports []string) {
	var (
		ops []Operation
	)
	switch {
	case column.Datatype.IsTime():
		ops = []Operation{
			{Name: Equals},
			{Name: Not},
			{Name: Before},
			{Name: After},
			{Name: BeforeOrEqual},
			{Name: AfterOrEqual},
		}
	case column.Datatype.IsNumeric():
		ops = []Operation{
			{Name: Equals},
			{Name: Not},
			{Name: GreaterThan},
			{Name: LessThan},
			{Name: GreaterOrEqual},
			{Name: LessOrEqual},
		}
	case column.Datatype.IsString():
		ops = []Operation{
			{Name: Equals},
			{Name: Not},
			{Name: Contains},
			{Name: ContainsNot},
			{Name: StartsWith},
			{Name: StartsWithNot},
			{Name: EndsWith},
			{Name: EndsWithNot},
		}
	case column.Datatype.IsBinary():
		ops = []Operation{
			{Name: Equals},
			{Name: Not},
		}
	}

	if column.Nullable {
		ops = append(ops, Operation{Name: IsNull, NullCheck: true}, Operation{Name: IsNotNull, NullCheck: true})
	}

	for _, op := range ops {
		imports = append(imports, op.imports()...)
	}

	return ops, imports
}

func DeDup[T comparable](in []T) (out []T) {
	exists := make(map[T]bool)
	for _, v := range in {
		if _, ok := exists[v]; !ok {
			exists[v] = true
			out = append(out, v)
		}
	}

	return out
}

type QueryFileParams struct {
	Columns             []ColumnParams
	RepositoriesPackage string
	PackageName         string
	Imports             []string
}

type ColumnParams struct {
	schema.Column
	Operations []Operation
}


type Operation struct {
	Name      string
	NullCheck bool
}

func (o Operation) funcName(fieldName string) string {
	if o.Name == Equals {
		return fieldName
	}
	return fmt.Sprintf("%s%s", fieldName, o.Name)
}

func (o Operation) Val() string {
	switch o.Name {
	case Contains, ContainsNot:
		return `fmt.Sprintf("'%%%s%%'", val)`
	case StartsWith, StartsWithNot:
		return `fmt.Sprintf("'%s%%'", val)`
	case EndsWith, EndsWithNot:
		return `fmt.Sprintf("'%%%s'", val)`
	case IsNull, IsNotNull:
		return `nil`
	default:
		return "val"
	}
}

func (o Operation) Operator() (operator string) {
	switch o.Name {
	case Contains:
		operator = "Like"
	case ContainsNot:
		operator = "NotLike"
	case StartsWith:
		operator = "Like"
	case StartsWithNot:
		operator = "NotLike"
	case EndsWith:
		operator = "Like"
	case EndsWithNot:
		operator = "NotLike"
	case Not:
		operator = "NotEquals"
	default:
		operator = o.Name
	}
	return operator
}

func (o Operation) imports() (imports []string) {
	switch o.Name {
	case Contains, ContainsNot, StartsWith, StartsWithNot, EndsWith, EndsWithNot:
		imports = append(imports, `"fmt"`)
	case Before, After, BeforeOrEqual, AfterOrEqual:
		imports = append(imports, `"time"`)
	}
	return imports
}

func sortedUnique(in []string) (out []string) {
	m := make(map[string]bool)
	for i := range in {
		if _, ok := m[in[i]]; ok {
			continue
		}
		out = append(out, in[i])
		m[in[i]] = true
	}
	sort.Strings(out)
	return out
}
