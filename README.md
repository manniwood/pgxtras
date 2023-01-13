# pgxtras - Extra functionality that compliments pgx

A few extra functions to extend functionality in
the excellent github.com/jackc/pgx/v5 library.

## `CollectOneRowOK()`

In a `psql` session, whether one row is expected or many rows are expected,
getting back 0 rows is not a SQL error.

It has always bothered me that `myPgxConnection.QueryRow()` returns
an error when no row is found (`pgx.ErrNoRows`) instead of an empty
result. After all, `rows, err := myPgxConnection.Query()` returns no rows
(strictly speaking, the first call to `rows.Next()` is false)

The same is true of the `pgx.CollectOneRow()` convenience method:
when `pgx.CollectOneRow()` finds no rows, it returns an error (`ErrNoRows`).
But when `pgx.CollectRows()` finds no rows, it does not return an error;
it's just that the returned slice is length 0.

`pgxtras.CollectOneRowOK()` provides a way to also return without error when no
rows are found. Whereas a caller of `pgx.CollectRows()` can check if the
returned slice is length > 0 to determine if rows were found, a caller of
`pgxtras.CollectOneRowOK()` can check the second return value (usually named `ok`)
to see if a row was found.

This follows the "comma-OK" idiom found in other Go libraries. For instance,
in the `os` package of the standard library, `os.GetEnv()` returns an
empty string if an env var is not found; but `os.LookupEnv()` returns
a second boolean value (by convention named `ok`) that is set to true
if the env var was present but set to the empty string, or false if
the env var truly was not present. Also, when getting a value from a map,
the "comma-OK" idiom can be used for the same purpose.

## `pgxtras.RowToStructBySnakeToCamelName()` and `pgxtras.RowToAddrOfStructBySnakeToCamelName()`

The pgx library has convenience functions `pgx.RowToStructByName()` and
`pgx.RowToAddrOfStructByName()`, which ignore case when assigning
column results to struct fields.

The pgextras convenience functions, `pgxtras.RowToStructBySnakeToCamelName()`
and `pgxtras.RowToAddrOfStructBySnakeToCamelName()`, cover the common case
of database columns being named in `snake_case` and Go struct fields
being named in `CamelCase`. These are both rather common naming conventions,
and these two convenince methods translate between them to relieve the user
of having to use any special tags in Go structs, or any `as` column
aliasing in SQL.
