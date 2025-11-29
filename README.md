[![Say Thanks!](https://img.shields.io/badge/Say%20Thanks-!-1EAEDB.svg)](https://docs.google.com/forms/d/e/1FAIpQLSfBEe5B_zo69OBk19l3hzvBmz3cOV6ol1ufjh0ER1q3-xd2Rg/viewform)

# kanziSFX: Kanzi SelF-eXtracting archive
kanziSFX is a minimal Kanzi decompressor package to decompress an embedded Kanzi bit stream, with built-in support to also untar a compressed tar archive. So, the embedded Kanzi bit stream could contain either a single file of arbitrary type or a tar archive which will be both decompressed and unarchived.

To import kanziSFX into your project:  
`go get github.com/ScriptTiger/kanziSFX`  
Then just `import "github.com/ScriptTiger/kanziSFX"` and get started with using its functions.

Please refer to the dev package docs and reference implementations for more details and ideas on how to integrate kanziSFX into your project.  

Dev package docs:  
https://pkg.go.dev/github.com/ScriptTiger/kanziSFX

Reference implementations:  
https://github.com/ScriptTiger/kanziSFX/blob/main/_ref

# CLI Reference Implementation

Usage: `kanzisfx [options...]`
Argument                  | Description
--------------------------|-----------------------------------------------------------------------------------------------------
 `-knz`                   | Output Kanzi bit stream
 `-o <file\|directory>`   | Destination file or directory
 `-info`                  | Show Kanzi bit stream info

`-` can be used in place of `<file>` to designate standard output as the destination, but cannot be used in place of a directory for extracting tar archives.

Without any arguments, the embedded Kanzi stream will be decompressed into the working directory to a file of the same name as the executable, except with the `.exe` or `.app` extension removed. Or, if the Kanzi stream contains a tar archive, the tar archive will be both decompressed and unarchived into a folder within the working directory of the same name as the executable, except with the `.exe` or `.app` extension removed. So, command-line usage is only optional and the end user can just execute the application as they would any other application for this default behavior.

# GUI Reference Implementation

[![ScriptTiger/kanziSFX](https://scripttiger.github.io/images/kanziSFX-Interface.png)](https://github.com/ScriptTiger/kanziSFX)

The cno GUI package was used for its extremely minimal and lightweight nature to allow kanziSFX to retain a small footprint, as is necessary for a self-extracting archive decompressor. However, cno only supports native Windows GUIs at the moment, so this reference implementation is only available for Windows.

For additional notes on cno, please refer to its documentation:  
https://github.com/ScriptTiger/cno

# Appending a Kanzi archive to a kanziSFX executable
Compile a kanziSFX executable from source or download the latest pre-built release for the intended target system:  
https://github.com/ScriptTiger/kanziSFX/releases/latest

For appending a Kanzi archive to a kanziSFX executable, issue one of the following commands.

For Windows:
```
copy /b "kanziSFX.exe"+"file.knz" "MyKanziSFX.exe"
```

For Linux and Mac:
```
cat "kanziSFX" "file.knz" > "MyKanziSFX"
```

# More About ScriptTiger

For more ScriptTiger scripts and goodies, check out ScriptTiger's GitHub Pages website:  
https://scripttiger.github.io/
