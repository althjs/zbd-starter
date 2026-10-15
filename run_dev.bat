@echo off
chcp 65001 >nul
title zbd-starter
rem 개발 모드(Windows) — 더블클릭. 실제 일은 run_dev.ps1(PowerShell)이 한다.
powershell -NoProfile -ExecutionPolicy Bypass -File "%~dp0run_dev.ps1" %*
if errorlevel 1 pause
