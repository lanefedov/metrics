package storage

import (
	"errors"
	"fmt"
	"testing"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

func TestIsRetriablePostgresError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "connection exception",
			err:  &pgconn.PgError{Code: pgerrcode.ConnectionException},
			want: true,
		},
		{
			name: "wrapped class 08 connection error",
			err:  fmt.Errorf("query failed: %w", &pgconn.PgError{Code: "08006"}),
			want: true,
		},
		{
			name: "unique violation",
			err:  &pgconn.PgError{Code: pgerrcode.UniqueViolation},
			want: false,
		},
		{
			name: "regular error",
			err:  errors.New("network failed outside postgres"),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isRetriablePostgresError(tt.err)
			if got != tt.want {
				t.Fatalf("isRetriablePostgresError() = %t, want %t", got, tt.want)
			}
		})
	}
}
