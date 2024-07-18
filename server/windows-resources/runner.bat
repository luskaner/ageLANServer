@echo off
setlocal
set binary=server
set binaryExe=%binary%.exe
if "%PROCESSOR_ARCHITECTURE%"=="x86" (set ARCH=x86-32) else (set ARCH=x86-64)
if "%PROCESSOR_ARCHITEW6432%"=="ARM64" (set ARCH=arm64)

if not defined ARCH (set ARCH=arm32)

copy /Y "%~dp0bin\%binary%\%binary%_%ARCH%.exe" "%~dp0%binaryExe%" > nul
cd "%~dp0"
.\%binaryExe% %*
del .\%binaryExe%
exit /B %errorlevel%