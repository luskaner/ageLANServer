@echo off
cd /d "%~dp0"
launcher -e age2 --gameConfig resources/config.aoe2.toml
if %ERRORLEVEL%==0 (
    echo Program finished successfully, closing in 10 seconds...
    timeout /t 10
) else (
    echo Program finished with errors...
    pause
)
