@echo off
cd /d "%~dp0"
launcher -e age1 --gameConfig resources/config.aoe1.toml
if %ERRORLEVEL%==0 (
    echo Program finished successfully, closing in 10 seconds...
    timeout /t 10
) else (
    echo Program finished with errors...
    echo You may try running "cleanup-aoe1.bat" as regular user.
    pause
)
