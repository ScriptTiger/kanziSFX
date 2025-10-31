package main

import (
	"os"
	"strings"

	"github.com/ScriptTiger/kanziSFX"
)

// Function to display help text and exit
func help(err int) {
	os.Stdout.WriteString(
		"Usage: kanzisfx [options...]\n"+
		" -knz                 Output Kanzi bit stream\n"+
		" -o <file|directory>  Destination file or directory\n"+
		" -info                Show Kanzi bit stream info\n",
	)
	os.Exit(err)
}

func main() {

	// Check for invalid number of arguments
	if len(os.Args) > 4 {
		help(1)
	}

	var (
		outNamePtr *string
		infoStringPtr *string
		knzenc bool
		orw bool
		info bool
	)

	outNamePtr = new(string)
	infoStringPtr = new(string)

	// Push arguments to variables and pointers
	for i := 1; i < len(os.Args); i++ {
		if strings.HasPrefix(os.Args[i], "-") {
			switch strings.TrimPrefix(os.Args[i], "-") {
				case "knz":
					if knzenc {help(2)}
					knzenc = true
					continue
				case "o":
					if orw {help(3)}
					i++
					outNamePtr = &os.Args[i]
					orw = true
					continue
				case "info":
					if info {help(4)}
					info = true
					continue
				default:
					help(5)
			}
		} else {help(6)}
	}

	if *outNamePtr != "-" && !info {os.Stdout.WriteString("Checking Kanzi bit stream...\n")}

	err := kanziSFX.Extract(outNamePtr, infoStringPtr, knzenc, orw, info, true)

	os.Stdout.WriteString(*infoStringPtr)

	if err != nil {
		if err.Error() == kanziSFX.TAR_STDOUT_ERR {help(7)}
		os.Stdout.WriteString(err.Error()+"\n")
	}

}