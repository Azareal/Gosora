@echo off

echo Updating the dependencies
go get
if %errorlevel% neq 0 (
	pause
	exit /b %errorlevel%
)

echo Building the updater
go generate
go build ./updater
if %errorlevel% neq 0 (
	pause
	exit /b %errorlevel%
)
updater.exe
