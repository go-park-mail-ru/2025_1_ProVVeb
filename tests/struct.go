package tests

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/mock"
)

type MockRows struct {
	mock.Mock
	current int
	data    [][]interface{}
	err     error
}

func (m *MockRows) Next() bool {
	m.current++
	return m.current <= len(m.data)
}

func (m *MockRows) Scan(dest ...interface{}) error {
	if m.current == 0 || m.current > len(m.data) {
		return fmt.Errorf("no current row to scan")
	}
	row := m.data[m.current-1]
	for i := range dest {
		switch d := dest[i].(type) {
		case *int:
			*d = row[i].(int)
		case *bool:
			*d = row[i].(bool)
		case *sql.NullTime:
			*d = row[i].(sql.NullTime)
		case *sql.NullInt64:
			fmt.Println(row[i])
			*d = row[i].(sql.NullInt64)
		case *string:
			if nullStr, ok := row[i].(sql.NullString); ok {
				*d = nullStr.String // Берём String из NullString
			} else {
				*d = row[i].(string) // Или обычный string, если передан
			}
		case *sql.NullString:
			*d = row[i].(sql.NullString)
		case *sql.NullBool:
			*d = row[i].(sql.NullBool)
		default:
			return fmt.Errorf("unsupported scan type %T", d)
		}
	}
	return nil
}

func (m *MockRows) Close() {}

func (m *MockRows) Err() error {
	return nil
}

func (m *MockRows) CommandTag() pgconn.CommandTag {
	return pgconn.CommandTag{}
}

func (m *MockRows) Conn() *pgx.Conn {
	return nil
}

func (m *MockRows) RawValues() [][]byte {
	return nil
}

func (m *MockRows) FieldDescriptions() []pgconn.FieldDescription {
	return nil
}

func (m *MockRows) Values() ([]interface{}, error) {
	if m.current == 0 || m.current > len(m.data) {
		return nil, fmt.Errorf("no current row")
	}
	return m.data[m.current-1], nil
}

type MockDB struct {
	mock.Mock
}

func (m *MockDB) Query(ctx context.Context, query string, args ...interface{}) (pgx.Rows, error) {
	argsRet := m.Called(ctx, query, args)
	return argsRet.Get(0).(pgx.Rows), argsRet.Error(1)
}

func (m *MockDB) Begin(ctx context.Context) (pgx.Tx, error) {
	args := m.Called(ctx)
	return args.Get(0).(pgx.Tx), args.Error(1)
}

func (m *MockDB) Exec(ctx context.Context, query string, args ...interface{}) (pgconn.CommandTag, error) {
	argsRet := m.Called(ctx, query, args)
	return argsRet.Get(0).(pgconn.CommandTag), argsRet.Error(1)
}

func (m *MockDB) QueryRow(ctx context.Context, query string, args ...interface{}) pgx.Row {
	argsRet := m.Called(ctx, query, args)
	return argsRet.Get(0).(pgx.Row)
}
