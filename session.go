package univedo

import (
	"bytes"
	"errors"

	"code.google.com/p/go.net/websocket"
	// TODO remove
	_ "fmt"
	"net/url"
)

// registeredRemoteObjects is a map from RO name to factory function
var registeredRemoteObjects = make(map[string]func(id uint64, s sender) RemoteObject)

// RegisterRemoteObject adds a remote object factory for a RO name
func RegisterRemoteObject(name string, factory func(id uint64, session sender) RemoteObject) {
	registeredRemoteObjects[name] = factory
}

// A Session with an univedo server
type Session struct {
	RemoteObject
	ws            *websocket.Conn
	urologin      RemoteObject
	remoteObjects map[uint64]RemoteObject
}

// originForURL returns an origin matching the given URL
func originForURL(s string) (string, error) {
	u, err := url.Parse(s)
	if err != nil {
		return "", err
	}
	origin := &url.URL{Scheme: "http", Host: u.Host}
	return origin.String(), nil
}

// Dial opens a new connection with an univedo server
func Dial(url string) (*Session, error) {
	// Extract the origin from the URL
	origin, err := originForURL(url)
	if err != nil {
		return nil, err
	}

	// Dial the websocket
	ws, err := websocket.Dial(url, "", origin)
	if err != nil {
		return nil, err
	}

	s := &Session{ws: ws, remoteObjects: make(map[uint64]RemoteObject)}
	go func() {
		// TODO error handling
		err := s.handleWebsocket()
		/*		fmt.Printf("%s\n", err.Error())*/
		_ = err
	}()

	// Login
	s.urologin = NewBasicRO(0, s)
	s.remoteObjects[0] = s.urologin

	creds := map[string]interface{}{"username": "marvin"}
	iSession, err := s.urologin.CallROM("getSession", creds)
	if err != nil {
		ws.Close()
		return nil, err
	}
	session, ok := iSession.(RemoteObject)
	if !ok {
		ws.Close()
		return nil, errors.New("getSession did not return a remote object")
	}

	s.RemoteObject = session

	return s, nil
}

// Close the connection
func (s *Session) Close() {
	s.ws.Close()
}

// Ping the server
func (s *Session) Ping(v interface{}) (interface{}, error) {
	return s.CallROM("ping", v)
}

// ApplyUTS uses a uts file
func (s *Session) ApplyUTS(utsString string) (interface{}, error) {
	return s.CallROM("applyUts", utsString)
}

func (s *Session) sendMessage(data ...interface{}) error {
	m := &message{buffer: &bytes.Buffer{}}
	for _, v := range data {
		m.send(v)
	}
	return websocket.Message.Send(s.ws, m.buffer.Bytes())
}

func (s *Session) handleWebsocket() error {
	for {
		var buffer []byte
		err := websocket.Message.Receive(s.ws, &buffer)

		if err != nil {
			return err
		}

		msg := &message{buffer: bytes.NewBuffer(buffer), createRO: s.receiveRO}

		iRoID, err := msg.read()
		if err != nil {
			return err
		}
		roID, ok := iRoID.(uint64)
		if !ok {
			return errors.New("ro id should be int")
		}

		ro := s.remoteObjects[roID]
		if ro == nil {
			return errors.New("ro not known")
		}

		var data []interface{}
		for !msg.empty() {
			v, err := msg.read()
			if err != nil {
				return err
			}
			data = append(data, v)
		}

		err = ro.receive(data)
		if err != nil {
			return err
		}
	}
}

func (s *Session) receiveRO(id uint64, name string) interface{} {
	var ro RemoteObject
	factory := registeredRemoteObjects[name]
	if factory != nil {
		ro = factory(id, s)
	} else {
		ro = NewBasicRO(id, s)
	}
	s.remoteObjects[id] = ro
	return ro
}
