# Ollamacron Windows Installation Script
# Automated installation for Windows 10/11

#Requires -RunAsAdministrator

param(
    [string]$Version = "latest",
    [string]$InstallDir = "$env:ProgramFiles\Ollamacron",
    [string]$ConfigDir = "$env:ProgramData\Ollamacron",
    [string]$DataDir = "$env:ProgramData\Ollamacron\data",
    [string]$LogDir = "$env:ProgramData\Ollamacron\logs",
    [switch]$Uninstall,
    [switch]$Help
)

# Configuration
$GitHubRepo = "ollama-distributed/ollamacron"
$GitHubApiUrl = "https://api.github.com/repos/$GitHubRepo"
$ServiceName = "Ollamacron"
$ServiceDisplayName = "Ollamacron Distributed AI Inference Service"

# Color output functions
function Write-ColorOutput {
    param(
        [string]$Message,
        [string]$Color = "White"
    )
    
    $timestamp = Get-Date -Format "yyyy-MM-dd HH:mm:ss"
    Write-Host "[$timestamp] $Message" -ForegroundColor $Color
}

function Write-Success { param([string]$Message) Write-ColorOutput $Message "Green" }
function Write-Error { param([string]$Message) Write-ColorOutput $Message "Red" }
function Write-Warning { param([string]$Message) Write-ColorOutput $Message "Yellow" }
function Write-Info { param([string]$Message) Write-ColorOutput $Message "Cyan" }

# Check if running as administrator
function Test-Administrator {
    $currentUser = [Security.Principal.WindowsPrincipal][Security.Principal.WindowsIdentity]::GetCurrent()
    return $currentUser.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
}

# Detect system architecture
function Get-SystemArchitecture {
    $arch = $env:PROCESSOR_ARCHITECTURE
    switch ($arch) {
        "AMD64" { return "amd64" }
        "ARM64" { return "arm64" }
        default { 
            Write-Error "Unsupported architecture: $arch"
            exit 1
        }
    }
}

# Check Windows version
function Test-WindowsVersion {
    $version = [System.Environment]::OSVersion.Version
    $buildNumber = (Get-ItemProperty "HKLM:\SOFTWARE\Microsoft\Windows NT\CurrentVersion").CurrentBuild
    
    if ($version.Major -lt 10) {
        Write-Error "Windows 10 or later is required"
        exit 1
    }
    
    Write-Info "Windows version: $($version.Major).$($version.Minor) (Build $buildNumber)"
}

# Install Chocolatey if not present
function Install-Chocolatey {
    if (-not (Get-Command choco -ErrorAction SilentlyContinue)) {
        Write-Success "Installing Chocolatey..."
        Set-ExecutionPolicy Bypass -Scope Process -Force
        [System.Net.ServicePointManager]::SecurityProtocol = [System.Net.ServicePointManager]::SecurityProtocol -bor 3072
        iex ((New-Object System.Net.WebClient).DownloadString('https://community.chocolatey.org/install.ps1'))
        
        # Refresh environment variables
        $env:PATH = [System.Environment]::GetEnvironmentVariable("PATH", "Machine") + ";" + [System.Environment]::GetEnvironmentVariable("PATH", "User")
        
        Write-Success "Chocolatey installed successfully"
    } else {
        Write-Success "Chocolatey already installed"
    }
}

# Install dependencies
function Install-Dependencies {
    Write-Success "Installing dependencies..."
    
    # Install required packages
    $packages = @("curl", "wget", "jq", "golang", "docker-desktop")
    
    foreach ($package in $packages) {
        Write-Info "Installing $package..."
        try {
            choco install $package -y --limit-output
        } catch {
            Write-Warning "Failed to install $package, continuing..."
        }
    }
    
    # Install Windows Service Wrapper (WinSW)
    if (-not (Test-Path "$InstallDir\winsw.exe")) {
        Write-Info "Installing Windows Service Wrapper..."
        $winswUrl = "https://github.com/winsw/winsw/releases/latest/download/WinSW-net4.exe"
        New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null
        Invoke-WebRequest -Uri $winswUrl -OutFile "$InstallDir\winsw.exe"
    }
    
    Write-Success "Dependencies installed successfully"
}

