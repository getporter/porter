Write-Host "Installing Docker Desktop..."

$ProgressPreference = 'SilentlyContinue' 
wget "https://desktop.docker.com/win/main/amd64/Docker%20Desktop%20Installer.exe" -outfile "DockerDesktopInstaller.exe"
Start-Process 'DockerDesktopInstaller.exe' -ArgumentList "install --quiet --accept-licence --no-windows-containers --backend=wsl-2" -Wait