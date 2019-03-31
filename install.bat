@echo off

echo Installing the dependencies
go get -u github.com/mailru/easyjson/...
if %errorlevel% neq 0 (
	pause
	exit /b %errorlevel%
)

easyjson -pkg common
if %errorlevel% neq 0 (
	pause
	exit /b %errorlevel%
)

go get
if %errorlevel% neq 0 (
	pause
	exit /b %errorlevel%
)

echo Building the installer
go generate
go build "./cmd/install"
if %errorlevel% neq 0 (
	pause
	exit /b %errorlevel%
)
install.exe