# Download Ollamacron binary
function Install-OllamacronBinary {
    Write-Success "Downloading Ollamacron..."
    
    $arch = Get-SystemArchitecture
    
    # Get latest release info
    if ($Version -eq "latest") {
        $releaseInfo = Invoke-RestMethod -Uri "$GitHubApiUrl/releases/latest"
        $downloadUrl = $releaseInfo.assets | Where-Object { $_.name -like "*windows-$arch*" } | Select-Object -First 1 | Select-Object -ExpandProperty browser_download_url
    } else {
        $downloadUrl = "https://github.com/$GitHubRepo/releases/download/$Version/ollamacron-windows-$arch.exe"
    }
    
    if (-not $downloadUrl) {
        Write-Error "Could not find download URL for Ollamacron"
        exit 1
    }
    
    # Download binary
    $tempFile = [System.IO.Path]::GetTempFileName() + ".exe"
    try {
        Invoke-WebRequest -Uri $downloadUrl -OutFile $tempFile
        
        # Create installation directory
        New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null
        
        # Install binary
        Move-Item -Path $tempFile -Destination "$InstallDir\ollamacron.exe" -Force
        
        Write-Success "Ollamacron binary installed to $InstallDir\ollamacron.exe"
    } catch {
        Write-Error "Failed to download Ollamacron: $_"
        exit 1
    }
}

# Create directory structure
function New-DirectoryStructure {
    Write-Success "Creating directory structure..."
    
    # Create directories
    New-Item -ItemType Directory -Force -Path $ConfigDir | Out-Null
    New-Item -ItemType Directory -Force -Path $DataDir | Out-Null
    New-Item -ItemType Directory -Force -Path $LogDir | Out-Null
    
    # Set permissions
    $acl = Get-Acl $ConfigDir
    $accessRule = New-Object System.Security.AccessControl.FileSystemAccessRule("NETWORK SERVICE", "FullControl", "ContainerInherit,ObjectInherit", "None", "Allow")
    $acl.SetAccessRule($accessRule)
    Set-Acl -Path $ConfigDir -AclObject $acl
    Set-Acl -Path $DataDir -AclObject $acl
    Set-Acl -Path $LogDir -AclObject $acl
    
    Write-Success "Directory structure created"
}

# Create default configuration
function New-DefaultConfig {
    Write-Success "Creating default configuration..."
    
    $configContent = @"
# Ollamacron Configuration
server:
  bind: "127.0.0.1:8080"
  tls:
    enabled: false
    cert_file: ""
    key_file: ""

p2p:
  enabled: true
  listen_addr: "/ip4/0.0.0.0/tcp/9000"
  bootstrap_peers: []
  discovery:
    enabled: true
    rendezvous: "ollamacron-v1"

models:
  cache_dir: "$($DataDir -replace '\\', '/')/models"
  auto_pull: true
  sync_interval: "5m"

logging:
  level: "info"
  format: "json"
  output: "$($LogDir -replace '\\', '/')/ollamacron.log"

metrics:
  enabled: true
  bind: "127.0.0.1:9090"
  path: "/metrics"

health:
  enabled: true
  bind: "127.0.0.1:8081"
  path: "/health"
"@
    
    $configContent | Out-File -FilePath "$ConfigDir\config.yaml" -Encoding UTF8
    Write-Success "Default configuration created"
}

# Create Windows service
function New-WindowsService {
    Write-Success "Creating Windows service..."
    
    # Create service configuration
    $serviceConfig = @"
<service>
    <id>$ServiceName</id>
    <name>$ServiceDisplayName</name>
    <description>Distributed AI inference service using P2P networking</description>
    <executable>$InstallDir\ollamacron.exe</executable>
    <arguments>server --config "$ConfigDir\config.yaml"</arguments>
    <workingdirectory>$InstallDir</workingdirectory>
    <logmode>rotate</logmode>
    <logpath>$LogDir</logpath>
    <env name="OLLAMACRON_CONFIG" value="$ConfigDir\config.yaml"/>
    <env name="OLLAMACRON_DATA_DIR" value="$DataDir"/>
    <env name="OLLAMACRON_LOG_DIR" value="$LogDir"/>
    <startmode>Automatic</startmode>
    <delayedAutoStart>true</delayedAutoStart>
    <serviceaccount>
        <domain>NT AUTHORITY</domain>
        <user>NETWORK SERVICE</user>
    </serviceaccount>
    <onfailure action="restart" delay="10 sec"/>
    <onfailure action="restart" delay="20 sec"/>
    <onfailure action="none"/>
    <resetfailure>1 hour</resetfailure>
</service>
"@
    
    $serviceConfig | Out-File -FilePath "$InstallDir\$ServiceName.xml" -Encoding UTF8
    
    # Install service
    try {
        & "$InstallDir\winsw.exe" install "$InstallDir\$ServiceName.xml"
        Write-Success "Windows service created"
    } catch {
        Write-Error "Failed to create Windows service: $_"
        exit 1
    }
}

