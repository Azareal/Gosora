@echo off
rem TODO: Make these deletes a little less noisy
del "template_*.go"
del "tmpl_*.go"
del "gen_*.go"
del ".\tmpl_client\template_*"
del ".\tmpl_client\tmpl_*"
del ".\common\gen_extend.go"
del "gosora.exe"

echo Generating the dynamic code
go generate
if %errorlevel% neq 0 (
	pause
	exit /b %errorlevel%
)

echo Building the router generator
go build -ldflags="-s -w" ./router_gen
if %errorlevel% neq 0 (
	pause
	exit /b %errorlevel%
)
echo Running the router generator
router_gen.exe
if %errorlevel% neq 0 (
	pause
	exit /b %errorlevel%
)

echo Building the hook stub generator
go build -ldflags="-s -w" "./cmd/hook_stub_gen"
if %errorlevel% neq 0 (
	pause
	exit /b %errorlevel%
)
echo Running the hook stub generator
hook_stub_gen.exe
if %errorlevel% neq 0 (
	pause
	exit /b %errorlevel%
)

echo Building the hook generator
go build -tags hookgen -ldflags="-s -w" "./cmd/hook_gen"
if %errorlevel% neq 0 (
	pause
	exit /b %errorlevel%
)
echo Running the hook generator
hook_gen.exe
if %errorlevel% neq 0 (
	pause
	exit /b %errorlevel%
)

echo Generating the JSON handlers
easyjson -pkg common

echo Building the query generator
go build -ldflags="-s -w" "./cmd/query_gen"
if %errorlevel% neq 0 (
	pause
	exit /b %errorlevel%
)
echo Running the query generator
query_gen.exe
if %errorlevel% neq 0 (
	pause
	exit /b %errorlevel%
)

echo Building the executable
go test
if %errorlevel% neq 0 (
	pause
	exit /b %errorlevel%
)
pause