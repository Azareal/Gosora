@echo off

echo Updating the dependencies
go get
if %errorlevel% neq 0 (
	pause
	exit /b %errorlevel%
)

go get -u github.com/mailru/easyjson/...

echo The dependencies were successfully updated
pause
