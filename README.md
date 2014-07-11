# univedo-go

Go client for univedo

## Usage

```go
const testURL = "ws://localhost/f8018f09-fb75-4d3d-8e11-44b2dc796130/cefb4ed2-4ce3-4825-8550-b68a3c142f0a"

db, _ := sql.Open("univedo", testURL)
rows, _ := db.Query("select * from dummy")
i := 0
for rows.Next() {
  i++
}
println(i)
```
