# scripts/install.ps1 — OpenFiltr one-line installer for Windows
#
# Usage (run in an elevated PowerShell session):
#   irm https://raw.githubusercontent.com/openfiltr/openfiltr/main/scripts/install.ps1 | iex
#   .\install.ps1 -Version v1.0.0 -DryRun
#
# For Linux / macOS, use scripts/install.sh instead.

#Requires -Version 5.1
[CmdletBinding()]
param(
    [string]$Version = $(if ($env:OPENFILTR_VERSION) { $env:OPENFILTR_VERSION } else { "latest" }),
    [switch]$DryRun,
    [switch]$NoRoot
)

$ErrorActionPreference = "Stop"
$ProgressPreference    = "SilentlyContinue"   # speeds up Invoke-WebRequest

$Repo       = "openfiltr/openfiltr"
$BinaryName = "openfiltr"

# ── Colour helpers ─────────────────────────────────────────────────────────────
function Write-Info($msg)    { Write-Host "  ->  $msg" -ForegroundColor Cyan }
function Write-Success($msg) { Write-Host "  v  $msg"  -ForegroundColor Green }
function Write-Warn($msg)    { Write-Host "  !  $msg"  -ForegroundColor Yellow }
function Write-Fatal($msg)   { Write-Host "  x  $msg"  -ForegroundColor Red; exit 1 }

if ($DryRun) { Write-Warn "DRY RUN -- no changes will be made" }

# ── Architecture detection ─────────────────────────────────────────────────────
$Arch = if ($env:PROCESSOR_ARCHITECTURE -eq "ARM64") { "arm64" } else { "amd64" }
$OS   = "windows"

# ── Paths ──────────────────────────────────────────────────────────────────────
if ($NoRoot) {
    $InstallDir = Join-Path $env:LOCALAPPDATA "OpenFiltr\bin"
    $ConfigDir  = Join-Path $env:APPDATA       "openfiltr"
    $DataDir    = Join-Path $env:LOCALAPPDATA  "openfiltr\data"
} else {
    $InstallDir = Join-Path $env:ProgramFiles  "OpenFiltr"
    $ConfigDir  = Join-Path $env:ProgramData   "openfiltr"
    $DataDir    = Join-Path $env:ProgramData   "openfiltr\data"
}

# ── Resolve latest version ─────────────────────────────────────────────────────
if ($Version -eq "latest") {
    Write-Info "Resolving latest release..."
    try {
        $Release = Invoke-RestMethod "https://api.github.com/repos/$Repo/releases/latest"
        $Version = $Release.tag_name
    } catch {
        Write-Fatal "Could not determine latest release version: $_"
    }
    if (-not $Version) { Write-Fatal "Could not determine latest release version" }
}

Write-Success "Installing OpenFiltr $Version (${OS}/${Arch})"

$ArchiveFilename = "${BinaryName}-${OS}-${Arch}.zip"
$BinaryFilename  = "${BinaryName}-${OS}-${Arch}.exe"
$DownloadUrl     = "https://github.com/$Repo/releases/download/$Version/$ArchiveFilename"
$ChecksumUrl     = "https://github.com/$Repo/releases/download/$Version/checksums.txt"
$TmpDir          = Join-Path ([System.IO.Path]::GetTempPath()) ([System.Guid]::NewGuid().ToString())

