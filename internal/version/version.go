package version

import (
	"fmt"
	"runtime"
	"strconv"
	"time"
)

var Version string = "0.0.0"
var GitCommit string = "unknown"
var BuildTime string = "unknown"

func ToString() string {

	built := BuildTime
	i, err := strconv.ParseInt(BuildTime, 10, 64)
	if err == nil {
		t := time.Unix(i, 0)
		built = t.Format(time.RFC3339)
	}
	return fmt.Sprintf("%s (git=%s) (built %s) using %s\n", Version, GitCommit, built, runtime.Version())
}
