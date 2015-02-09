package univedo

import (
	"io/ioutil"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

const testURL = "ws://localhost:9011/"

func pingTest(v interface{}) {
	connection, err := Dial(testURL)
	So(err, ShouldBeNil)
	So(connection, ShouldNotBeNil)
	session, err := connection.GetSession("79CB0F8E-3D90-484A-9A88-B13E97FA65D9", map[string]interface{}{"username": "marvin"})
	So(err, ShouldBeNil)
	So(session, ShouldNotBeNil)
	pong, err := session.Ping(v)
	So(err, ShouldBeNil)
	So(pong, ShouldResemble, v)
	connection.Close()
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

		Convey("pings strings", func() {
			pingTest("foobar")
		})

		Convey("pings arrays", func() {
			pingTest([]interface{}{"1", "2"})
		})

		Convey("pings maps", func() {
			pingTest(map[string]interface{}{"1": "a", "2": "b"})
		})

		Convey("pings times", func() {
			t, _ := time.Parse(time.RFC3339Nano, "2013-03-21T20:04:00.000001Z")
			pingTest(t)
		})

		Convey("apply uts", func() {
			connection, err := Dial(testURL)
			So(err, ShouldBeNil)
			So(connection, ShouldNotBeNil)
			session, err := connection.GetSession("79CB0F8E-3D90-484A-9A88-B13E97FA65D9", map[string]interface{}{"username": "marvin"})
			So(err, ShouldBeNil)
			So(session, ShouldNotBeNil)
			testFile, err := ioutil.ReadFile("test.uts")
			So(err, ShouldBeNil)
			err = session.ApplyUTS(string(testFile))
			So(err, ShouldBeNil)
			defer connection.Close()
		})
	})
}
