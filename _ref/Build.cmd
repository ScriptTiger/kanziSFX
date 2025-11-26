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
call :Build_App

if %dev% == 1 exit /b

set GOOS=linux
set EXT=
call :Build_App

set GOOS=darwin
set EXT=.app
call :Build_App

exit /b

:Build_App

set app=CLI
set source=CLI
set flags=-s -w
call :Build

if %dev% == 1 exit /b
if not %GOOS% == windows exit /b
exit /b

cd ..\GUI

if exist go.mod (
	choice /m "Rebuild GUI go.mod and go.sum?"
	if !errorlevel! == 1 (del go.mod go.sum)
)

if not exist go.mod (
	echo Initializing go module...
	go mod init main 2> nul
	go mod tidy 2> nul
)

set app=GUI
set source=GUI_Windows
set flags=-s -w -H=windowsgui
call :Build

cd ..\CLI

exit /b

:Build
echo Building %mod%_%app%_%GOOS%_%GOARCH%%EXT%...
go build -ldflags="%flags%" -o "../Release/%mod%_%app%_%GOOS%_%GOARCH%%EXT%" %source%.go
exit /b