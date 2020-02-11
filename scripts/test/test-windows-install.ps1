$ErrorActionPreference = "Stop"

$env:PATH+=";$env:USERPROFILE\.porter"

& $PSScriptRoot\..\install\install-windows.ps1 -PORTER_PERMALINK canary

& $PSScriptRoot\..\install\install-windows.ps1 -PORTER_PERMALINK v0.23.0-beta.1
if (-Not (porter version | Select-String -Pattern 'v0.23.0-beta.1' -SimpleMatch))
{
    echo "Failed to install a specific version of porter"
    Exit 1
}

& $PSScriptRoot\..\install\install-windows.ps1  -PORTER_PERMALINK latest
