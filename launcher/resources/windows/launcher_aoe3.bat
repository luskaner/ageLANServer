@echo off
cd /d "%~dp0"
launcher -e age3 --gameConfig resources/config.aoe3.toml
if %ERRORLEVEL%==0 (
    echo Program finished successfully, closing in 10 seconds...
    timeout /t 10
) else (
    echo Program finished with errors...
    echo You may try running "cleanup-aoe3.bat" as regular user.
    pause
)
