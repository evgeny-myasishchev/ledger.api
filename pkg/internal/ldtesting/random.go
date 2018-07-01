package ldtesting

import (
	"math/rand"
	"time"
)

var timeMin = time.Date(1970, 1, 0, 0, 0, 0, 0, time.UTC).Unix()
var timeMax = time.Date(2070, 1, 0, 0, 0, 0, 0, time.UTC).Unix()
var timeDelta = timeMax - timeMin

// RandomDate returns random date between 1970 and 2070
func RandomDate() time.Time {
	sec := rand.Int63n(timeDelta) + timeMin
	return time.Unix(sec, 0)
}
