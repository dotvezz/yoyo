package template

import (
	_ "embed"
)

//go:embed null_types.gotpl
var NullTypeFile string

//go:embed entity.gotpl
var EntityFile string

//go:embed node.gotpl
var NodeFile string

//go:embed query_file.gotpl
var QueryFile string

//go:embed repositories_file.gotpl
var RepositoriesFile string

//go:embed repository.gotpl
var RepositoryFile string