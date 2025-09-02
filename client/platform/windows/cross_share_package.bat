@echo off
chcp 65001 > nul

:: ================= CONFIGURATION =================
set NSIS_SCRIPT="%~dp0nsis_setup.nsi"
set PRODUCT_VERSION=2.1.17.0
set COMPANY_NAME=Realtek
set PRODUCT_NAME=CrossShare
set MAIN_EXE=windows_clipboard.exe
set SERVICE_EXE=cross_share_serv.exe
set CLIENT_FOLDER=CrossShareWindowsClient
set APP_ICON=application.ico
set OUTPUT_EXE="%~dp0CrossShareSetup.exe"
set GO_SERVER_DLL=%CLIENT_FOLDER%\client_windows.dll
:: =================================================

echo ===================================
echo Building %PRODUCT_NAME% Installer...
echo Version: %PRODUCT_VERSION%
echo Company: %COMPANY_NAME%
echo Installer: %OUTPUT_EXE%
echo ===================================

del /q %OUTPUT_EXE% 2>nul

if not exist %GO_SERVER_DLL% (
	echo %GO_SERVER_DLL% file does not exist!
	echo Press any key to exit...
	pause > nul
	exit /b 1
)

:: Extract filename from full path for NSIS
for %%I in (%OUTPUT_EXE%) do set OUTPUT_FILENAME=%%~nxI

:: Copy application icon
echo Copying application icon...
xcopy /F /Y .\source_code\windows_clipboard\src\resource\%APP_ICON% .\ > nul
xcopy /F /Y .\source_code\windows_clipboard\src\resource\%APP_ICON% .\%CLIENT_FOLDER% > nul
if %errorlevel% neq 0 (
    echo ERROR: Failed to copy icon file
    exit /b 1
)

:: Pass environment variables to NSIS
set __MAKENSISFLAGS= ^
    /DPRODUCT_VERSION="%PRODUCT_VERSION%" ^
    /DCOMPANY_NAME="%COMPANY_NAME%" ^
    /DPRODUCT_NAME="%PRODUCT_NAME%" ^
    /DMAIN_EXE="%MAIN_EXE%" ^
    /DSERVICE_EXE="%SERVICE_EXE%" ^
    /DCLIENT_FOLDER="%CLIENT_FOLDER%" ^
    /DAPP_ICON="%APP_ICON%" ^
    /DOUTPUT_FILENAME="%OUTPUT_FILENAME%" ^
    %__MAKENSISFLAGS%

:: Run NSIS compiler with detailed output
echo.
echo Starting NSIS compilation...
echo =============================
makensis /V4 %__MAKENSISFLAGS% %NSIS_SCRIPT% -X"OutFile %OUTPUT_EXE%"
echo =============================
echo.

:: Cleanup temporary files
echo Cleaning temporary files...
del /Q .\%CLIENT_FOLDER%\%APP_ICON% > nul 2>&1
del /Q .\%APP_ICON% > nul 2>&1

:: Handle build result
if %errorlevel% equ 0 (
    echo -----------------------------------
    echo BUILD SUCCESSFUL!
    echo Installer created: %OUTPUT_EXE%
    echo -----------------------------------
) else (
    echo -----------------------------------
    echo BUILD FAILED! Check errors above
    echo -----------------------------------
    exit /b 1
)

echo Press any key to exit...
pause > nul