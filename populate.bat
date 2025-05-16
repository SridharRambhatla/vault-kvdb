@echo off
setlocal enabledelayedexpansion

REM Generate a random number using %RANDOM%
set /a rand1=%RANDOM%
echo %rand1%

for %%s in (localhost:8080 localhost:8081) do (
    set /a rand2=%RANDOM%
    echo curl "http://%%s/set?key=key-!rand1!&value=value-!rand2!&bucketName=default"
)
