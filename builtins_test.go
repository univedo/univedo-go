package univedo

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func shouldBeOk(v interface{}, err error) {
	So(err, ShouldBeNil)
	So(v, ShouldNotBeNil)
}

func TestBuiltins(t *testing.T) {
	Convey("builtins", t, func() {
		Convey("gets perspectives", func() {
			session, err := Dial(testURL)
			shouldBeOk(session, err)
			perspective, err := session.GetPerspective("cefb4ed2-4ce3-4825-8550-b68a3c142f0a")
			shouldBeOk(perspective, err)
			session.Close()
		})

		Convey("runs empty selects", func() {
			session, err := Dial(testURL)
			shouldBeOk(session, err)
			perspective, err := session.GetPerspective("cefb4ed2-4ce3-4825-8550-b68a3c142f0a")
			shouldBeOk(perspective, err)
			query, err := perspective.Query()
			shouldBeOk(query, err)
			statement, err := query.Prepare("select * from dummy where dummy_uuid = 'foo'")
			shouldBeOk(statement, err)
			result, err := statement.Execute()
			shouldBeOk(result, err)
			session.Close()
		})
	})
}
