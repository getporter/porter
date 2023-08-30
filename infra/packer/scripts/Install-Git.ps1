Write-Host "Installing Git..."

$ProgressPreference = 'SilentlyContinue' 

wget "https://github.com/git-for-windows/git/releases/download/v2.42.0.windows.1/Git-2.42.0-64-bit.exe" -outfile git.exe
Start-Process 'git.exe' -ArgumentList "/VERYSILENT /NORESTART /NOCANCEL /SP- /CLOSEAPPLICATIONS /RESTARTAPPLICATIONS /o:PathOption=CmdTools /o:BashTerminalOption=ConHost /o:EnableSymlinks=Enabled /COMPONENTS=gitlfs" -Wait



