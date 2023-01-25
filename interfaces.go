package pgxtras

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// Querier is an interface that wraps the Query method used by
// pgx.Conn, pgxpool.Pool, pgx.Tx, and pgxpool.Tx.
//
// Feel free to copy this interface directly into your code
// instead of importing this module just to use this interface:
// The licence is quite permissive.
//
// Write your database querying functions and methods with this interface
// as one of the args. By doing so, your function
// or method can run a query using a pgx.Conn, or a pgxpool.Pool
// , or a pgx.Tx, or a pgxpool.Tx.
//
// This makes your databae query function
// work in the following scenarios with no code changes:
// 1) a command-line tool (a single database connection pgx.Conn)
// 2) a web server (a pool pgxpool.Pool)
// 3) as part of a larger transaction (pgx.Tx or pgxpool.Tx),
type Querier interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
}

// Execer is an interface that wraps the Exec method used by
// pgx.Conn, pgxpool.Pool, pgx.Tx, and pgxpool.Tx.
//
// Feel free to copy this interface directly into your code
// instead of importing this module just to use this interface:
// The licence is quite permissive.
//
// Write your database querying functions and methods with this interface
// as one of the args. By doing so, your function
// or method can run a query using a pgx.Conn, or a pgxpool.Pool
// , or a pgx.Tx, or a pgxpool.Tx.
//
// This makes your databae query function
// work in the following scenarios with no code changes:
// 1) a command-line tool (a single database connection pgx.Conn)
// 2) a web server (a pool pgxpool.Pool)
// 3) as part of a larger transaction (pgx.Tx or pgxpool.Tx),
type Execer interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

// QuerierExecer combines the Querier and Execer interfaces for situations
// where you write a function or method that performs both select statements
// and inserts/updates/DDL, etc.
//
// Feel free to copy this interface directly into your code
// instead of importing this module just to use this interface:
// The licence is quite permissive.
type QuerierExecer interface {
	Querier
	Execer
}
