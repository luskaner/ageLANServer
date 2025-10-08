@echo off
cd /d "%~dp0"
battle-server-manager start -e athens
if %ERRORLEVEL%==0 (
    echo Program finished successfully, closing in 10 seconds...
    timeout /t 10
) else (
    echo Program finished with errors...
    pause
)
