package repo

import (
	"database/sql"
	"reflect"
)

func getTable[T any](rows *sql.Rows) ([]T, error) {
	var table []T
	for rows.Next() {
		var data T
		s := reflect.ValueOf(&data).Elem()
		numCols := s.NumField()
		columns := make([]any, numCols)

		for i := range numCols {
			field := s.Field(i)
			columns[i] = field.Addr().Interface()
		}

		if err := rows.Scan(columns...); err != nil {
			return nil, err
		}

		table = append(table, data)
	}
	return table, nil
}
