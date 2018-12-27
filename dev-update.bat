@echo off

echo Updating the dependencies
go get
if %errorlevel% neq 0 (
	pause
	exit /b %errorlevel%
)

echo Updating Gosora
git stash
if %errorlevel% neq 0 (
	pause
	exit /b %errorlevel%
)
git pull origin master
if %errorlevel% neq 0 (
	pause
	exit /b %errorlevel%
)
git stash apply
if %errorlevel% neq 0 (
	pause
	exit /b %errorlevel%
)

echo Patching Gosora
go generate
go build ./patcher
patcher.exe