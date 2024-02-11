package pgxtras_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/jackc/pgx/v5"
	"github.com/manniwood/pgxtras"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestRowToStructBySnakeToCamelName(t *testing.T) {
	type person struct {
		LastName      string
		FirstName     string
		LikesStarTrek bool
		Age           int32
	}

	defaultConnTestRunner.RunTest(context.Background(), t, func(ctx context.Context, t testing.TB, conn *pgx.Conn) {
		rows, _ := conn.Query(ctx, `
select 'John'  as first_name,
       'Smith' as last_name,
       true    as likes_star_trek,
       n       as age
  from generate_series(0, 9) n`)
		slice, err := pgx.CollectRows(rows, pgxtras.RowToStructBySnakeToCamelName[person])
		assert.NoError(t, err)

		assert.Len(t, slice, 10)
		for i := range slice {
			assert.Equal(t, "Smith", slice[i].LastName)
			assert.Equal(t, "John", slice[i].FirstName)
			assert.True(t, slice[i].LikesStarTrek)
			assert.EqualValues(t, i, slice[i].Age)
		}

		// check missing fields in a returned row
		rows, _ = conn.Query(ctx, `
select 'Smith' as last_name,
       true    as likes_star_trek,
       n       as age
  from generate_series(0, 9) n`)
		_, err = pgx.CollectRows(rows, pgxtras.RowToStructBySnakeToCamelName[person])
		assert.ErrorContains(t, err, "no column in returned row matches struct field FirstName")

		// check missing field in a destination struct
		rows, _ = conn.Query(ctx, `
select 'John'  as first_name,
       'Smith' as last_name,
       true    as likes_star_trek,
       n       as age,
       null    as ignore
  from generate_series(0, 9) n`)
		_, err = pgx.CollectRows(rows, pgxtras.RowToAddrOfStructBySnakeToCamelName[person])
		assert.ErrorContains(t, err, "struct doesn't have corresponding field to match returned column ignore")
	})
}

func TestRowToStructBySnakeToCamelNameEmbeddedStruct(t *testing.T) {
	type Name struct {
		LastName  string
		FirstName string
	}

	type person struct {
		Name
		LikesStarTrek bool
		Age           int32
	}

	defaultConnTestRunner.RunTest(context.Background(), t, func(ctx context.Context, t testing.TB, conn *pgx.Conn) {
		rows, _ := conn.Query(ctx, `
select 'John'  as first_name,
       'Smith' as last_name,
       true    as likes_star_trek,
       n       as age
  from generate_series(0, 9) n`)
		slice, err := pgx.CollectRows(rows, pgxtras.RowToStructBySnakeToCamelName[person])
		assert.NoError(t, err)

		assert.Len(t, slice, 10)
		for i := range slice {
			assert.Equal(t, "Smith", slice[i].Name.LastName)
			assert.Equal(t, "John", slice[i].Name.FirstName)
			assert.True(t, slice[i].LikesStarTrek)
			assert.EqualValues(t, i, slice[i].Age)
		}

		// check missing fields in a returned row
		rows, _ = conn.Query(ctx, `
select 'Smith' as last_name,
       true    as likes_star_trek,
       n       as age
  from generate_series(0, 9) n`)
		_, err = pgx.CollectRows(rows, pgxtras.RowToStructBySnakeToCamelName[person])
		assert.ErrorContains(t, err, "no column in returned row matches struct field FirstName")

		// check missing field in a destination struct
		rows, _ = conn.Query(ctx, `
select 'John'  as first_name,
       'Smith' as last_name,
       true    as likes_star_trek,
       n       as age,
       null    as ignore
  from generate_series(0, 9) n`)
		_, err = pgx.CollectRows(rows, pgxtras.RowToAddrOfStructBySnakeToCamelName[person])
		assert.ErrorContains(t, err, "struct doesn't have corresponding field to match returned column ignore")
	})
}

