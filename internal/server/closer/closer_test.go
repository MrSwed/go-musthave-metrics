package closer

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCloser_Add(t *testing.T) {
	type args struct {
		n string
		f Func
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "name 1",
			args: args{
				f: func(ctx context.Context) (e error) {
					return
				},
				n: "shutdown 1",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Closer{}
			c.Add(tt.args.n, tt.args.f)
			assert.Len(t, c.names, 1)
			assert.Equal(t, tt.args.n, c.names[0])
			assert.Equal(t, fmt.Sprintf("%p", tt.args.f), fmt.Sprintf("%p", c.funcs[0]))
		})
	}
}

func TestCloser_Close(t *testing.T) {
	type fields struct {
		funcs []Func
		names []string
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "close 1",
			fields: fields{
				funcs: []Func{func(ctx context.Context) error {
					// empty func, do nothing
					return nil
				}},
				names: []string{"shutdown 1"},
			},
			args: args{
				ctx: context.Background(),
			},
			wantErr: false,
		},
		{
			name: "close 2, errr",
			fields: fields{
				funcs: []Func{func(ctx context.Context) error {
					// empty func, do nothing
					return errors.New("errr")
				}},
				names: []string{"shutdown 1"},
			},
			args: args{
				ctx: context.Background(),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Closer{
				funcs: tt.fields.funcs,
				names: tt.fields.names,
			}
			if err := c.Close(tt.args.ctx); (err != nil) != tt.wantErr {
				t.Errorf("Close() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
