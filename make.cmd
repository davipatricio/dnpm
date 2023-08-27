@echo off

REM Define the BUILD_DIR variable
set BUILD_DIR=dist

REM Define targets and their respective commands
:build
    echo Building dnpm...
    go build -o %BUILD_DIR%\dnpm.exe cmd\dnpm.go
    echo Build complete. Binary is located in %BUILD_DIR%\
    goto :eof

:clean
    echo Cleaning up...
    rmdir /s /q %BUILD_DIR%
    echo Cleanup complete.
    goto :eof

:default
    echo Please specify a valid target: build or clean.
    goto :eof

REM Entry point
if "%1" == "build" goto build
if "%1" == "clean" goto clean
goto default