func TestRowToStructBySimpleName(t *testing.T) {
	type person struct {
		ID            int
		HTTPHandler   string
		LastName      string
		FirstName     string
		LikesStarTrek bool
		Age           int32
	}

	defaultConnTestRunner.RunTest(context.Background(), t, func(ctx context.Context, t testing.TB, conn *pgx.Conn) {
		rows, _ := conn.Query(ctx, `
select 1       as id,
       'foo'   as http_handler,
       'John'  as first_name,
       'Smith' as last_name,
       true    as likes_star_trek,
       n       as age
  from generate_series(0, 9) n`)
		slice, err := pgx.CollectRows(rows, pgxtras.RowToStructBySimpleName[person])
		assert.NoError(t, err)

		assert.Len(t, slice, 10)
		for i := range slice {
			assert.Equal(t, 1, slice[i].ID)
			assert.Equal(t, "foo", slice[i].HTTPHandler)
			assert.Equal(t, "Smith", slice[i].LastName)
			assert.Equal(t, "John", slice[i].FirstName)
			assert.True(t, slice[i].LikesStarTrek)
			assert.EqualValues(t, i, slice[i].Age)
		}

		// check missing fields in a returned row
		rows, _ = conn.Query(ctx, `
select 1       as id,
       'foo'   as http_handler,
      'Smith' as last_name,
       true    as likes_star_trek,
       n       as age
  from generate_series(0, 9) n`)
		_, err = pgx.CollectRows(rows, pgxtras.RowToStructBySimpleName[person])
		assert.ErrorContains(t, err, "no column in returned row matches struct field FirstName")

		// check missing field in a destination struct
		rows, _ = conn.Query(ctx, `
select 1       as id,
       'foo'   as http_handler,
       'John'  as first_name,
       'Smith' as last_name,
       true    as likes_star_trek,
       n       as age,
       null    as ignore
  from generate_series(0, 9) n`)
		_, err = pgx.CollectRows(rows, pgxtras.RowToAddrOfStructBySimpleName[person])
		assert.ErrorContains(t, err, "struct doesn't have corresponding field to match returned column ignore")
	})
}

func TestRowToStructBySimpleNameEmbeddedStruct(t *testing.T) {
	type Name struct {
		LastName  string
		FirstName string
	}

	type person struct {
		Name
		LikesStarTrek bool
		Age           int32
		ID            int
		HTTPHandler   string
	}

	defaultConnTestRunner.RunTest(context.Background(), t, func(ctx context.Context, t testing.TB, conn *pgx.Conn) {
		rows, _ := conn.Query(ctx, `
select 1       as id,
       'foo'   as http_handler,
       'John'  as first_name,
       'Smith' as last_name,
       true    as likes_star_trek,
       n       as age
  from generate_series(0, 9) n`)
		slice, err := pgx.CollectRows(rows, pgxtras.RowToStructBySimpleName[person])
		assert.NoError(t, err)

		assert.Len(t, slice, 10)
		for i := range slice {
			assert.Equal(t, 1, slice[i].ID)
			assert.Equal(t, "foo", slice[i].HTTPHandler)
			assert.Equal(t, "Smith", slice[i].Name.LastName)
			assert.Equal(t, "John", slice[i].Name.FirstName)
			assert.True(t, slice[i].LikesStarTrek)
			assert.EqualValues(t, i, slice[i].Age)
		}

		// check missing fields in a returned row
		rows, _ = conn.Query(ctx, `
select 1       as id,
       'foo'   as http_handler,
       'Smith' as last_name,
       true    as likes_star_trek,
       n       as age
  from generate_series(0, 9) n`)
		_, err = pgx.CollectRows(rows, pgxtras.RowToStructBySimpleName[person])
		assert.ErrorContains(t, err, "no column in returned row matches struct field FirstName")

		// check missing field in a destination struct
		rows, _ = conn.Query(ctx, `
select 1       as id,
       'foo'   as http_handler,
       'John'  as first_name,
       'Smith' as last_name,
       true    as likes_star_trek,
       n       as age,
       null    as ignore
  from generate_series(0, 9) n`)
		_, err = pgx.CollectRows(rows, pgxtras.RowToAddrOfStructBySimpleName[person])
		assert.ErrorContains(t, err, "struct doesn't have corresponding field to match returned column ignore")
	})
}

func TestRowToMapStrStr(t *testing.T) {
	defaultConnTestRunner.RunTest(context.Background(), t, func(ctx context.Context, t testing.TB, conn *pgx.Conn) {
		rows, _ := conn.Query(ctx, `select 'Joe' as name, n::text as age from generate_series(0, 9) n`)
		slice, err := pgx.CollectRows(rows, pgxtras.RowToMapStrStr)
		require.NoError(t, err)

		assert.Len(t, slice, 10)
		for i := range slice {
			assert.Equal(t, "Joe", slice[i]["name"])
			assert.EqualValues(t, fmt.Sprintf("%d", i), slice[i]["age"])
		}
	})
}
