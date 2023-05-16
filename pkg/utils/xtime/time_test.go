package xtime

import (
	"fmt"
	"testing"
	"time"
)

func TestTimeAfter(t *testing.T) {
	ts := Parse("2023-05-16T03:43:16.669Z")
	fmt.Println(ts)
	ts2, _ := time.Parse(time.RFC3339, "2023-05-16T03:10:18.357Z")
	fmt.Println(ts2)

	FixedZone("Asia/Shanghai")
	now := Now()
	fmt.Println(now)

	fmt.Println(ts.After(now))
}
