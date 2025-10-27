package main

import (
	"archive/tar"
	"bufio"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	kanzi "github.com/flanglet/kanzi-go/v2/io"
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

	// Currently supported bit stream version (backwards compatible)
	const BIT_STREAM_VERSION = 6

	// Check for invalid number of arguments
	if len(os.Args) > 4 {
		help(1)
	}

	var (
		outNamePtr *string
		knzenc bool
		orw bool
		info bool
		err error
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

	if *outNamePtr != "-" && !info {os.Stdout.WriteString("Checking Kanzi bit stream...\n")}

	// Locate executable
	filePath, _ := os.Executable()
	filePath, _ = filepath.EvalSymlinks(filePath)

	// Open file
	sfxFile, _ := os.Open(filePath)
	defer sfxFile.Close()

	// Determine length of KanziSFX / start of Kanzi stream
	sfxSize := int64(1500000)
	sfxFile.Seek(sfxSize, io.SeekStart)
	sfxReader := bufio.NewReader(sfxFile)
	knzMagic := make([]byte, 5)
	for {
		for i := 0; i < 4; i++ {knzMagic[i] = knzMagic[i+1]}
		knzMagic[4], err = sfxReader.ReadByte()

		if err != nil {
			os.Stdout.WriteString("No Kanzi stream found.\n")
			sfxFile.Close()
			os.Exit(7)
		}

		if string(knzMagic) == "\x00KANZ" {break}

		sfxSize++
	}

	// Roll back sfxSize to beginning of Kanzi stream / end of sfx
	sfxSize = sfxSize-3

	// Determine bit stream version
	readByte := make([]byte, 1)
	sfxFile.Seek(sfxSize+4, io.SeekStart)
	sfxFile.Read(readByte)
	version := int(readByte[0]>>4)
	if version > BIT_STREAM_VERSION && !knzenc {
		os.Stdout.WriteString(
			"The Kanzi bit stream is version "+strconv.Itoa(version)+"!\n"+
			"Your current version of KanziSFX can only support decompressing bit streams up to version "+
			strconv.Itoa(BIT_STREAM_VERSION)+"!\n",
		)
		sfxFile.Close()
		os.Exit(8)
	}

	// Create a Kanzi reader for the Kanzi stream
	sfxFile.Seek(sfxSize, io.SeekStart)
	knzReader, _ := kanzi.NewReader(sfxFile, 4)

	// Determine if tar archive is present
	tarSeeker := bufio.NewReader(knzReader)
	var isTar bool
	tarMagic := make([]byte, 6)
	for {
		for i := 0; i < 5; i++ {tarMagic[i] = tarMagic[i+1]}
		tarMagic[5], err = tarSeeker.ReadByte()

		if err != nil {break}
		if string(tarMagic) == "\x00ustar" {
			isTar = true
			break
		}
	}

	// Exit if there is a tar and output is Stdout
	if *outNamePtr == "-" && isTar && !knzenc  {help(7)}

	// Show info and exit
	if info {
		// Determine bit length of uncompressed file size
		sfxFile.Seek(sfxSize+14, io.SeekStart)
		sfxFile.Read(readByte)
		sizeBytes := uint8(readByte[0]&0x03)

		// Determine file size
		var size uint64
		sizeBuffer := make([]byte, (sizeBytes&0x03)*2)
		sfxFile.Read(sizeBuffer)
		for _, sizeByte := range sizeBuffer {size = (size<<8)|uint64(sizeByte)}

		os.Stdout.WriteString(
			"bit_stream_version="+strconv.Itoa(version)+"\n"+
			"uncompressed_byte_size="+strconv.FormatUint(size, 10)+"\n"+
			"tar="+strconv.FormatBool(isTar)+"\n",
		)
		sfxFile.Close()
		os.Exit(0)
	}

	// Rewrite file/directory name as needed
	if !orw {
		if knzenc {*outNamePtr = strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filePath))+".knz"
		} else {*outNamePtr = strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filePath))}
	}

	// Create output file
	var output *os.File
	if *outNamePtr == "-" {output = os.Stdout
	} else if !isTar || (isTar && knzenc) {
		output, _ = os.Create(*outNamePtr)
		os.Stdout.WriteString("Extracting \""+*outNamePtr+"\"...\n")
		defer output.Close()
	}

	// If knz flag set, dump Kanzi stream and exit
	if knzenc {
		sfxFile.Seek(sfxSize, io.SeekStart)
		io.Copy(output, sfxFile)
		sfxFile.Close()
		output.Close()
		os.Exit(0)
	}

	// Decompress Kanzi stream, and unarchive tar if applicable
	sfxFile.Seek(sfxSize, io.SeekStart)
	knzReader, _ = kanzi.NewReader(sfxFile, 4)
	if isTar {
		tarReader := tar.NewReader(knzReader)
		os.MkdirAll(*outNamePtr, 0755)
		for {
			tarHeader, err := tarReader.Next()
			if err != nil {break}
			name := filepath.Join(*outNamePtr, tarHeader.Name)
			if tarHeader.Typeflag == tar.TypeDir {os.Mkdir(name, 0755)
			} else {
				os.Stdout.WriteString("Extracting "+name+"...\n")
				outputTar, _ := os.Create(name)
				io.Copy(outputTar, tarReader)
				outputTar.Close()
			}
		}
	} else {
		io.Copy(output, knzReader)
	}

}
