param([string]$function="")

if ($function -eq "")
{
    Write-Host "`n[ERROR] Please provide the function name`n" -ForegroundColor Red;
} 
else 
{
    Write-Host "`n[INFO] Attempting to create lambda package '$function'" -ForegroundColor Blue;

    $env:GOOS = "linux"; 
    $binary = "$function.bin"
    $zip = "$function.zip"
    $entrypoint = "./endpoints/$function.go"


    go build -o $binary $entrypoint; 

    if ($? -ne $true) {
        Write-Host "[ERROR] Error while building binary '$binary'`n" -ForegroundColor Red;
        exit
    }
    
    build-lambda-zip.exe --output $zip $binary;

    if ($? -ne $true) {
        Write-Host "[ERROR] Error while building AWS lambda package $zip'`n" -ForegroundColor Red;
        exit
    }

    Write-Host "`n[INFO] Successfully created AWS lambda package '$zip'`n" -ForegroundColor Green;
}

