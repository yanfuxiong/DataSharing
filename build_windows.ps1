$sourceFolder = ".\clipboard\"
$cppFiles = Get-ChildItem -Path $sourceFolder -Recurse -Filter *.cpp
$buildFolder = "clipboard\libs"
$clipboardLib = "libclipboard.a"
$clientGo = "main\client_main.go"

# go build param
$version = "2.1.16"
$buildDate = (Get-Date).ToString("yyyy-MM-ddTHH:mm:ss")
$isBuildRelease = $true
$isBuildDebug = $true

$ldflags_release = "-X rtk-cross-share/client/buildConfig.Version=$version -X rtk-cross-share/client/buildConfig.BuildDate=$buildDate -X rtk-cross-share/client/buildConfig.Debug=0 -H=windowsgui"
$ldflags_debug = "-X rtk-cross-share/client/buildConfig.Version=$version -X rtk-cross-share/client/buildConfig.BuildDate=$buildDate -X rtk-cross-share/client/buildConfig.Debug=1"

if (-Not (Test-Path -Path $buildFolder)) {
    mkdir $buildFolder
}

Write-Host "Compile Start"
foreach ($file in $cppFiles) {
    $outfile_o = Join-Path $buildFolder ($file.BaseName + ".o")
    Write-Host "Compiling $($file.FullName) to $outfile_o"
    g++ -I "clipboard/pipeserver-asio" -I "clipboard/asio-1.30.2" -c $file.FullName -o $outfile_o
}

Write-Host "Compiling all .o to $buildFolder\$clipboardLib"
ar rcs $buildFolder\$clipboardLib $buildFolder\*.o
rm $buildFolder\*.o
Write-Host "Compiling ($clientGo) to client_windows.exe"
cd "client"
if ($isBuildRelease) {
    go build `
        -ldflags "$ldflags_release" `
        -a `
        -o client_windows_release.exe `
        $clientGo
}

if ($isBuildDebug) {
    go build `
        -ldflags "$ldflags_debug" `
        -a `
        -o client_windows_debug.exe `
        $clientGo
}

cd ..
Write-Host "Compile Done"
