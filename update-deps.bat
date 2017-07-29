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

echo The dependencies were successfully updated
pause
