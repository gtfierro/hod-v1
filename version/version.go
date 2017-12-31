package version

import (
	"fmt"
)

var Commit = "unset"
var Release = "unset"
var LOGO = fmt.Sprintf(" _    _           _ _____  ____ \n"+
	"| |  | |         | |  __ \\|  _ \\  \n"+
	"| |__| | ___   __| | |  | | |_) | \n"+
	"|  __  |/ _ \\ / _` | |  | |  _ <  \n"+
	"| |  | | (_) | (_| | |__| | |_) | \n"+
	"|_|  |_|\\___/ \\__,_|_____/|____/  \n"+
	"Commit: %s   Release: %s\n", Commit, Release)
