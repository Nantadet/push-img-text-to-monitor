@echo off
chcp 65001 >nul
setlocal

set URL=http://localhost:3001/display
set CHROME="C:\Program Files\Google\Chrome\Application\chrome.exe"

echo Opening Chrome with autoplay bypass...
echo URL: %URL%
echo.

%CHROME% ^
  --autoplay-policy=no-user-gesture-required ^
  --disable-features=AutoplayRequireUserGesture ^
  --app=%URL% ^
  --start-fullscreen ^
  --no-first-run ^
  --no-default-browser-check ^
  --disable-infobars ^
  --disable-session-crashed-bubble ^
  --disable-features=TranslateUI ^
  --disk-cache-dir="%TEMP%\hllc-chrome-cache" ^
  --user-data-dir="%TEMP%\hllc-chrome-profile"

endlocal
