// Package repository provide abstract implementation for data storage of model/model.go struct
package dbrepo

import (
	"context"
	"database/sql"
	"reflect"
	"testing"

	"github.com/Pklerik/urlshortener/internal/model"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func TestDBLinksRepository_getShort(t *testing.T) {
	type fields struct {
		db *sql.DB
	}
	type args struct {
		ctx   context.Context
		tx    *sql.Tx
		short string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.LinkData
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &DBLinksRepository{
				db: tt.fields.db,
			}
			got, err := r.getShort(tt.args.ctx, tt.args.tx, tt.args.short)
			if (err != nil) != tt.wantErr {
				t.Errorf("DBLinksRepository.getShort() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DBLinksRepository.getShort() = %v, want %v", got, tt.want)
			}
		})
	}
}
