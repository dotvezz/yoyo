package schema

import (
	"reflect"
	"testing"
)

func TestReference_ColNames(t *testing.T) {
	type fields struct {
		ColumnNames []string
	}
	type args struct {
		fTable Table
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []string
	}{
		{
			name: "single column from table",
			args: args{fTable: Table{
				Name:    "ftable",
				Columns: []Column{{Name: "id", PrimaryKey: true}},
			}},
			want: []string{"fk_ftable_id"},
		},
		{
			name: "two columns from table",
			args: args{fTable: Table{
				Name: "ftable",
				Columns: []Column{
					{Name: "id", PrimaryKey: true},
					{Name: "id2", PrimaryKey: true},
				},
			}},
			want: []string{"fk_ftable_id", "fk_ftable_id2"},
		},
		{
			name: "single id from two column table",
			args: args{fTable: Table{
				Name: "ftable",
				Columns: []Column{
					{Name: "id", PrimaryKey: true},
					{Name: "no"},
				},
			}},
			want: []string{"fk_ftable_id"},
		},
		{
			name:   "single column explicitly declared",
			fields: fields{ColumnNames: []string{"ftable_id"}},
			args: args{fTable: Table{
				Name: "ftable",
				Columns: []Column{
					{Name: "id", PrimaryKey: true},
					{Name: "no"},
				},
			}},
			want: []string{"ftable_id"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Reference{
				ColumnNames: tt.fields.ColumnNames,
			}
			if got := r.ColNames(tt.args.fTable); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ColNames()\nwant %#v\n got %#v", tt.want, got)
			}
		})
	}
}
