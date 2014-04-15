package univedo

import (
	"errors"
)

const (
	romCall   uint64 = 1
	romAnswer        = 2
	romNotify        = 3
	romDelete        = 4
)

// A remote method result can either be an error or a message
type romResult struct {
	err   error
	value interface{}
}

type sender interface {
	sendMessage(...interface{}) error
}

// RemoteObject provides methods for calling remote methods
type RemoteObject interface {
	CallROM(string, ...interface{}) (interface{}, error)
	SendNotification(string, ...interface{}) error
	receive(msg []interface{}) error
}

// BasicRemoteObject can be used as a simple remote object without convenience wrappers
type BasicRemoteObject struct {
	id            uint64
	session       sender
	callID        uint64
	callResults   map[uint64]chan romResult
	Notifications map[string]func([]interface{})
}

// NewBasicRO creates a new remote object
func NewBasicRO(id uint64, session sender) *BasicRemoteObject {
	m := make(map[uint64]chan romResult)
	n := make(map[string]func([]interface{}))
	return &BasicRemoteObject{id: id, session: session, callResults: m, Notifications: n}
}

// CallROM calls a method on the remote object and returns its result
func (ro *BasicRemoteObject) CallROM(name string, args ...interface{}) (interface{}, error) {
	c := make(chan romResult)
	ro.callResults[ro.callID] = c
	defer delete(ro.callResults, ro.callID)

	err := ro.session.sendMessage(ro.id, uint64(romCall), ro.callID, name, args)
	if err != nil {
		return nil, err
	}

	ro.callID++

	result := <-c
	if result.err != nil {
		return nil, result.err
	}
	return result.value, nil
}

// SendNotification sends a notification to the remote object
func (ro *BasicRemoteObject) SendNotification(name string, args ...interface{}) error {
	return ro.session.sendMessage(ro.id, uint64(romNotify), name, args)
}

func shiftSlice(s []interface{}) (interface{}, []interface{}) {
	if len(s) == 0 {
		return nil, nil
	}
	return s[0], s[1:len(s)]
}

func (ro *BasicRemoteObject) receive(msg []interface{}) error {
	iOpcode, msg := shiftSlice(msg)
	if msg == nil {
		return errors.New("unexpected end of message")
	}
	opcode, ok := iOpcode.(uint64)
	if !ok {
		return errors.New("opcode must be an integer")
	}

	switch opcode {
	case romAnswer:
		iCallID, msg := shiftSlice(msg)
		if msg == nil {
			return errors.New("unexpected end of message")
		}
		callID, ok := iCallID.(uint64)
		if !ok {
			return errors.New("call id must be an uint")
		}

		c := ro.callResults[callID]
		defer close(c)

		iStatus, msg := shiftSlice(msg)
		if msg == nil {
			return errors.New("unexpected end of message")
		}
		status, ok := iStatus.(uint64)
		if !ok {
			return errors.New("status must be an uint")
		}

		switch status {
		case 0:
			result, msg := shiftSlice(msg)
			if msg == nil {
				return errors.New("unexpected end of message")
			}

			if c == nil {
				return errors.New("received answer to nonexistant call")
			}
			c <- romResult{value: result}

		case 2:
			err, msg := shiftSlice(msg)
			if msg == nil {
				return errors.New("unexpected end of message")
			}
			errString, ok := err.(string)
			if !ok {
				return errors.New("error must be a string")
			}
			c <- romResult{err: errors.New(errString)}

		default:
			return errors.New("unknown status in remote object")
		}

		return nil

	case romNotify:
		name, msg := shiftSlice(msg)
		if name == nil {
			return errors.New("unexpected end of message")
		}
		nameString, ok := name.(string)
		if !ok {
			return errors.New("notification name must be a string")
		}
		args, msg := shiftSlice(msg)
		if args == nil {
			return errors.New("unexpected end of message")
		}
		argsSlice, ok := args.([]interface{})
		if !ok {
			return errors.New("notification args must be a string")
		}

		f := ro.Notifications[nameString]
		if f == nil {
			return errors.New("unknown notification " + nameString)
		}

		go f(argsSlice)
		return nil
	default:
		return errors.New("unknown opcode in remote object")
	}
}
