$sourceFolder = ".\clipboard\"
$cppFiles = Get-ChildItem -Path $sourceFolder -Recurse -Filter *.cpp
$buildFolder = "clipboard\libs"
$clipboardLib = "libclipboard.a"
$clientGo = "main\client_main.go"

# go build param
$version = "2.0.8"
$buildDate = (Get-Date).ToString("yyyy-MM-ddTHH:mm:ss")
$platform = "windows"
$isHiddenWin = $true
if ($isHiddenWin) {
    $ldflags = "-X rtk-cross-share/buildConfig.Version=$version -X rtk-cross-share/buildConfig.Platform=$platform -X rtk-cross-share/buildConfig.BuildDate=$buildDate -H=windowsgui"
} else {
    $ldflags = "-X rtk-cross-share/buildConfig.Version=$version -X rtk-cross-share/buildConfig.Platform=$platform -X rtk-cross-share/buildConfig.BuildDate=$buildDate"
}


if (-Not (Test-Path -Path $buildFolder)) {
    mkdir $buildFolder
}

Write-Host "Compile Start"
foreach ($file in $cppFiles) {
    $outfile_o = Join-Path $buildFolder ($file.BaseName + ".o")
    Write-Host "Compiling $($file.FullName) to $outfile_o"
    g++ -c $file.FullName -o $outfile_o
}

Write-Host "Compiling all .o to $buildFolder\$clipboardLib"
ar rcs $buildFolder\$clipboardLib $buildFolder\*.o
rm $buildFolder\*.o
Write-Host "Compiling ($clientGo) to client_windows.exe"
cd "client"
go build `
    -ldflags "$ldflags" `
    -a `
    -o client_windows.exe `
    $clientGo

cd ..
Write-Host "Compile Done"