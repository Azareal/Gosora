@echo off
echo Updating the MySQL Driver
go get -u github.com/go-sql-driver/mysql
if %errorlevel% neq 0 (
	pause
	exit /b %errorlevel%
)
echo Updating the PostgreSQL Driver
go get -u github.com/lib/pq
if %errorlevel% neq 0 (
	pause
	exit /b %errorlevel%
)

echo Updating the MSSQL Driver
go get -u github.com/denisenkom/go-mssqldb
if %errorlevel% neq 0 (
	pause
	exit /b %errorlevel%
)

echo Updating bcrypt
go get -u golang.org/x/crypto/bcrypt
if %errorlevel% neq 0 (
	pause
	exit /b %errorlevel%
)

echo Updating /x/system/windows (dependency for gopsutil)
go get -u golang.org/x/sys/windows
if %errorlevel% neq 0 (
	pause
	exit /b %errorlevel%
)

echo Updating wmi (dependency for gopsutil)
go get -u github.com/StackExchange/wmi
if %errorlevel% neq 0 (
	pause
	exit /b %errorlevel%
)

echo Updating gopsutil
go get -u github.com/Azareal/gopsutil
if %errorlevel% neq 0 (
	pause
	exit /b %errorlevel%
)

echo Updating Gorilla Websockets
go get -u github.com/gorilla/websocket
if %errorlevel% neq 0 (
	pause
	exit /b %errorlevel%
)

echo Updating Sourcemap (dependency for OttoJS)
go get -u gopkg.in/sourcemap.v1
if %errorlevel% neq 0 (
	pause
	exit /b %errorlevel%
)

echo Updating OttoJS
go get -u github.com/robertkrimen/otto
if %errorlevel% neq 0 (
	pause
	exit /b %errorlevel%
)

echo Updating the Riot Search Engine
go get -u github.com/go-ego/riot
if %errorlevel% neq 0 (
	pause
	exit /b %errorlevel%
)

echo Updating the Rez Image Resizer
go get -u github.com/bamiaux/rez
if %errorlevel% neq 0 (
	pause
	exit /b %errorlevel%
)

echo Updating fsnotify
go get -u github.com/fsnotify/fsnotify
if %errorlevel% neq 0 (
	pause
	exit /b %errorlevel%
)

echo The dependencies were successfully updated
pause
