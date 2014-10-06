package univedo

// Session on univedo
type Session struct {
	*BasicRemoteObject
}

// Ping the server
func (s *Session) Ping(val interface{}) (interface{}, error) {
	return s.CallROM("ping", val)
}

// ApplyUTS applies a uts
func (s *Session) ApplyUTS(uts string) error {
	_, err := s.CallROM("applyUts", uts)
	return err
}

func newSession(id uint64, send sender) RemoteObject {
	return &Session{NewBasicRO(id, send)}
}

func init() {
	RegisterRemoteObject("com.univedo.session", newSession)
}
