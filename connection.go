package univedo

import (
	"bytes"
	"errors"
	"strings"

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

// A Connection with an univedo server
type Connection struct {
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
func Dial(url string) (*Connection, error) {
	// Extract the origin from the URL
	origin, err := originForURL(url)
	if err != nil {
		return nil, err
	}

	// Dial the websocket
	if !strings.HasSuffix(url, "/") {
		url += "/"
	}

	url += "v1"
	ws, err := websocket.Dial(url, "", origin)
	if err != nil {
		return nil, err
	}

	c := &Connection{ws: ws, remoteObjects: make(map[uint64]RemoteObject)}
	go func() {
		// TODO error handling
		err := c.handleWebsocket()
		/*		fmt.Printf("%s\n", err.Error())*/
		_ = err
	}()

	// Login
	c.urologin = NewBasicRO(0, c)
	c.remoteObjects[0] = c.urologin

	return c, nil
}

// Close the connection
func (c *Connection) Close() {
	c.ws.Close()
}

// GetSession connects to a bucket with credentials
func (c *Connection) GetSession(bucket string, creds map[string]interface{}) (*Session, error) {
	iSession, err := c.urologin.CallROM("getSession", bucket, creds)
	if err != nil {
		return nil, err
	}
	session, ok := iSession.(*Session)
	if !ok {
		return nil, errors.New("getSession did not return a remote object")
	}
	return session, nil
}

func (c *Connection) sendMessage(data ...interface{}) error {
	m := &message{buffer: &bytes.Buffer{}}
	for _, v := range data {
		m.send(v)
	}
	return websocket.Message.Send(c.ws, m.buffer.Bytes())
}

func (c *Connection) handleWebsocket() error {
	for {
		var buffer []byte
		err := websocket.Message.Receive(c.ws, &buffer)

		if err != nil {
			return err
		}

		msg := &message{buffer: bytes.NewBuffer(buffer), createRO: c.receiveRO}

		iRoID, err := msg.read()
		if err != nil {
			return err
		}
		roID, ok := iRoID.(uint64)
		if !ok {
			return errors.New("ro id should be int")
		}

		ro := c.remoteObjects[roID]
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

func (c *Connection) receiveRO(id uint64, name string) interface{} {
	var ro RemoteObject
	factory := registeredRemoteObjects[name]
	if factory != nil {
		ro = factory(id, c)
	} else {
		ro = NewBasicRO(id, c)
	}
	c.remoteObjects[id] = ro
	return ro
}
