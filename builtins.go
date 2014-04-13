package univedo

import (
	"errors"
)

// A Perspective in univedo
type Perspective struct {
	*BasicRemoteObject
}

func newPerspective(id uint64, s sender) RemoteObject {
	return &Perspective{NewBasicRO(id, s)}
}

// Query from a perspective
func (p *Perspective) Query() (*Query, error) {
	ro, err := p.CallROM("query", []interface{}{})
	if err != nil {
		return nil, err
	}
	query, ok := ro.(*Query)
	if !ok {
		return nil, errors.New("got unexpected RO type from query")
	}
	return query, nil
}

// A Query in univedo
type Query struct {
	*BasicRemoteObject
}

func newQuery(id uint64, s sender) RemoteObject {
	return &Query{NewBasicRO(id, s)}
}

// Prepare a statement
func (p *Query) Prepare(queryString string) (*Statement, error) {
	ro, err := p.CallROM("prepare", []interface{}{queryString})
	if err != nil {
		return nil, err
	}
	statement, ok := ro.(*Statement)
	if !ok {
		return nil, errors.New("got unexpected RO type from prepare")
	}
	return statement, nil
}

// A Statement in univedo
type Statement struct {
	*BasicRemoteObject
}

func newStatement(id uint64, s sender) RemoteObject {
	return &Statement{NewBasicRO(id, s)}
}

// Execute a statement
func (p *Statement) Execute() (*Result, error) {
	ro, err := p.CallROM("execute", []interface{}{})
	if err != nil {
		return nil, err
	}
	result, ok := ro.(*Result)
	if !ok {
		return nil, errors.New("got unexpected RO type from execute")
	}
	return result, nil
}

// A Result in univedo
type Result struct {
	*BasicRemoteObject
	Rows chan []interface{}
}

func newResult(id uint64, s sender) RemoteObject {
	r := new(Result)
	r.BasicRemoteObject = NewBasicRO(id, s)
	r.Rows = make(chan []interface{}, 10)
	r.Notifications["setComplete"] = func([]interface{}) {
		close(r.Rows)
	}
	r.Notifications["appendRow"] = func(args []interface{}) {
		// TODO error handling
		if len(args) != 1 {
			panic("appendRow without args")
		}
		row, ok := args[0].([]interface{})
		if !ok {
			panic("appendRow without list")
		}
		r.Rows <- row
	}
	return r
}

func init() {
	RegisteredRemoteObjects["com.univedo.perspective"] = newPerspective
	RegisteredRemoteObjects["com.univedo.query"] = newQuery
	RegisteredRemoteObjects["com.univedo.statement"] = newStatement
	RegisteredRemoteObjects["com.univedo.result"] = newResult
}
