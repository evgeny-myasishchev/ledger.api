package tags

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestGetTagIDsFromString(t *testing.T) {
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	Convey("Given braces enclosed string", t, func() {
		Convey("When string is empty", func() {
			Convey("It should return empty slice", func() {
				So(GetTagIDsFromString(""), ShouldResemble, []int{})
			})
		})

		Convey("When single tagID is enclosed", func() {
			tagID := rnd.Intn(1000)
			Convey("It should return a slice with single item", func() {
				So(GetTagIDsFromString(fmt.Sprintf("{%v}", tagID)), ShouldResemble, []int{tagID})
			})
		})

		Convey("When multiple tagIDs are separated with comma", func() {
			tagID1 := rnd.Intn(1000)
			tagID2 := rnd.Intn(1000)
			tagID3 := rnd.Intn(1000)
			Convey("It should return a slice with all values parsed", func() {
				So(
					GetTagIDsFromString(fmt.Sprintf("{%v},{%v},{%v}", tagID1, tagID2, tagID3)),
					ShouldResemble,
					[]int{tagID1, tagID2, tagID3})
			})
		})
	})
}
