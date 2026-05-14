@echo off
setlocal

cd /d "%~dp0..\backend"
if not exist ".gocache" mkdir ".gocache"
set "GOCACHE=%CD%\.gocache"

go run .
