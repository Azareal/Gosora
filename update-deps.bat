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

echo The dependencies were successfully updated
pause