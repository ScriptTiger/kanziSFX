package kanziSFX

import (
	"archive/tar"
	"bufio"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	kanzi "github.com/flanglet/kanzi-go/v2/io"
)

// Currently supported bit stream version (backwards compatible)
const BIT_STREAM_VERSION = 6

// Error constants
const TAR_STDOUT_ERR = "Cannot output TAR to standard output"

// Public extract function
func Extract(outNamePtr, infoStringPtr *string, knzenc, orw, info, verbose bool) (err error) {

	// Locate executable
	filePath, err := os.Executable()
	if err != nil {return err}
	filePath, err = filepath.EvalSymlinks(filePath)
	if err != nil {return err}

	// Open file
	sfxFile, err := os.Open(filePath)
	if err != nil {return err}
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
			sfxFile.Close()
			return errors.New("No Kanzi stream found")
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
		sfxFile.Close()
		return errors.New(
			"The Kanzi bit stream is version "+strconv.Itoa(version)+"!\n"+
			"Your current version of KanziSFX can only support decompressing bit streams up to version "+
			strconv.Itoa(BIT_STREAM_VERSION)+"!\n",
		)
	}

	// Create a Kanzi reader for the Kanzi stream
	sfxFile.Seek(sfxSize, io.SeekStart)
	knzReader, err := kanzi.NewReader(sfxFile, 4)
	if err != nil {return err}

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

	// Return if there is a tar and output is Stdout
	if *outNamePtr == "-" && isTar && !knzenc  {return errors.New(TAR_STDOUT_ERR)}

	// Build info string and return
	if info {
		sfxFile.Close()
		*infoStringPtr = "bit_stream_version="+strconv.Itoa(version)+"\n"+
				 "tar="+strconv.FormatBool(isTar)+"\n"
		return nil
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
		output, err = os.Create(*outNamePtr)
		if err != nil {return err}
		if verbose {os.Stdout.WriteString("Extracting \""+*outNamePtr+"\"...\n")}
		defer output.Close()
	}

	// If knz flag set, dump Kanzi stream and return
	if knzenc {
		sfxFile.Seek(sfxSize, io.SeekStart)
		io.Copy(output, sfxFile)
		sfxFile.Close()
		output.Close()
		return nil
	}

	// Decompress Kanzi stream, and unarchive tar if applicable
	sfxFile.Seek(sfxSize, io.SeekStart)
	knzReader, err = kanzi.NewReader(sfxFile, 4)
	if err != nil {return err}
	if isTar {
		tarReader := tar.NewReader(knzReader)
		os.MkdirAll(*outNamePtr, 0755)
		for {
			tarHeader, err := tarReader.Next()
			if err != nil {break}
			name := filepath.Join(*outNamePtr, tarHeader.Name)
			if tarHeader.Typeflag == tar.TypeDir {os.Mkdir(name, 0755)
			} else {
				if verbose {os.Stdout.WriteString("Extracting "+name+"...\n")}
				outputTar, err := os.Create(name)
				if err != nil {return err}
				io.Copy(outputTar, tarReader)
				outputTar.Close()
			}
		}
	} else {io.Copy(output, knzReader)}
	return nil
}