try {
    New-Item -ItemType Directory -Path $TmpDir | Out-Null

    # ── Download ──────────────────────────────────────────────────────────────
    Write-Info "Downloading binary from $DownloadUrl..."
    if (-not $DryRun) {
        $ArchivePath = Join-Path $TmpDir $ArchiveFilename
        try {
            Invoke-WebRequest -Uri $DownloadUrl -OutFile $ArchivePath -UseBasicParsing
        } catch {
            Write-Fatal "Download failed: $_"
        }

        # Verify checksum
        try {
            $ChecksumPath = Join-Path $TmpDir "checksums.txt"
            Invoke-WebRequest -Uri $ChecksumUrl -OutFile $ChecksumPath -UseBasicParsing
            $Line     = Get-Content $ChecksumPath | Where-Object { $_ -match [regex]::Escape($ArchiveFilename) } | Select-Object -First 1
            $Expected = ($Line -split '\s+')[0].ToLower()
            $Actual   = (Get-FileHash -Path $ArchivePath -Algorithm SHA256).Hash.ToLower()
            if ($Expected -and $Actual -eq $Expected) {
                Write-Success "Checksum verified"
            } else {
                Write-Fatal "Checksum verification failed -- aborting"
            }
        } catch {
            Write-Warn "Checksum file not found -- skipping verification"
        }

        # Extract
        Expand-Archive -Path $ArchivePath -DestinationPath $TmpDir -Force
    }

    # ── Install binary ────────────────────────────────────────────────────────
    $Destination = Join-Path $InstallDir "$BinaryName.exe"
    Write-Info "Installing binary to $Destination..."
    if (-not $DryRun) {
        New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
        Copy-Item -Path (Join-Path $TmpDir $BinaryFilename) -Destination $Destination -Force
    }
    Write-Success "Binary installed"

    # ── Create directories ────────────────────────────────────────────────────
    Write-Info "Creating directories..."
    if (-not $DryRun) {
        New-Item -ItemType Directory -Path $ConfigDir -Force | Out-Null
        New-Item -ItemType Directory -Path $DataDir   -Force | Out-Null
    }
    Write-Success "Directories created"

    # ── Write default config ──────────────────────────────────────────────────
    $ConfigFile = Join-Path $ConfigDir "app.yaml"
    Write-Info "Writing default configuration to $ConfigFile..."
    if (-not $DryRun -and -not (Test-Path $ConfigFile)) {
        $DefaultConfig = @"
version: 1
server:
  listen_http: ":3000"
  listen_dns: ":53"
dns:
  upstream_servers:
    - name: Cloudflare
      address: "1.1.1.1:53"
    - name: Quad9
      address: "9.9.9.9:53"
storage:
  database_path: "openfiltr.db"
  # database_url: "postgres://openfiltr:openfiltr@localhost:5432/openfiltr?sslmode=disable"
auth:
  token_expiry_hours: 24
"@
        Set-Content -Path $ConfigFile -Value $DefaultConfig -Encoding UTF8
    }
    Write-Success "Configuration written"

    # ── Add install directory to PATH ─────────────────────────────────────────
    $Scope       = if ($NoRoot) { "User" } else { "Machine" }
    $CurrentPath = [System.Environment]::GetEnvironmentVariable("Path", $Scope)
    if ($CurrentPath -notlike "*$InstallDir*") {
        Write-Info "Adding $InstallDir to $Scope PATH..."
        if (-not $DryRun) {
            [System.Environment]::SetEnvironmentVariable("Path", "$CurrentPath;$InstallDir", $Scope)
            $env:Path += ";$InstallDir"
        }
        Write-Success "PATH updated (restart your shell to use '$BinaryName' directly)"
    }

    # ── Register Windows service (system install only) ────────────────────────
    if (-not $NoRoot) {
        Write-Info "Registering Windows service..."
        if (-not $DryRun) {
            $Existing = Get-Service -Name $BinaryName -ErrorAction SilentlyContinue
            if (-not $Existing) {
                New-Service `
                    -Name        $BinaryName `
                    -BinaryPathName "`"$Destination`" --config `"$ConfigFile`"" `
                    -DisplayName "OpenFiltr DNS Filtering Service" `
                    -Description "OpenFiltr self-hosted DNS filtering server" `
                    -StartupType Automatic
                Start-Service -Name $BinaryName
            } else {
                Write-Warn "Service '$BinaryName' already exists -- skipping"
            }
        }
        Write-Success "Windows service registered and started"
    }

    # ── Done ──────────────────────────────────────────────────────────────────
    Write-Host ""
    Write-Host "╔══════════════════════════════════════════════╗" -ForegroundColor Green
    Write-Host "║      OpenFiltr installed successfully!       ║" -ForegroundColor Green
    Write-Host "║                                              ║" -ForegroundColor Green
    Write-Host "║  Open your browser:  http://localhost:3000   ║" -ForegroundColor Green
    Write-Host "║  Complete setup to create your admin user.   ║" -ForegroundColor Green
    Write-Host "╚══════════════════════════════════════════════╝" -ForegroundColor Green
    Write-Host ""
    Write-Info "To view logs:  Get-WinEvent -LogName Application -ProviderName $BinaryName -MaxEvents 50 (or check Event Viewer)"
    Write-Info "To stop:       Stop-Service $BinaryName"
    Write-Warn "The default backend is bbolt. Set database_url only if you deliberately want PostgreSQL."

} finally {
    Remove-Item -Path $TmpDir -Recurse -Force -ErrorAction SilentlyContinue
}
