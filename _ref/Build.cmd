@echo off
setlocal ENABLEDELAYEDEXPANSION

set mod=kanziSFX
if not exist "Release" md "Release"

cd ..

if exist go.mod (
	choice /m "Rebuild module go.mod and go.sum?"
	if !errorlevel! == 1 (del go.mod go.sum)
)

if not exist go.mod (
	echo Initializing go module...
	go mod init github.com/ScriptTiger/%mod% 2> nul
	go mod tidy 2> nul
)

cd _ref\CLI

choice /m "Dev build?"
if %errorlevel% == 1 (set dev=1) else set dev=0

set GOARCH=amd64
call :Build_OS

:Exit
pause
exit /b

:Build_OS

set GOOS=windows
set EXT=.exe
set INCLUDE=include_other.go
call :Build_App

if %dev% == 1 exit /b

set GOOS=linux
set EXT=.elf
set INCLUDE=include_other.go
call :Build_App

set GOOS=darwin
set EXT=.app
set INCLUDE=include_mac.go
call :Build_App

exit /b

:Build_App

set app=CLI
set flags=-s -w
set RELEASE=../Release
call :Build

if %dev% == 1 exit /b

if %GOOS% == darwin exit /b
if %GOOS% == windows cd ..\GUI\windows
if %GOOS% == linux cd ..\GUI\linux

if exist go.mod (
	choice /m "Rebuild %GOOS% GUI go.mod and go.sum?"
	if !errorlevel! == 1 (del go.mod go.sum)
)

if not exist go.mod (
	echo Initializing go module...
	go mod init main 2> nul
	go mod tidy 2> nul
)

set app=GUI
if %GOOS% == windows set flags=-s -w -H=windowsgui
if %GOOS% == linux set flags=-s -w
set INCLUDE=
set RELEASE=../../Release
call :Build

cd ..\..\CLI

exit /b

:Build
echo Building %mod%_%app%_%GOOS%_%GOARCH%%EXT%...
go build -ldflags="%flags%" -o "%RELEASE%/%mod%_%app%_%GOOS%_%GOARCH%%EXT%" %mod%.go %INCLUDE%
if %errorlevel% == 0 if not %GOOS% == darwin call upx --lzma "%RELEASE%/%mod%_%app%_%GOOS%_%GOARCH%%EXT%" 1> nul
exit /b