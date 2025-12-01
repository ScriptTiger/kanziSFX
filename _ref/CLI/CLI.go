package main

import (
	"os"
	"strconv"
	"strings"
	"time"

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

// Function to display error and exit
func displayError(err error) {
	if err.Error() == TAR_STDOUT_ERR {help(8)}
	os.Stdout.WriteString("\r"+err.Error()+"\n")
	os.Exit(-1)
}

// Function to display progress
func printProgress(progress *[2]int64) {
	if progress[1] != 0 {
		os.Stdout.WriteString(
			"\r"+
			strconv.Itoa(int((float64(progress[0])/float64(progress[1]))*100))+
			"% | "+strconv.FormatInt(progress[0], 10)+" bytes of "+strconv.FormatInt(progress[1], 10),
		)
	} else {os.Stdout.WriteString("\r--% | "+strconv.FormatInt(progress[0], 10)+" bytes of --")}
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

	// If info requested, initialize map to receive CTX
	// If not, initialize progress tracker to track progress of extraction if not extracting to standard output
	var ctx map[string]any
	var progress *[2]int64
	if info {ctx = make(map[string]any)
	} else if *outNamePtr != "-" {progress = new([2]int64)}

	// Set flags to pass to kanziSFX
	var ops uint8
	if knzenc {ops|=OUTPUT_KANZI}
	if orw {ops|=REWRITE_PATH}
	if info {ops|=INFO}

	// Call kanziSFX, within a go routine if progress tracking is needed and without if not
	var running bool
	var err error
	if *outNamePtr == "-" || info {
		err = Extract(outNamePtr, accelerator, ctx, nil, ops)
		if err != nil {displayError(err)}
	} else {
		go func() {
			running = true
			err = Extract(outNamePtr, accelerator, nil, progress, ops)
			if err != nil {displayError(err)}
			running = false
		}()
	}

	// If extracting to a file and not standard output, display progress
	// If info requested and CTX map given, output CTX data
	if *outNamePtr != "-" && !info {
		os.Stdout.WriteString("--% | -- bytes of --")
		for ; running; {
			if progress[0] != 0 {printProgress(progress)}
			time.Sleep(100*time.Millisecond)
		}
		if progress[0] != 0 {
			printProgress(progress)
			os.Stdout.WriteString("\nExtraction complete!\n")
		}
	} else if info {
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
			if value, hasKey := ctx["inputSize"]; hasKey {
				sb.WriteString("input_size="+strconv.FormatInt(value.(int64), 10)+"\n")
			}
			if value, hasKey := ctx["outputSize"]; hasKey {
				sb.WriteString("output_size="+strconv.FormatInt(value.(int64), 10)+"\n")
			}
			if value, hasKey := ctx["jobs"]; hasKey {
				sb.WriteString("jobs="+strconv.Itoa(int(value.(uint)))+"\n")
			}
			os.Stdout.WriteString(sb.String())
		}
	}
}
