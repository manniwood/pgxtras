package pgxtras_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/jackc/pgx/v5"
	"github.com/manniwood/pgxtras"
	"github.com/stretchr/testify/assert"
)

func TestCollectOneRowOK(t *testing.T) {
	defaultConnTestRunner.RunTest(context.Background(), t, func(ctx context.Context, t testing.TB, conn *pgx.Conn) {
		rows, _ := conn.Query(ctx, `select 42`)
		n, ok, err := pgxtras.CollectOneRowOK(rows, func(row pgx.CollectableRow) (int32, error) {
			var n int32
			err := row.Scan(&n)
			return n, err
		})
		assert.NoError(t, err)
		assert.Equal(t, int32(42), n)
		assert.True(t, ok)
	})
}

func TestCollectOneRowOKNotFound(t *testing.T) {
	defaultConnTestRunner.RunTest(context.Background(), t, func(ctx context.Context, t testing.TB, conn *pgx.Conn) {
		rows, _ := conn.Query(ctx, `select 42 where false`)
		n, ok, err := pgxtras.CollectOneRowOK(rows, func(row pgx.CollectableRow) (int32, error) {
			var n int32
			err := row.Scan(&n)
			return n, err
		})
		assert.NoError(t, err)
		assert.Equal(t, int32(0), n)
		assert.False(t, ok)
	})
}

// TODO: TestRowToStructBySnakeToCamelName
// TODO: TestRowToAddrOfStructBySnakeToCamelName

func TestSnakeToCamel(t *testing.T) {
	tests := map[string]struct {
		input string
		want  string
	}{
		"one segment":         {input: "col", want: "Col"},
		"typical":             {input: "a_col_name", want: "AColName"},
		"empty":               {input: "", want: ""},
		"leading underscore":  {input: "_a_col", want: "ACol"},
		"trailing underscore": {input: "a_col_", want: "ACol"},
	}
	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			got := pgxtras.SnakeToCamel(testCase.input)
			diff := cmp.Diff(testCase.want, got)
			if diff != "" {
				t.Fatalf(diff)
			}
		})
	}
}
