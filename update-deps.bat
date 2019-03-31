@echo off

echo Updating the dependencies
go get -u github.com/mailru/easyjson/...
if %errorlevel% neq 0 (
	pause
	exit /b %errorlevel%
)

go get
if %errorlevel% neq 0 (
	pause
	exit /b %errorlevel%
)

echo The dependencies were successfully updated
pause
