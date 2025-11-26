package main

import (
	"os"
	"strconv"
	"strings"

	. "github.com/ScriptTiger/kanziSFX"
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
		knzenc bool
		orw bool
		info bool
	)

	outNamePtr = new(string)

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

	// Validate options
	if info && (knzenc || orw) {help(7)}

	if *outNamePtr != "-" && !info {os.Stdout.WriteString("Checking Kanzi bit stream...\n")}

	// Initialize map for info
	ctx := make(map[string]any)

	// Set flags to pass to kanziSFX
	var ops uint8
	if knzenc {ops|=OUTPUT_KANZI}
	if orw {ops|=REWRITE_PATH}
	if info {ops|=INFO}

	// Call kanziSFX 
	err := Extract(outNamePtr, ctx, ops|VERBOSE)

	// Output CTX data if given
	if tar, hasKey := ctx["tar"]; hasKey {
		var sb strings.Builder
		sb.WriteString("tar="+strconv.FormatBool(tar.(bool))+"\n")
		if value, hasKey := ctx["bsVersion"]; hasKey {
			sb.WriteString("bit_stream_version="+strconv.Itoa(int(value.(uint)))+"\n")
		}
		if value, hasKey := ctx["blockSize"]; hasKey {
			sb.WriteString("block_size="+strconv.Itoa(int(value.(uint)))+"\n")
		}
		if value, hasKey := ctx["entropy"]; hasKey {
			sb.WriteString("entropy="+value.(string)+"\n")
		}
		if value, hasKey := ctx["transform"]; hasKey {
			sb.WriteString("transform="+value.(string)+"\n")
		}
		if value, hasKey := ctx["outputSize"]; hasKey {
			sb.WriteString("output_size="+strconv.FormatInt(value.(int64), 10)+"\n")
		}
		if value, hasKey := ctx["jobs"]; hasKey {
			sb.WriteString("jobs="+strconv.Itoa(int(value.(uint)))+"\n")
		}
		os.Stdout.WriteString(sb.String())
	}

	// Output if there's an error
	if err != nil {
		if err.Error() == TAR_STDOUT_ERR {help(8)}
		os.Stdout.WriteString(err.Error()+"\n")
	}
}
