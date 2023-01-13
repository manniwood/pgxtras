package pgxtras

import (
	"fmt"
	"reflect"
	"strings"
	"unicode"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// CollectOneRowOK is CollectOneRow with the "comma OK idiom": think 'val, ok := myMap["foo"]'
// or os.GetEnv (no comma OK idiom) versus os.LookupEnv (follows comma OK idiom).
// If no rows are found, the second return value is false (instead of returning error ErrNoRows
// as CollectOneRow does). If a row is found, the second return value is true.
func CollectOneRowOK[T any](rows pgx.Rows, fn pgx.RowToFunc[T]) (T, bool, error) {
	var value T
	var err error
	value, err = pgx.CollectOneRow(rows, fn)
	if err != nil {
		if err == pgx.ErrNoRows {
			return value, false, nil
		}
		return value, false, err
	}
	return value, true, nil
}

// RowToStructBySnakeToCamelName returns a T scanned from row. T must be a struct. T must have the same number of named public
// fields as row has fields. The row and T fields will by matched by name, converting snake case column names to camel case struct field names.
func RowToStructBySnakeToCamelName[T any](row pgx.CollectableRow) (T, error) {
	var value T
	err := row.Scan(&namedCamelStructRowScanner{ptrToStruct: &value})
	return value, err
}

// RowToAddrOfStructBySnakeToCamelName returns the address of a T scanned from row. T must be a struct. T must have the same number of named public
// fields as row has fields. The row and T fields will by matched by name, converting snake case column names to camel case struct field names.
func RowToAddrOfStructBySnakeToCamelName[T any](row pgx.CollectableRow) (*T, error) {
	var value T
	err := row.Scan(&namedCamelStructRowScanner{ptrToStruct: &value})
	return &value, err
}

type namedCamelStructRowScanner struct {
	ptrToStruct any
}

func (rs *namedCamelStructRowScanner) ScanRow(rows pgx.Rows) error {
	dst := rs.ptrToStruct
	dstValue := reflect.ValueOf(dst)
	if dstValue.Kind() != reflect.Ptr {
		return fmt.Errorf("dst not a pointer")
	}

	dstElemValue := dstValue.Elem()
	scanTargets, err := rs.appendScanTargets(dstElemValue, nil, rows.FieldDescriptions())

	if err != nil {
		return err
	}

	for i, t := range scanTargets {
		if t == nil {
			return fmt.Errorf("struct doesn't have corresponding field to match returned column %s", rows.FieldDescriptions()[i].Name)
		}
	}

	return rows.Scan(scanTargets...)
}

const structTagKey = "db"

func fieldPosByCamelName(fldDescs []pgconn.FieldDescription, field string) (i int) {
	i = -1
	for i, desc := range fldDescs {
		if SnakeToCamel(desc.Name) == field {
			return i
		}
	}
	return
}

// SnakeToCamel takes a_typical_db_col in snake case and translates it to
// TheCamelCase found in public fields of Go structs.
func SnakeToCamel(s string) string {
	snakeParts := strings.Split(s, "_")
	var sb strings.Builder
	for _, snakePart := range snakeParts {
		theRunes := []rune(snakePart)
		for i, rn := range theRunes {
			if i == 0 {
				rn = unicode.ToUpper(rn)
			}
			sb.WriteRune(rn)
		}
	}
	return sb.String()
}

func (rs *namedCamelStructRowScanner) appendScanTargets(dstElemValue reflect.Value, scanTargets []any, fldDescs []pgconn.FieldDescription) ([]any, error) {
	var err error
	dstElemType := dstElemValue.Type()

	if scanTargets == nil {
		scanTargets = make([]any, len(fldDescs))
	}

	for i := 0; i < dstElemType.NumField(); i++ {
		sf := dstElemType.Field(i)
		if sf.PkgPath != "" && !sf.Anonymous {
			// Field is unexported, skip it.
			continue
		}
		// Handle anoymous struct embedding, but do not try to handle embedded pointers.
		if sf.Anonymous && sf.Type.Kind() == reflect.Struct {
			scanTargets, err = rs.appendScanTargets(dstElemValue.Field(i), scanTargets, fldDescs)
			if err != nil {
				return nil, err
			}
		} else {
			dbTag, dbTagPresent := sf.Tag.Lookup(structTagKey)
			if dbTagPresent {
				dbTag = strings.Split(dbTag, ",")[0]
			}
			if dbTag == "-" {
				// Field is ignored, skip it.
				continue
			}
			colName := dbTag
			if !dbTagPresent {
				colName = sf.Name
			}
			fpos := fieldPosByCamelName(fldDescs, colName)
			if fpos == -1 || fpos >= len(scanTargets) {
				return nil, fmt.Errorf("no column in returned row matches struct field %s", colName)
			}
			scanTargets[fpos] = dstElemValue.Field(i).Addr().Interface()
		}
	}

	return scanTargets, err
}
