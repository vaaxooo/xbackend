package sqlmock

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"io"
	"regexp"
	"sync"
	"time"
)

// anyArg matches any argument in expectations.
type anyArg struct{}

// NewResult creates a driver.Result with the provided values.
func NewResult(lastInsertID int64, rowsAffected int64) driver.Result {
	return result{lastInsertID: lastInsertID, rowsAffected: rowsAffected}
}

type result struct {
	lastInsertID int64
	rowsAffected int64
}

func (r result) LastInsertId() (int64, error) { return r.lastInsertID, nil }
func (r result) RowsAffected() (int64, error) { return r.rowsAffected, nil }

// Rows represents a mocked set of rows.
type Rows struct {
	columns []string
	data    [][]driver.Value
	pos     int
}

// NewRows constructs a mocked row set.
func NewRows(columns []string) *Rows { return &Rows{columns: columns} }

// AddRow appends a row to the mocked data.
func (r *Rows) AddRow(values ...any) *Rows {
	row := make([]driver.Value, len(values))
	for i, v := range values {
		row[i] = driver.Value(v)
	}
	r.data = append(r.data, row)
	return r
}

func (r *Rows) Columns() []string { return r.columns }
func (r *Rows) Close() error      { return nil }
func (r *Rows) Next(dest []driver.Value) error {
	if r.pos >= len(r.data) {
		return io.EOF
	}
	copy(dest, r.data[r.pos])
	r.pos++
	return nil
}

// Mock holds expectations for a mocked database.
type Mock struct {
	mu            sync.Mutex
	expectations  []expectation
	consumedIndex int
}

// expectation describes a queued expectation.
type expectation interface {
	match(query string, args []driver.Value) (driver.Result, *Rows, error)
	isQuery() bool
}

type execExpectation struct {
	pattern *regexp.Regexp
	args    []any
	result  driver.Result
	err     error
}

type beginExpectation struct{ err error }
type commitExpectation struct{ err error }
type rollbackExpectation struct{ err error }

func (e *execExpectation) match(query string, args []driver.Value) (driver.Result, *Rows, error) {
	if !e.pattern.MatchString(query) {
		return nil, nil, fmt.Errorf("unexpected exec query: %s", query)
	}
	if err := compareArgs(e.args, args); err != nil {
		return nil, nil, err
	}
	return e.result, nil, e.err
}

func (e *execExpectation) isQuery() bool { return false }

func (e *execExpectation) WithArgs(args ...any) *execExpectation {
	e.args = args
	return e
}

func (e *execExpectation) WillReturnResult(res driver.Result) {
	e.result = res
}

func (e *execExpectation) WillReturnError(err error) { e.err = err }

func (e *beginExpectation) match(string, []driver.Value) (driver.Result, *Rows, error) {
	return nil, nil, e.err
}

func (e *beginExpectation) isQuery() bool { return false }

func (e *commitExpectation) match(string, []driver.Value) (driver.Result, *Rows, error) {
	return nil, nil, e.err
}

func (e *commitExpectation) isQuery() bool { return false }

func (e *rollbackExpectation) match(string, []driver.Value) (driver.Result, *Rows, error) {
	return nil, nil, e.err
}

func (e *rollbackExpectation) isQuery() bool { return false }

type queryExpectation struct {
	pattern *regexp.Regexp
	args    []any
	rows    *Rows
	err     error
}

func (q *queryExpectation) match(query string, args []driver.Value) (driver.Result, *Rows, error) {
	if !q.pattern.MatchString(query) {
		return nil, nil, fmt.Errorf("unexpected query: %s", query)
	}
	if err := compareArgs(q.args, args); err != nil {
		return nil, nil, err
	}
	return nil, q.rows, q.err
}

func (q *queryExpectation) isQuery() bool { return true }

func (q *queryExpectation) WithArgs(args ...any) *queryExpectation {
	q.args = args
	return q
}

func (q *queryExpectation) WillReturnRows(rows *Rows) *queryExpectation {
	q.rows = rows
	return q
}

func (q *queryExpectation) WillReturnError(err error) { q.err = err }

// New creates a new mocked sql.DB with attached expectations.
func New() (*sql.DB, *Mock, error) {
	mock := &Mock{}
	driverName := fmt.Sprintf("mock-sql-%d", time.Now().UnixNano())
	sql.Register(driverName, &mockDriver{mock: mock})
	db, err := sql.Open(driverName, "")
	if err != nil {
		return nil, nil, err
	}
	return db, mock, nil
}

func (m *Mock) ExpectExec(pattern string) *execExpectation {
	exp := &execExpectation{pattern: regexp.MustCompile(pattern), result: result{}}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.expectations = append(m.expectations, exp)
	return exp
}

