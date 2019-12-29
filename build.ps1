param([string]$function="")

# Only run if an arg was provided
if ($function -eq "")
{
    Write-Host "`n[ERROR] Please provide the function name`n" -ForegroundColor Red;
} 
else 
{
    Write-Host "`n[INFO] Attempting to create lambda package '$function'" -ForegroundColor Blue;

    # Environment
    $env:GOOS = "linux"
    $env:GOARCH = "amd64"
    $env:CGO_ENABLED = "0"

    # Vars
    $binary = "./build/$function"
    $zip = "./build/$function.zip"
    $entrypoint = "./internal/endpoints/$function.go"

    # Build go binary
    go build -o $binary $entrypoint; 

    if ($? -ne $true) {
        Write-Host "[ERROR] Error while building binary '$binary'`n" -ForegroundColor Red;
        exit
    }
    
    # Build aws lambda zip
    build-lambda-zip --output $zip $binary;

    if ($? -ne $true) {
        Write-Host "[ERROR] Error while building AWS lambda package $zip'`n" -ForegroundColor Red;
        exit
    }

    Write-Host "`n[INFO] Successfully created AWS lambda package '$zip'`n" -ForegroundColor Green;
}

