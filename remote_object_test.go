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
		Convey("send notifications", func() {
			s := new(testSession)
			s.onMessage = func() {}
			ro := NewBasicRO(23, s)
			err := ro.SendNotification("foo", []interface{}{1, "2", 3})
			So(err, ShouldBeNil)
			So(s.msg, ShouldResemble, []interface{}{uint64(23), uint64(3), "foo", []interface{}{1, "2", 3}})
		})

		Convey("does rom calls", func() {
			s := new(testSession)
			ro := NewBasicRO(23, s)
			var rcvErr error
			s.onMessage = func() {
				rcvErr = ro.receive([]interface{}{uint64(2), uint64(0), uint64(0), 42})
			}
			res, err := ro.CallROM("foo", []interface{}{1, "2", 3})
			So(rcvErr, ShouldBeNil)
			So(s.msg, ShouldResemble, []interface{}{uint64(23), uint64(1), uint64(0), "foo", []interface{}{1, "2", 3}})
			So(err, ShouldBeNil)
			So(res, ShouldEqual, 42)
		})

		Convey("errors from rom calls", func() {
			s := new(testSession)
			ro := NewBasicRO(23, s)
			var rcvErr error
			s.onMessage = func() {
				rcvErr = ro.receive([]interface{}{uint64(2), uint64(0), uint64(2), "boom"})
			}
			res, err := ro.CallROM("foo", []interface{}{1, "2", 3})
			So(rcvErr, ShouldBeNil)
			So(s.msg, ShouldResemble, []interface{}{uint64(23), uint64(1), uint64(0), "foo", []interface{}{1, "2", 3}})
			So(err, ShouldNotBeNil)
			So(res, ShouldBeNil)
		})
	})
}
