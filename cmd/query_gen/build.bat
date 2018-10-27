@echo off
echo Building the query generator
go build
if %errorlevel% neq 0 (
	pause
	exit /b %errorlevel%
)
echo The query generator was successfully built
pause