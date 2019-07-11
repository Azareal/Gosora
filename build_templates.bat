echo Building the templates
gosora.exe -build-templates

echo Rebuilding the executable
go build -ldflags="-s -w" -o gosora.exe
if %errorlevel% neq 0 (
	pause
	exit /b %errorlevel%
)

pause