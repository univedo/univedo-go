package univedo

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

type testSession struct {
	msg       []interface{}
	onMessage func()
}

func (s *testSession) sendMessage(msg []interface{}) error {
	s.msg = msg
	go s.onMessage()
	return nil
}

func TestRemoteObject(t *testing.T) {
	Convey("remote objects", t, func() {
		Convey("does rom calls", func() {
			s := new(testSession)
			ro := createBasicRO(23, s)
			s.onMessage = func() {
				So(s.msg, ShouldResemble, []interface{}{23, 1, 0, "foo", []interface{}{1, "2", 3}})
				So(ro.receive([]interface{}{2, 0, 0, 42}), ShouldBeNil)
			}
			res, err := ro.CallROM("foo", []interface{}{1, "2", 3})
			So(err, ShouldBeNil)
			So(res, ShouldEqual, 42)
		})

		Convey("errors from rom calls", func() {
			s := new(testSession)
			ro := createBasicRO(23, s)
			s.onMessage = func() {
				So(s.msg, ShouldResemble, []interface{}{23, 1, 0, "foo", []interface{}{1, "2", 3}})
				So(ro.receive([]interface{}{2, 0, 2, "boom"}), ShouldBeNil)
			}
			res, err := ro.CallROM("foo", []interface{}{1, "2", 3})
			So(err, ShouldNotBeNil)
			So(res, ShouldBeNil)
		})
	})
}
