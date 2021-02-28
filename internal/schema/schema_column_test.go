package schema

import (
	"github.com/yoyo-project/yoyo/internal/datatype"
	"testing"
)

func TestColumn_ExportedGoName(t *testing.T) {
	type fields struct {
		Name   string
		GoName string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "with Name only",
			fields: fields{
				Name: "column",
			},
			want: "Column",
		},
		{
			name: "with GoName only",
			fields: fields{
				GoName: "column",
			},
			want: "Column",
		},
		{
			name: "with GoName and Name",
			fields: fields{
				Name:   "no",
				GoName: "yes",
			},
			want: "Yes",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Column{
				Name:   tt.fields.Name,
				GoName: tt.fields.GoName,
			}
			if got := c.ExportedGoName(); got != tt.want {
				t.Errorf("ExportedGoName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestColumn_GoTypeString(t *testing.T) {
	type fields struct {
		Datatype datatype.Datatype
		Unsigned bool
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name:   "int",
			fields: fields{Datatype: datatype.Integer, Unsigned: true},
			want:   "int32",
		},
		{
			name:   "unsigned int",
			fields: fields{Datatype: datatype.Integer},
			want:   "uint32",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Column{
				Datatype: tt.fields.Datatype,
				Unsigned: tt.fields.Unsigned,
			}
			if got := c.GoTypeString(); got != tt.want {
				t.Errorf("GoTypeString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestColumn_RequiredImport(t *testing.T) {
	type fields struct {
		Datatype datatype.Datatype
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name:   "int",
			fields: fields{Datatype: datatype.Integer},
			want:   "",
		},
		{
			name:   "timestamp",
			fields: fields{Datatype: datatype.Date},
			want:   `"time"`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Column{
				Datatype: tt.fields.Datatype,
			}
			if got := c.RequiredImport(); got != tt.want {
				t.Errorf("RequiredImport() = %v, want %v", got, tt.want)
			}
		})
	}
}
