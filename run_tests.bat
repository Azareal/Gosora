@echo off
echo Generating the dynamic code
go generate
if %errorlevel% neq 0 (
	pause
	exit /b %errorlevel%
)

echo Building the router generator
go build ./router_gen
if %errorlevel% neq 0 (
	pause
	exit /b %errorlevel%
)
echo Running the router generator
router_gen.exe

echo Building the query generator
go build ./query_gen
if %errorlevel% neq 0 (
	pause
	exit /b %errorlevel%
)
echo Running the query generator
query_gen.exe

echo Building the executable
go test
if %errorlevel% neq 0 (
	pause
	exit /b %errorlevel%
)
pause