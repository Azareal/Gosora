@echo off
echo Updating the MySQL Driver
go get -u github.com/go-sql-driver/mysql
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

echo Updating wmi (dependency for gopsutil)
go get -u github.com/StackExchange/wmi
if %errorlevel% neq 0 (
	pause
	exit /b %errorlevel%
)

echo Updating gopsutil
go get -u github.com/shirou/gopsutil
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

echo The dependencies were successfully updated
pause