func (m *Mock) ExpectBegin() *beginExpectation {
	exp := &beginExpectation{}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.expectations = append(m.expectations, exp)
	return exp
}

func (m *Mock) ExpectCommit() *commitExpectation {
	exp := &commitExpectation{}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.expectations = append(m.expectations, exp)
	return exp
}

func (m *Mock) ExpectRollback() *rollbackExpectation {
	exp := &rollbackExpectation{}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.expectations = append(m.expectations, exp)
	return exp
}

func (m *Mock) ExpectQuery(pattern string) *queryExpectation {
	exp := &queryExpectation{pattern: regexp.MustCompile(pattern)}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.expectations = append(m.expectations, exp)
	return exp
}

// ExpectationsWereMet ensures all queued expectations were used.
func (m *Mock) ExpectationsWereMet() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.consumedIndex != len(m.expectations) {
		return fmt.Errorf("there are %d unmet expectations", len(m.expectations)-m.consumedIndex)
	}
	return nil
}

func (m *Mock) nextExpectation() expectation {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.consumedIndex >= len(m.expectations) {
		return nil
	}
	exp := m.expectations[m.consumedIndex]
	m.consumedIndex++
	return exp
}

type mockDriver struct {
	mock *Mock
}

func (d *mockDriver) Open(name string) (driver.Conn, error) {
	return &mockConn{mock: d.mock}, nil
}

type mockConn struct {
	mock *Mock
}

func (c *mockConn) Prepare(string) (driver.Stmt, error) {
	return nil, errors.New("prepare not supported")
}
func (c *mockConn) Close() error { return nil }
func (c *mockConn) Begin() (driver.Tx, error) {
	return c.BeginTx(context.Background(), driver.TxOptions{})
}

func (c *mockConn) BeginTx(context.Context, driver.TxOptions) (driver.Tx, error) {
	exp := c.mock.nextExpectation()
	if exp == nil {
		return nil, errors.New("unexpected begin")
	}
	if _, _, err := exp.match("begin", nil); err != nil {
		return nil, err
	}
	return &mockTx{mock: c.mock}, nil
}

func (c *mockConn) ExecContext(_ context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	return c.mock.exec(query, args)
}

func (c *mockConn) QueryContext(_ context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	return c.mock.query(query, args)
}

func (c *mockConn) CheckNamedValue(nv *driver.NamedValue) error { return nil }

type mockTx struct {
	mock *Mock
}

func (t *mockTx) Commit() error {
	exp := t.mock.nextExpectation()
	if exp == nil {
		return errors.New("unexpected commit")
	}
	if _, _, err := exp.match("commit", nil); err != nil {
		return err
	}
	return nil
}

func (t *mockTx) Rollback() error {
	exp := t.mock.nextExpectation()
	if exp == nil {
		return nil
	}
	if _, _, err := exp.match("rollback", nil); err != nil {
		return err
	}
	return nil
}

func (t *mockTx) ExecContext(_ context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	return t.mock.exec(query, args)
}

func (t *mockTx) QueryContext(_ context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	return t.mock.query(query, args)
}

func (m *Mock) exec(query string, args []driver.NamedValue) (driver.Result, error) {
	exp := m.nextExpectation()
	if exp == nil || exp.isQuery() {
		return nil, fmt.Errorf("unexpected exec: %s", query)
	}
	res, _, err := exp.match(query, namedToValues(args))
	return res, err
}

func (m *Mock) query(query string, args []driver.NamedValue) (driver.Rows, error) {
	exp := m.nextExpectation()
	if exp == nil || !exp.isQuery() {
		return nil, fmt.Errorf("unexpected query: %s", query)
	}
	_, rows, err := exp.match(query, namedToValues(args))
	return rows, err
}

func namedToValues(args []driver.NamedValue) []driver.Value {
	values := make([]driver.Value, len(args))
	for i, nv := range args {
		values[i] = nv.Value
	}
	return values
}

func compareArgs(expected []any, actual []driver.Value) error {
	if len(expected) == 0 {
		return nil
	}
	if len(expected) != len(actual) {
		return fmt.Errorf("argument count mismatch: expected %d got %d", len(expected), len(actual))
	}
	for i := range expected {
		if _, ok := expected[i].(anyArg); ok {
			continue
		}
		if !valuesEqual(expected[i], actual[i]) {
			return fmt.Errorf("argument %d mismatch: expected %v got %v", i, expected[i], actual[i])
		}
	}
	return nil
}

func valuesEqual(exp any, act any) bool {
	switch ev := exp.(type) {
	case time.Time:
		if av, ok := act.(time.Time); ok {
			return ev.Equal(av)
		}
	}
	return fmt.Sprint(exp) == fmt.Sprint(act)
}

// AnyArg returns a wildcard argument matcher.
func AnyArg() anyArg { return anyArg{} }
