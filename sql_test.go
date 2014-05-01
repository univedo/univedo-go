package univedo

import (
	. "github.com/smartystreets/goconvey/convey"

	"database/sql"
	"testing"
)

const testPerspectiveURL = "ws://vagrant:80/f8018f09-fb75-4d3d-8e11-44b2dc796130/cefb4ed2-4ce3-4825-8550-b68a3c142f0a"

func TestSql(t *testing.T) {
	Convey("sql", t, func() {
		Convey("connects", func() {
			db, err := sql.Open("univedo", testPerspectiveURL)
			So(err, ShouldBeNil)
			So(db.Ping(), ShouldBeNil)
			db.Close()
		})

		Convey("runs empty selects", func() {
			db, err := sql.Open("univedo", testPerspectiveURL)
			So(err, ShouldBeNil)
			rows, err := db.Query("select * from dummy where dummy_uuid = '1AF6B99E-5908-4516-A5FA-B22AFD27E003'")
			So(err, ShouldBeNil)
			So(rows.Next(), ShouldBeFalse)
		})

		Convey("runs selects", func() {
			db, err := sql.Open("univedo", testPerspectiveURL)
			So(err, ShouldBeNil)
			rows, err := db.Query("select * from fields_inclusive")
			So(err, ShouldBeNil)
			i := 0
			for rows.Next() {
				i++
			}
			So(i, ShouldBeGreaterThan, 100)
		})

		Convey("selects counts", func() {
			db, err := sql.Open("univedo", testPerspectiveURL)
			So(err, ShouldBeNil)
			rows, err := db.Query("select count(*) from fields_inclusive")
			So(err, ShouldBeNil)
			cols, err := rows.Columns()
			So(err, ShouldBeNil)
			So(cols, ShouldResemble, []string{"COUNT(*)"})
			var c int
			for rows.Next() {
				err = rows.Scan(&c)
				So(err, ShouldBeNil)
			}
			So(rows.Next(), ShouldBeFalse)
			So(c, ShouldBeGreaterThan, 100)
		})

		Convey("inserts", func() {
			db, err := sql.Open("univedo", testPerspectiveURL)
			So(err, ShouldBeNil)
			result, err := db.Exec("insert into dummy (dummy_int8) values (?)", 42)
			So(err, ShouldBeNil)
			id, err := result.LastInsertId()
			So(err, ShouldBeNil)
			rows, err := db.Query("select id, dummy_int8 from dummy where id = ?", id)
			So(err, ShouldBeNil)
			var id2, dummyInt int64
			for rows.Next() {
				err = rows.Scan(&id2, &dummyInt)
			}
			So(rows.Next(), ShouldBeFalse)
			So(id2, ShouldEqual, id)
			So(dummyInt, ShouldEqual, 42)
		})
	})
}
