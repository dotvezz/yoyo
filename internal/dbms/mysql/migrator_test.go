package mysql

import (
	"github.com/dotvezz/yoyo/internal/datatype"
	"github.com/dotvezz/yoyo/internal/dbms/base"
	"github.com/dotvezz/yoyo/internal/dbms/dialect"
	"github.com/dotvezz/yoyo/internal/schema"
	"reflect"
	"strings"
	"testing"
)

func TestNewMigrator(t *testing.T) {
	tests := []struct {
		name string
		want *migrator
	}{
		{
			name: "just a migrator",
			want: &migrator{
				Base: base.Base{
					Dialect: dialect.MySQL,
				},
				validator: validator{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewMigrator(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewMigrator() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_migrator_TypeString(t *testing.T) {
	tests := map[string]struct {
		dt      datatype.Datatype
		wantS   string
		wantErr string
	}{
		datatype.Integer.String(): {
			dt:    datatype.Integer,
			wantS: "INT",
		},
		datatype.BigInt.String(): {
			dt:    datatype.BigInt,
			wantS: "BIGINT",
		},
		datatype.SmallInt.String(): {
			dt:    datatype.SmallInt,
			wantS: "SMALLINT",
		},
		"unsupported datatype": {
			dt:      datatype.Boolean,
			wantErr: "unsupported datatype",
		},
		"invalid datatype": {
			dt:      0,
			wantErr: "invalid datatype",
		},
	}

	m := &migrator{
		Base:      base.Base{Dialect: dialect.MySQL},
		validator: validator{},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			gotS, err := m.TypeString(tt.dt)
			if tt.wantErr == "" {
				if err != nil {
					t.Errorf("expected error `nil`, got error `%v`", err)
				} else if gotS != tt.wantS {
					t.Errorf("expected string `%s`, got string `%s`", tt.wantS, gotS)
				}
			}

			if tt.wantErr != "" {
				if err == nil {
					t.Errorf("expected error `%v`, got error `nil`", tt.wantErr)
				} else if !strings.Contains(err.Error(), tt.wantErr) {
					t.Errorf("expected error `%v`, got error `%v`", tt.wantErr, err)
				}
			}
		})
	}
}

func Test_migrator_CreateTable(t *testing.T) {
	tests := map[string]struct {
		tName string
		t     schema.Table
		wantS string
	}{
		"empty table": {
			tName: "table",
			wantS: "CREATE TABLE `table` (\n\n);",
		},
		"single column no primary key": {
			tName: "table",
			t: schema.Table{
				Columns: map[string]schema.Column{
					"column": {
						Datatype: datatype.Integer,
					},
				},
			},
			wantS: "CREATE TABLE `table` (\n    `column` INT SIGNED NOT NULL\n);",
		},
		"two column with primary key": {
			tName: "table",
			t: schema.Table{
				Columns: map[string]schema.Column{
					"column": {
						Datatype:   datatype.Integer,
						PrimaryKey: true,
					},
					"column2": {
						Datatype: datatype.Integer,
					},
				},
			},
			wantS: "CREATE TABLE `table` (\n" +
				"    `column` INT SIGNED NOT NULL,\n" +
				"    `column2` INT SIGNED NOT NULL\n" +
				"    PRIMARY KEY (`column`)\n);",
		},
	}

	m := &migrator{
		Base:      base.Base{Dialect: dialect.MySQL},
		validator: validator{},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			gotS := m.CreateTable(tt.tName, tt.t)
			if gotS != tt.wantS {
				t.Errorf("expected string `%s`, got string `%s`", tt.wantS, gotS)
			}
		})
	}
}

func Test_migrator_AddColumn(t *testing.T) {
	tests := map[string]struct {
		tName string
		cName string
		c     schema.Column
		wantS string
	}{
		"basic int column": {
			tName: "table",
			cName: "column",
			c: schema.Column{
				Datatype: datatype.Integer,
			},
			wantS: "ALTER TABLE `table` ADD COLUMN `column` INT SIGNED NOT NULL;",
		},
	}

	m := &migrator{
		Base:      base.Base{Dialect: dialect.MySQL},
		validator: validator{},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			gotS := m.AddColumn(tt.tName, tt.cName, tt.c)
			if gotS != tt.wantS {
				t.Errorf("expected string `%s`, got string `%s`", tt.wantS, gotS)
			}
		})
	}
}

