@echo off
setlocal ENABLEDELAYEDEXPANSION

set mod=kanziSFX
if not exist "Release" md "Release"

cd ..

call :mod module github.com/ScriptTiger/%mod%

cd _ref\CLI

call :mod CLI main

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
call :Build_App

if %dev% == 1 exit /b

set GOOS=linux
set EXT=.elf
call :Build_App

set GOOS=darwin
set EXT=.app
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

call :mod "%GOOS% GUI" main

set app=GUI
if %GOOS% == windows set flags=-s -w -H=windowsgui
if %GOOS% == linux set flags=-s -w
set RELEASE=../../Release
call :Build

cd ..\..\CLI

exit /b

:Build
echo Building %mod%_%app%_%GOOS%_%GOARCH%%EXT%...
go build -ldflags="%flags%" -o "%RELEASE%/%mod%_%app%_%GOOS%_%GOARCH%%EXT%"
if %errorlevel% == 0 if not %GOOS% == darwin call upx --lzma "%RELEASE%/%mod%_%app%_%GOOS%_%GOARCH%%EXT%" 1> nul
exit /b

:mod
if exist go.mod (
	choice /m "Rebuild %~1 go.mod and go.sum?"
	if !errorlevel! == 1 (del go.mod go.sum)
)
if not exist go.mod (
	echo Building go.mod and go.sum...
	go mod init %2 2> nul
	go mod tidy 2> nul
)
exit /b