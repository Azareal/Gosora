@echo off
echo Installing the dependencies

echo Installing the MySQL Driver
go get -u github.com/go-sql-driver/mysql
if %errorlevel% neq 0 (
	pause
	exit /b %errorlevel%
)

echo Installing the PostgreSQL Driver
go get -u github.com/lib/pq
if %errorlevel% neq 0 (
	pause
	exit /b %errorlevel%
)

echo Installing the MSSQL Driver
go get -u github.com/denisenkom/go-mssqldb
if %errorlevel% neq 0 (
	pause
	exit /b %errorlevel%
)

echo Installing the bcrypt library
go get -u golang.org/x/crypto/bcrypt
if %errorlevel% neq 0 (
	pause
	exit /b %errorlevel%
)

echo Installing the Argon2 library
go get -u golang.org/x/crypto/argon2
if %errorlevel% neq 0 (
	pause
	exit /b %errorlevel%
)

echo Installing /x/sys/windows (dependency for gopsutil)
go get -u golang.org/x/sys/windows
if %errorlevel% neq 0 (
	pause
	exit /b %errorlevel%
)

echo Installing wmi (dependency for gopsutil)
go get -u github.com/StackExchange/wmi
if %errorlevel% neq 0 (
	pause
	exit /b %errorlevel%
)

echo Installing the gopsutil library
go get -u github.com/Azareal/gopsutil
if %errorlevel% neq 0 (
	pause
	exit /b %errorlevel%
)

echo Installing the WebSockets library
go get -u github.com/gorilla/websocket
if %errorlevel% neq 0 (
	pause
	exit /b %errorlevel%
)

echo Installing Sourcemap (dependency for OttoJS)
go get -u gopkg.in/sourcemap.v1
if %errorlevel% neq 0 (
	pause
	exit /b %errorlevel%
)

echo Installing the OttoJS
go get -u github.com/robertkrimen/otto
if %errorlevel% neq 0 (
	pause
	exit /b %errorlevel%
)

echo Installing the Rez Image Resizer
go get -u github.com/bamiaux/rez
if %errorlevel% neq 0 (
	pause
	exit /b %errorlevel%
)

echo Installing Caire
go get -u github.com/esimov/caire
if %errorlevel% neq 0 (
	pause
	exit /b %errorlevel%
)

echo Installing some error helpers
go get -u github.com/pkg/errors
if %errorlevel% neq 0 (
	pause
	exit /b %errorlevel%
)

echo Installing fsnotify
go get -u github.com/fsnotify/fsnotify
if %errorlevel% neq 0 (
	pause
	exit /b %errorlevel%
)


echo Building the installer
go generate
go build ./install
if %errorlevel% neq 0 (
	pause
	exit /b %errorlevel%
)
install.exe
