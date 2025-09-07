@echo off
color 0A

title Building Launcher (Golang)
ping localhost -n 1 >nul
echo Building launcher
echo.
ping localhost -n 3 >nul
echo Launcher is done! :))
go build

