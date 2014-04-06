package univedo

import (
	"bytes"
	"github.com/nu7hatch/gouuid"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
	"time"
)

func readMessage(s string) (interface{}, error) {
	m := &message{buffer: bytes.NewBufferString(s)}
	return m.read()
}

func TestCbor(t *testing.T) {
	Convey("cbor reads", t, func() {
		Convey("null", func() {
			r, err := readMessage("\xf6")
			So(r, ShouldBeNil)
			So(err, ShouldBeNil)
		})

		Convey("true", func() {
			r, err := readMessage("\xf5")
			So(r, ShouldBeTrue)
			So(err, ShouldBeNil)
		})

		Convey("false", func() {
			r, err := readMessage("\xf4")
			So(r, ShouldBeFalse)
			So(err, ShouldBeNil)
		})

		Convey("float32", func() {
			r, err := readMessage("\xfa\x47\xc3\x50\x00")
			So(r, ShouldEqual, 100000.0)
			So(err, ShouldBeNil)
		})

		Convey("float64", func() {
			r, err := readMessage("\xfb\x3f\xf1\x99\x99\x99\x99\x99\x9a")
			So(r, ShouldEqual, 1.1)
			So(err, ShouldBeNil)
		})

		Convey("small uint", func() {
			r, err := readMessage("\x0a")
			So(r, ShouldEqual, 10)
			So(err, ShouldBeNil)
		})

		Convey("uint8", func() {
			r, err := readMessage("\x18\x2a")
			So(r, ShouldEqual, 42)
			So(err, ShouldBeNil)
		})

		Convey("uint16", func() {
			r, err := readMessage("\x19\x03\xe8")
			So(r, ShouldEqual, 1000)
			So(err, ShouldBeNil)
		})

		Convey("uint32", func() {
			r, err := readMessage("\x1a\x00\x0f\x42\x40")
			So(r, ShouldEqual, 1000000)
			So(err, ShouldBeNil)
		})

		Convey("uint64", func() {
			r, err := readMessage("\x1b\x00\x00\x00\xe8\xd4\xa5\x10\x00")
			So(r, ShouldEqual, 1000000000000)
			So(err, ShouldBeNil)
		})

		Convey("small int", func() {
			r, err := readMessage("\x20")
			So(r, ShouldEqual, -1)
			So(err, ShouldBeNil)
		})

		Convey("int8", func() {
			r, err := readMessage("\x38\x63")
			So(r, ShouldEqual, -100)
			So(err, ShouldBeNil)
		})

		Convey("int16", func() {
			r, err := readMessage("\x39\x03\xe7")
			So(r, ShouldEqual, -1000)
			So(err, ShouldBeNil)
		})

		Convey("int32", func() {
			r, err := readMessage("\x3a\x00\x0f\x42\x3f")
			So(r, ShouldEqual, -1000000)
			So(err, ShouldBeNil)
		})

		Convey("int64", func() {
			r, err := readMessage("\x3b\x00\x00\x00\xe8\xd4\xa5\x0f\xff")
			So(r, ShouldEqual, -1000000000000)
			So(err, ShouldBeNil)
		})

		Convey("strings", func() {
			r, err := readMessage("\x66foobar")
			So(err, ShouldBeNil)
			So(r, ShouldEqual, "foobar")
		})

		Convey("utf8 strings", func() {
			r, err := readMessage("\x67f\xc3\xb6obar")
			So(err, ShouldBeNil)
			So(r, ShouldEqual, "föobar")
		})

		Convey("blobs", func() {
			r, err := readMessage("\x46foobar")
			So(err, ShouldBeNil)
			So(r, ShouldResemble, []byte("foobar"))
		})

		Convey("lists", func() {
			r, err := readMessage("\x82\x63foo\x63bar")
			So(err, ShouldBeNil)
			So(r, ShouldResemble, []interface{}{"foo", "bar"})
		})

		Convey("maps", func() {
			r, err := readMessage("\xa2\x63bar\x02\x63foo\x01")
			So(err, ShouldBeNil)
			m := r.(map[string]interface{})
			So(len(m), ShouldEqual, 2)
			So(m["foo"], ShouldEqual, 1)
			So(m["bar"], ShouldEqual, 2)
		})

		Convey("records", func() {
			r, err := readMessage("\xc8\x18\x2a")
			So(err, ShouldBeNil)
			So(r, ShouldEqual, 42)
		})

		Convey("datetimes", func() {
			r, err := readMessage("\xc0\x74\x32\x30\x31\x33\x2d\x30\x33\x2d\x32\x31\x54\x32\x30\x3a\x30\x34\x3a\x30\x30\x5a")
			So(err, ShouldBeNil)
			ref, _ := time.Parse(time.RFC3339Nano, "2013-03-21T20:04:00Z")
			So(r, ShouldResemble, ref)
		})

		Convey("micro datetimes", func() {
			r, err := readMessage("\xc0\x78\x1b\x32\x30\x31\x33\x2d\x30\x33\x2d\x32\x31\x54\x32\x30\x3a\x30\x34\x3a\x30\x30.000001\x5a")
			So(err, ShouldBeNil)
			ref, _ := time.Parse(time.RFC3339Nano, "2013-03-21T20:04:00.000001Z")
			So(r, ShouldResemble, ref)
		})

		Convey("uuids", func() {
			r, err := readMessage("\xc7\x50\x68\x4E\xF8\x95\x72\xA2\x42\x98\xBC\x5B\x58\x0F\x1C\x1D\x27\x07")
			So(err, ShouldBeNil)
			ref, _ := uuid.ParseHex("684ef895-72a2-4298-bc5b-580f1c1d2707")
			So(r, ShouldResemble, ref)
		})
	})
}
