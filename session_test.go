package univedo

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
	"time"
)

const testURL = "ws://localhost:9000/f8018f09-fb75-4d3d-8e11-44b2dc796130"

func pingTest(v interface{}) {
	session, err := Dial(testURL)
	So(err, ShouldBeNil)
	So(session, ShouldNotBeNil)
	pong, err := session.Ping(v)
	So(err, ShouldBeNil)
	So(pong, ShouldResemble, v)
	session.Close()
}

func TestSession(t *testing.T) {
	Convey("session", t, func() {
		Convey("connects", func() {
			session, err := Dial(testURL)
			So(err, ShouldBeNil)
			So(session, ShouldNotBeNil)
			session.Close()
		})

		Convey("pings null", func() {
			pingTest(nil)
		})

		Convey("pings true", func() {
			pingTest(true)
		})

		Convey("pings false", func() {
			pingTest(false)
		})

		Convey("pings ints", func() {
			pingTest(uint64(42))
		})

		Convey("pings negative ints", func() {
			pingTest(int64(-42))
		})

		Convey("pings floats", func() {
			pingTest(1.1)
		})

		Convey("pings floats", func() {
			pingTest(1.1)
		})

		Convey("pings strings", func() {
			pingTest("foobar")
		})

		Convey("pings arrays", func() {
			pingTest([]interface{}{"1", "2"})
		})

		Convey("pings times", func() {
			t, _ := time.Parse(time.RFC3339Nano, "2013-03-21T20:04:00.000001Z")
			pingTest(t)
		})
	})
}
