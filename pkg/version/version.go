package version

import "fmt"

var Version = "unknown"
var Commit = "unknown"

func Full() string {
	return fmt.Sprintf("%s (commit %s)", Version, Commit)
}
