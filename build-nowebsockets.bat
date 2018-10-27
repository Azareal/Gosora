@echo off
rem TODO: Make these deletes a little less noisy
del "template_*.go"
del "gen_*.go"
del "tmpl_client/template_*.go"
del "gosora.exe"

echo Generating the dynamic code
go generate
if %errorlevel% neq 0 (
	pause
	exit /b %errorlevel%
)

echo Building the executable
go build -o gosora.exe -tags no_ws
if %errorlevel% neq 0 (
	pause
	exit /b %errorlevel%
)

echo Building the installer
go build "./cmd/install"
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

echo Building the query generator
go build "./cmd/query_gen"
if %errorlevel% neq 0 (
	pause
	exit /b %errorlevel%
)
echo Gosora was successfully built
pause