@echo off
chcp 65001 >nul
setlocal EnableDelayedExpansion

set URL=http://localhost:3001/display

set CHROME_PATH=C:\Program Files\Google\Chrome\Application\chrome.exe
set CHROME_PATH2=C:\Program Files (x86)\Google\Chrome\Application\chrome.exe
set EDGE_PATH=C:\Program Files\Microsoft\Edge\Application\msedge.exe
set EDGE_PATH2=C:\Program Files (x86)\Microsoft\Edge\Application\msedge.exe

if exist "%CHROME_PATH%" (
    set BROWSER="%CHROME_PATH%"
    echo Found Chrome
) else if exist "%CHROME_PATH2%" (
    set BROWSER="%CHROME_PATH2%"
    echo Found Chrome (x86)
) else if exist "%EDGE_PATH%" (
    set BROWSER="%EDGE_PATH%"
    echo Found Edge
) else if exist "%EDGE_PATH2%" (
    set BROWSER="%EDGE_PATH2%"
    echo Found Edge (x86)
)

if not defined BROWSER (
    echo ERROR: Could not find Chrome or Edge.
    echo Please install Chrome or Edge, or edit this script with the correct browser path.
    pause
    exit /b 1
)

echo.
echo Opening display with autoplay bypass...
echo URL: %URL%
echo.

%BROWSER% ^
  --autoplay-policy=no-user-gesture-required ^
  --disable-features=AutoplayRequireUserGesture ^
  --app=%URL% ^
  --start-fullscreen ^
  --no-first-run ^
  --no-default-browser-check ^
  --disable-infobars ^
  --disable-session-crashed-bubble ^
  --disable-features=TranslateUI ^
  --disk-cache-dir="%TEMP%\hllc-browser-cache" ^
  --user-data-dir="%TEMP%\hllc-browser-profile"

endlocal
