@echo off
chcp 65001 >nul
setlocal

set URL=http://localhost:3001/display
set EDGE="C:\Program Files (x86)\Microsoft\Edge\Application\msedge.exe"

echo Opening Edge with autoplay bypass...
echo URL: %URL%
echo.

%EDGE% ^
  --autoplay-policy=no-user-gesture-required ^
  --disable-features=AutoplayRequireUserGesture ^
  --app=%URL% ^
  --start-fullscreen ^
  --no-first-run ^
  --no-default-browser-check ^
  --disable-infobars ^
  --disable-session-crashed-bubble ^
  --disable-features=TranslateUI ^
  --disk-cache-dir="%TEMP%\hllc-edge-cache" ^
  --user-data-dir="%TEMP%\hllc-edge-profile"

endlocal
