echo Building the templates
gosora.exe -build-templates

echo Rebuilding the executable
go build -o gosora.exe
if %errorlevel% neq 0 (
	pause
	exit /b %errorlevel%
)

pause