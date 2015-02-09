package univedo

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"io"
	"net/url"
	"strconv"
	"strings"
)

// UnivedoDriver implements the interface required by database/sql
type UnivedoDriver struct{}

// Open a new connection
// You should probably use database/sql instead of this directly.
func (UnivedoDriver) Open(name string) (driver.Conn, error) {
	// Extract the perspective uuid from url
	u, err := url.Parse(name)
	if err != nil {
		return nil, err
	}
	pathComponents := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(pathComponents) != 2 {
		return nil, errors.New("requires bucket and perspective name")
	}
	bucketName := pathComponents[0]
	perspectiveName := pathComponents[1]

	u.Path = "/"

	// Extract credentials from url
	creds := map[string]interface{}{}
	for k, v := range u.Query() {
		creds[k] = v[0]
	}
	u.RawQuery = ""

	// Open connection and perspective
	connection, err := Dial(u.String())
	if err != nil {
		return nil, err
	}

	session, err := connection.GetSession(bucketName, creds)
	if err != nil {
		return nil, err
	}

	perspective, err := getRoFromROM(session, "getPerspective", perspectiveName)
	if err != nil {
		return nil, err
	}

	return &Conn{Connection: connection, perspective: perspective}, nil
}

// Conn implements a connection as required by database/sql
type Conn struct {
	Connection  *Connection
	perspective RemoteObject
}

// Begin a transaction as required by database/sql
func (conn *Conn) Begin() (driver.Tx, error) {
	return nil, errors.New("NI Begin()")
}

// Close the connection as required by database/sql
func (conn *Conn) Close() error {
	// TODO error handling
	conn.Connection.Close()
	return nil
}

// Prepare a statement as required by database/sql
func (conn *Conn) Prepare(query string) (driver.Stmt, error) {
	queryRO, err := getRoFromROM(conn.perspective, "query")
	if err != nil {
		return nil, err
	}
	stmtRO, err := getRoFromROM(queryRO, "prepare", query)
	if err != nil {
		return nil, err
	}
	s, ok := stmtRO.(*stmt)
	if !ok {
		return nil, errors.New("expected com.univedo.statement from prepare")
	}

	return s, nil
}

// A statement in univedo
// Implements the Stmt interface from sql/driver
type stmt struct {
	*BasicRemoteObject
	columnNames chan []string
}

func newStatement(id uint64, send sender) RemoteObject {
	s := new(stmt)
	s.BasicRemoteObject = NewBasicRO(id, send)

	s.columnNames = make(chan []string)

	s.Notifications["setColumnNames"] = func(args []interface{}) {
		// TODO error handling
		if len(args) != 1 {
			panic("setColumnNames without args")
		}
		colNamesI, ok := args[0].([]interface{})
		if !ok {
			panic("setColumnNames without []interface")
		}
		colNames := make([]string, len(colNamesI))
		for i, v := range colNamesI {
			str, ok := v.(string)
			if !ok {
				panic("setColumnNames without strings")
			}
			colNames[i] = str
		}
		s.columnNames <- colNames
		close(s.columnNames)
	}

	s.Notifications["setColumnTypes"] = func(args []interface{}) {}

	return s
}

func (s *stmt) Close() error {
	return errors.New("NI Close()")
}

func (s *stmt) Exec(binds []driver.Value) (driver.Result, error) {
	return execStatement(s, binds)
}

func (s *stmt) Query(binds []driver.Value) (driver.Rows, error) {
	cols := <-s.columnNames

	result, err := execStatement(s, binds)
	if err != nil {
		return nil, err
	}
	result.cols = cols
	return result, nil
}

func (s *stmt) NumInput() int {
	return -1
}

// A result in univedo
// Result implements both the Result and Rows interface for sqldatabase/sql
type result struct {
	*BasicRemoteObject
	cols           []string
	rows           chan []interface{}
	errors         chan error
	lastInsertedID chan uint64
	rowsAffected   chan uint64
}