# Install auto-update mechanism
function Install-AutoUpdate {
    Write-Success "Installing auto-update mechanism..."
    
    # Create update script
    $updateScript = @"
# Ollamacron Auto-Update Script
`$currentVersion = try { & "$InstallDir\ollamacron.exe" version 2>`$null | Select-String -Pattern "v\d+\.\d+\.\d+" | ForEach-Object { `$_.Matches.Value } } catch { "unknown" }
`$latestVersion = (Invoke-RestMethod -Uri "$GitHubApiUrl/releases/latest").tag_name

if (`$currentVersion -ne `$latestVersion) {
    Write-Host "Update available: `$currentVersion -> `$latestVersion"
    
    # Download new version
    `$arch = "$env:PROCESSOR_ARCHITECTURE"
    if (`$arch -eq "AMD64") { `$arch = "amd64" }
    elseif (`$arch -eq "ARM64") { `$arch = "arm64" }
    
    `$downloadUrl = "https://github.com/$GitHubRepo/releases/download/`$latestVersion/ollamacron-windows-`$arch.exe"
    `$tempFile = [System.IO.Path]::GetTempFileName() + ".exe"
    
    try {
        Invoke-WebRequest -Uri `$downloadUrl -OutFile `$tempFile
        
        # Stop service
        Stop-Service -Name "$ServiceName" -Force
        
        # Replace binary
        Move-Item -Path `$tempFile -Destination "$InstallDir\ollamacron.exe" -Force
        
        # Start service
        Start-Service -Name "$ServiceName"
        
        Write-Host "Updated to `$latestVersion"
    } catch {
        Write-Error "Update failed: `$_"
    }
} else {
    Write-Host "Already up to date: `$currentVersion"
}
"@
    
    $updateScript | Out-File -FilePath "$InstallDir\update.ps1" -Encoding UTF8
    
    # Create scheduled task for auto-update
    $action = New-ScheduledTaskAction -Execute "powershell.exe" -Argument "-ExecutionPolicy Bypass -File `"$InstallDir\update.ps1`""
    $trigger = New-ScheduledTaskTrigger -Daily -At "02:00"
    $settings = New-ScheduledTaskSettingsSet -ExecutionTimeLimit (New-TimeSpan -Hours 1)
    $principal = New-ScheduledTaskPrincipal -UserId "SYSTEM" -LogonType ServiceAccount
    
    Register-ScheduledTask -TaskName "OllamacronAutoUpdate" -Action $action -Trigger $trigger -Settings $settings -Principal $principal -Description "Auto-update Ollamacron daily"
    
    Write-Success "Auto-update mechanism installed"
}

# Configure Windows Firewall
function Set-FirewallRules {
    Write-Success "Configuring Windows Firewall..."
    
    $ports = @(8080, 9000, 9090, 8081)
    $descriptions = @("API Server", "P2P Networking", "Metrics", "Health Checks")
    
    for ($i = 0; $i -lt $ports.Length; $i++) {
        $port = $ports[$i]
        $description = $descriptions[$i]
        
        try {
            New-NetFirewallRule -DisplayName "Ollamacron $description" -Direction Inbound -Protocol TCP -LocalPort $port -Action Allow -Profile Any
            Write-Info "Added firewall rule for port $port ($description)"
        } catch {
            Write-Warning "Failed to add firewall rule for port $port"
        }
    }
    
    Write-Success "Firewall rules configured"
}

# Verify installation
function Test-Installation {
    Write-Success "Verifying installation..."
    
    # Check binary
    if (-not (Test-Path "$InstallDir\ollamacron.exe")) {
        Write-Error "Ollamacron binary not found"
        exit 1
    }
    
    # Check version
    try {
        $version = & "$InstallDir\ollamacron.exe" version 2>$null
        Write-Info "Installed version: $version"
    } catch {
        Write-Info "Version check failed, but binary exists"
    }
    
    # Check service
    $service = Get-Service -Name $ServiceName -ErrorAction SilentlyContinue
    if (-not $service) {
        Write-Error "Ollamacron service not found"
        exit 1
    }
    
    Write-Success "Installation verification completed"
}

# Uninstall function
function Uninstall-Ollamacron {
    Write-Success "Uninstalling Ollamacron..."
    
    # Stop and remove service
    try {
        Stop-Service -Name $ServiceName -Force -ErrorAction SilentlyContinue
        & "$InstallDir\winsw.exe" uninstall "$InstallDir\$ServiceName.xml"
    } catch {
        Write-Warning "Failed to remove service"
    }
    
    # Remove scheduled task
    try {
        Unregister-ScheduledTask -TaskName "OllamacronAutoUpdate" -Confirm:$false
    } catch {
        Write-Warning "Failed to remove scheduled task"
    }
    
    # Remove firewall rules
    try {
        Remove-NetFirewallRule -DisplayName "Ollamacron*" -ErrorAction SilentlyContinue
    } catch {
        Write-Warning "Failed to remove firewall rules"
    }
    
    # Remove directories
    try {
        Remove-Item -Path $InstallDir -Recurse -Force -ErrorAction SilentlyContinue
        Remove-Item -Path $ConfigDir -Recurse -Force -ErrorAction SilentlyContinue
        Remove-Item -Path $DataDir -Recurse -Force -ErrorAction SilentlyContinue
        Remove-Item -Path $LogDir -Recurse -Force -ErrorAction SilentlyContinue
    } catch {
        Write-Warning "Failed to remove some directories"
    }
    
    Write-Success "Ollamacron uninstalled"
}

# Show help
function Show-Help {
    Write-Host @"
Ollamacron Windows Installer v1.0.0

Usage: ./install.ps1 [options]

Options:
  -Version <version>     Version to install (default: latest)
  -InstallDir <path>     Installation directory (default: $env:ProgramFiles\Ollamacron)
  -ConfigDir <path>      Configuration directory (default: $env:ProgramData\Ollamacron)
  -DataDir <path>        Data directory (default: $env:ProgramData\Ollamacron\data)
  -LogDir <path>         Log directory (default: $env:ProgramData\Ollamacron\logs)
  -Uninstall             Uninstall Ollamacron
  -Help                  Show this help message

Examples:
  ./install.ps1
  ./install.ps1 -Version v1.0.0
  ./install.ps1 -InstallDir "C:\MyApps\Ollamacron"
  ./install.ps1 -Uninstall
"@
}

# Main installation function
function Install-Ollamacron {
    Write-Success "Starting Ollamacron installation..."
    
    # Check prerequisites
    if (-not (Test-Administrator)) {
        Write-Error "This script must be run as Administrator"
        exit 1
    }
    
    Test-WindowsVersion
    $arch = Get-SystemArchitecture
    Write-Info "Detected architecture: $arch"
    
    # Installation steps
    Install-Chocolatey
    Install-Dependencies
    New-DirectoryStructure
    Install-OllamacronBinary
    New-DefaultConfig
    New-WindowsService
    Install-AutoUpdate
    Set-FirewallRules
    
    # Start service
    Start-Service -Name $ServiceName
    
    # Verify installation
    Test-Installation
    
    Write-Success "Ollamacron installation completed successfully!"
    Write-Host ""
    Write-Host "🎉 Ollamacron is now installed and running!" -ForegroundColor Green
    Write-Host ""
    Write-Host "Next steps:" -ForegroundColor Cyan
    Write-Host "  1. Check service status: Get-Service -Name $ServiceName"
    Write-Host "  2. View logs: Get-Content $LogDir\ollamacron.log -Tail 50"
    Write-Host "  3. Test API: curl http://localhost:8080/health"
    Write-Host "  4. Edit config: notepad $ConfigDir\config.yaml"
    Write-Host "  5. View metrics: curl http://localhost:9090/metrics"
    Write-Host ""
    Write-Host "Service management:" -ForegroundColor Cyan
    Write-Host "  Start:   Start-Service -Name $ServiceName"
    Write-Host "  Stop:    Stop-Service -Name $ServiceName"
    Write-Host "  Restart: Restart-Service -Name $ServiceName"
    Write-Host ""
    Write-Host "Documentation:" -ForegroundColor Cyan
    Write-Host "  https://github.com/ollama-distributed/ollamacron/docs"
    Write-Host ""
}

# Handle script arguments
if ($Help) {
    Show-Help
    exit 0
}

if ($Uninstall) {
    Uninstall-Ollamacron
    exit 0
}

# Run main installation
Install-Ollamacron