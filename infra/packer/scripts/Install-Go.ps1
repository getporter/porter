Write-Host "Installing Go..."

$ProgressPreference = 'SilentlyContinue' 
wget https://go.dev/dl/go1.21.0.windows-amd64.msi -OutFile go.msi
Start-Process 'msiexec.exe' -ArgumentList "/i go.msi /qn /norestart" -Wait