func newResult(id uint64, s sender) RemoteObject {
	r := new(result)
	r.BasicRemoteObject = NewBasicRO(id, s)

	r.rows = make(chan []interface{}, 100)
	r.lastInsertedID = make(chan uint64, 1)
	r.rowsAffected = make(chan uint64, 1)
	r.errors = make(chan error, 1)

	r.Notifications["setError"] = func(args []interface{}) {
		if len(args) != 1 {
			panic("setError without args")
		}
		err, ok := args[0].(string)
		if !ok {
			panic("setError without error string")
		}
		r.errors <- errors.New(err)
	}

	r.Notifications["setComplete"] = func([]interface{}) {
		close(r.rows)
	}

	r.Notifications["setTuple"] = func(args []interface{}) {
		// TODO error handling
		if len(args) != 1 {
			panic("setTuple without args")
		}
		row, ok := args[0].([]interface{})
		if !ok {
			panic("setTuple without list")
		}
		r.rows <- row
	}

	r.Notifications["setId"] = func(args []interface{}) {
		// TODO remove
		defer func() {
			_ = recover()
		}()
		// TODO error handling
		if len(args) != 1 {
			panic("setId without args")
		}
		id, ok := args[0].(uint64)
		if !ok {
			panic("setId without uint64")
		}
		r.lastInsertedID <- id
		r.rowsAffected <- 1
		close(r.lastInsertedID)
	}

	r.Notifications["setNAffectedRecords"] = func(args []interface{}) {
		// TODO error handling
		if len(args) != 1 {
			panic("setNAffectedRecords without args")
		}
		num, ok := args[0].(uint64)
		if !ok {
			panic("setNAffectedRecords without uint64")
		}
		r.rowsAffected <- num
		close(r.rowsAffected)
	}

	return r
}

func (r *result) Close() error {
	// TODO implement
	return nil
}

func (r *result) Columns() []string {
	return r.cols
}

func (r *result) Next(dest []driver.Value) error {
	select {
	case err := <-r.errors:
		return err
	case row, ok := <-r.rows:
		if !ok {
			return io.EOF
		}
		for i := range dest {
			// TODO conversions
			dest[i] = row[i]
		}
		return nil
		// TODO handle invalid cases
	}
}

func (r *result) LastInsertId() (int64, error) {
	select {
	case err := <-r.errors:
		return 0, err
	case id, ok := <-r.lastInsertedID:
		if !ok {
			return 0, errors.New("lastInsertId called twice")
		}
		return int64(id), nil
		// TODO handle invalid cases
	}
}

func (r *result) RowsAffected() (int64, error) {
	select {
	case err := <-r.errors:
		return 0, err
	case num, ok := <-r.rowsAffected:
		if !ok {
			return 0, errors.New("rowsAffected called twice")
		}
		return int64(num), nil
		// TODO handle invalid cases
	}
}

func execStatement(stmt *stmt, binds []driver.Value) (*result, error) {
	bindsI := make(map[string]interface{})
	for i, v := range binds {
		bindsI[strconv.Itoa(i)] = v
	}
	r, err := getRoFromROM(stmt, "execute", bindsI)
	if err != nil {
		return nil, err
	}
	result, ok := r.(*result)
	if !ok {
		return nil, errors.New("expected com.univedo.result")
	}
	return result, nil
}

func getRoFromROM(ro RemoteObject, rom string, args ...interface{}) (RemoteObject, error) {
	roI, err := ro.CallROM(rom, args...)
	if err != nil {
		return nil, err
	}
	ro, ok := roI.(RemoteObject)
	if !ok {
		return nil, errors.New("expected RO as return value")
	}
	return ro, nil
}

func init() {
	sql.Register("univedo", &UnivedoDriver{})
	RegisterRemoteObject("com.univedo.result", newResult)
	RegisterRemoteObject("com.univedo.statement", newStatement)
}