func Test_migrator_AddIndex(t *testing.T) {
	tests := map[string]struct {
		tName string
		iName string
		i     schema.Index
		wantS string
	}{
		"non-unique single-column": {
			tName: "table",
			iName: "foreign",
			i: schema.Index{
				Columns: []string{"col"},
				Unique:  false,
			},
			wantS: "ALTER TABLE `table` ADD INDEX `foreign` (`col`);",
		},
		"non-unique two columns": {
			tName: "table",
			iName: "foreign",
			i: schema.Index{
				Columns: []string{"col", "col2"},
				Unique:  false,
			},
			wantS: "ALTER TABLE `table` ADD INDEX `foreign` (`col`, `col2`);",
		},
		"unique single-column": {
			tName: "table",
			iName: "foreign",
			i: schema.Index{
				Columns: []string{"col"},
				Unique:  true,
			},
			wantS: "ALTER TABLE `table` ADD UNIQUE INDEX `foreign` (`col`);",
		},
		"unique two columns": {
			tName: "table",
			iName: "foreign",
			i: schema.Index{
				Columns: []string{"col", "col2"},
				Unique:  true,
			},
			wantS: "ALTER TABLE `table` ADD UNIQUE INDEX `foreign` (`col`, `col2`);",
		},
	}

	m := &migrator{
		Base:      base.Base{Dialect: dialect.MySQL},
		validator: validator{},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			gotS := m.AddIndex(tt.tName, tt.iName, tt.i)
			if gotS != tt.wantS {
				t.Errorf("expected string `%s`, got string `%s`", tt.wantS, gotS)
			}
		})
	}
}

func Test_migrator_generateColumn(t *testing.T) {
	point := func(s string) *string {
		return &s
	}
	tests := map[string]struct {
		cName string
		c     schema.Column
		wantS string
	}{
		"int": {
			cName: "col",
			c: schema.Column{
				Datatype: datatype.Integer,
			},
			wantS: "`col` INT SIGNED NOT NULL",
		},
		"nullable int": {
			cName: "col",
			c: schema.Column{
				Datatype: datatype.Integer,
				Nullable: true,
			},
			wantS: "`col` INT SIGNED DEFAULT NULL NULL",
		},
		"int default 1": {
			cName: "col",
			c: schema.Column{
				Datatype: datatype.Integer,
				Default:  point("1"),
			},
			wantS: "`col` INT SIGNED DEFAULT 1 NOT NULL",
		},
		"unsigned int": {
			cName: "col",
			c: schema.Column{
				Datatype: datatype.Integer,
				Unsigned: true,
			},
			wantS: "`col` INT UNSIGNED NOT NULL",
		},
		"int auto_increment": {
			cName: "col",
			c: schema.Column{
				Datatype:      datatype.Integer,
				PrimaryKey:    true,
				AutoIncrement: true,
			},
			wantS: "`col` INT SIGNED NOT NULL AUTO_INCREMENT",
		},
		"decimal": {
			cName: "col",
			c: schema.Column{
				Datatype:  datatype.Decimal,
				Scale:     6,
				Precision: 4,
			},
			wantS: "`col` DECIMAL(6, 4) SIGNED NOT NULL",
		},
		"text default blah": {
			cName: "col",
			c: schema.Column{
				Datatype: datatype.Text,
				Default:  point("blah"),
			},
			wantS: "`col` TEXT DEFAULT \"blah\" NOT NULL",
		},
		"varchar": {
			cName: "col",
			c: schema.Column{
				Datatype: datatype.Varchar,
			},
			wantS: "`col` VARCHAR NOT NULL",
		},
		"sized varchar": {
			cName: "col",
			c: schema.Column{
				Datatype: datatype.Varchar,
				Scale:    64,
			},
			wantS: "`col` VARCHAR(64) NOT NULL",
		},
		"sized hypothetical thing": {
			cName: "col",
			c: schema.Column{
				Datatype:  0,
				Scale:     64,
				Precision: 43,
			},
			wantS: "`col` (64, 43) NOT NULL",
		},
	}

	m := &migrator{
		Base:      base.Base{Dialect: dialect.MySQL},
		validator: validator{},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			gotS := m.generateColumn(tt.cName, tt.c)
			if gotS != tt.wantS {
				t.Errorf("expected string `%s`, got string `%s`", tt.wantS, gotS)
			}
		})
	}
}
