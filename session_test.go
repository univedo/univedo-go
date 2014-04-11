package univedo

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

const URL = "ws://localhost:9000/f8018f09-fb75-4d3d-8e11-44b2dc796130"

func TestSession(t *testing.T) {
	Convey("session", t, func() {
		Convey("connects", func() {
			session, err := Dial(URL)
			So(err, ShouldBeNil)
			So(session, ShouldNotBeNil)
			session.Close()
		})

		Convey("pings null", func() {
			session, err := Dial(URL)
			So(err, ShouldBeNil)
			So(session, ShouldNotBeNil)
			pong, err := session.Ping(nil)
			So(err, ShouldBeNil)
			So(pong, ShouldBeNil)
			session.Close()
		})
	})
}
