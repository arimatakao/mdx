#!/usr/bin/env pwsh
Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

$Repo = "arimatakao/mdx"
$BinName = "mdx"
$VersionInput = "latest"
$AutoYes = $false
$InstallMode = "zip"
$InstallDir = Join-Path -Path $env:LOCALAPPDATA -ChildPath "Programs\mdx"
$ReinstallConfirmed = $false

function Write-Step {
    param([string]$Message)
    Write-Host "==> $Message"
}

function Show-Usage {
@"
Usage:
  pwsh -File install.ps1 [--msi|--zip] [--install-dir <path>] [--yes] [version]

Options:
  --msi                Install via MSI.
  --zip                Install from Windows zip archive (default).
  --install-dir <dir>  Target directory for --zip mode.
  -y, --yes            Skip confirmation prompt.
  -h, --help           Show this help.

Examples:
  pwsh -File install.ps1
  pwsh -File install.ps1 --zip
  pwsh -File install.ps1 1.13.1
  pwsh -File install.ps1 v1.13.1
  pwsh -File install.ps1 --zip --install-dir "$env:USERPROFILE\bin"
  pwsh -File install.ps1 --yes
"@ | Write-Host
}

function Parse-Args {
    param([string[]]$ArgsList)

    $positional = @()
    for ($i = 0; $i -lt $ArgsList.Count; $i++) {
        $arg = $ArgsList[$i]
        switch ($arg) {
            "-h" { Show-Usage; exit 0 }
            "--help" { Show-Usage; exit 0 }
            "-y" { $script:AutoYes = $true }
            "--yes" { $script:AutoYes = $true }
            "--msi" { $script:InstallMode = "msi" }
            "--zip" { $script:InstallMode = "zip" }
            "--install-dir" {
                if ($i + 1 -ge $ArgsList.Count) {
                    throw "Error: option '--install-dir' requires a value."
                }
                $i++
                $script:InstallDir = $ArgsList[$i]
            }
            default {
                if ($arg.StartsWith("-")) {
                    throw "Error: unknown option '$arg'."
                }
                $positional += $arg
            }
        }
    }

    if ($positional.Count -gt 1) {
        throw "Error: too many positional arguments."
    }

    if ($positional.Count -eq 1) {
        $script:VersionInput = $positional[0]
    }
}

function Confirm-Install {
    param([string]$Message)

    if ($script:AutoYes) {
        return
    }

    $answer = Read-Host "$Message [y/N]"
    if ($answer -notmatch "^(y|yes)$") {
        Write-Host "Installation cancelled."
        exit 0
    }
}

function Resolve-Version {
    if ($script:VersionInput -eq "latest") {
        $latest = Invoke-RestMethod -Uri "https://api.github.com/repos/$Repo/releases/latest"
        if (-not $latest.tag_name) {
            throw "Error: unable to resolve latest release tag."
        }
        return [string]$latest.tag_name
    }

    if ($script:VersionInput.StartsWith("v")) {
        return $script:VersionInput
    }

    return "v$($script:VersionInput)"
}

function Convert-ToVersionObject {
    param([string]$Value)

    if ([string]::IsNullOrWhiteSpace($Value)) {
        return $null
    }

    $normalized = ($Value.Trim() -replace "^[vV]", "")
    $normalized = $normalized.Split("-")[0]
    $parts = $normalized.Split(".")
    if ($parts.Count -eq 0 -or $parts.Count -gt 4) {
        return $null
    }
    if ($parts | Where-Object { $_ -notmatch "^\d+$" }) {
        return $null
    }

    $padded = @($parts)
    while ($padded.Count -lt 4) {
        $padded += "0"
    }

    return [version]::Parse(($padded -join "."))
}

function Get-InstalledVersion {
    $candidates = @()
    $cmd = Get-Command -Name $BinName -ErrorAction SilentlyContinue
    if ($cmd) {
        $candidates += $cmd.Source
    }

    $localExe = Join-Path -Path $script:InstallDir -ChildPath "$BinName.exe"
    if (Test-Path -Path $localExe -PathType Leaf) {
        $candidates += $localExe
    }

    foreach ($candidate in ($candidates | Select-Object -Unique)) {
        try {
            $output = & $candidate -v 2>$null
            if (-not $output) { $output = & $candidate --version 2>$null }
            if (-not $output) { $output = & $candidate version 2>$null }
            if ($output) {
                $match = [regex]::Match(($output | Out-String), "v?\d+(\.\d+){1,3}([-.][0-9A-Za-z]+)?")
                if ($match.Success) {
                    return $match.Value
                }
            }
        }
        catch {
            continue
        }
    }

    return $null
}

function Confirm-UpgradeIfNeeded {
    param([string]$TargetVersion)

    Write-Step "Checking existing $BinName installation"
    $installedVersion = Get-InstalledVersion
    if (-not $installedVersion) {
        Write-Host "No existing $BinName installation detected."
        return
    }

    $targetObj = Convert-ToVersionObject -Value $TargetVersion
    $installedObj = Convert-ToVersionObject -Value $installedVersion
    if (-not $targetObj -or -not $installedObj) {
        return
    }

    if ($targetObj -gt $installedObj) {
        Confirm-Install "$BinName is already installed (version $installedVersion). Do you want to update to $TargetVersion?"
        return
    }

    if ($targetObj -eq $installedObj) {
        Confirm-Install "$BinName is already installed (version $installedVersion). Do you want to reinstall $TargetVersion?"
        $script:ReinstallConfirmed = $true
    }
}

function Download-File {
    param(
        [string]$Url,
        [string]$OutFile
    )

    $previousProgressPreference = $ProgressPreference
    $ProgressPreference = "SilentlyContinue"
    try {
        Invoke-WebRequest -Uri $Url -OutFile $OutFile
    }
    finally {
        $ProgressPreference = $previousProgressPreference
    }
}

function Find-MsiAssetUrl {
    param([string]$Version)

    $release = Invoke-RestMethod -Uri "https://api.github.com/repos/$Repo/releases/tags/$Version"
    $expected = "$BinName-$Version-windows-installer.msi"

    foreach ($asset in $release.assets) {
        if ($asset.name -eq $expected) {
            return [string]$asset.browser_download_url
        }
    }

    foreach ($asset in $release.assets) {
        if ($asset.name -match "windows-installer\.msi$") {
            return [string]$asset.browser_download_url
        }
    }

    throw "Error: no Windows MSI asset found in release '$Version'."
}

function Get-WindowsArch {
    $arch = $env:PROCESSOR_ARCHITECTURE
    switch ($arch) {
        "AMD64" { return "amd64" }
        "ARM64" { return "arm64" }
        "x86" { return "386" }
        default { return "amd64" }
    }
}

function Find-ZipAssetUrl {
    param([string]$Version)

    $release = Invoke-RestMethod -Uri "https://api.github.com/repos/$Repo/releases/tags/$Version"
    $arch = Get-WindowsArch
    $expected = "${BinName}_${Version}_windows_${arch}.zip"

    foreach ($asset in $release.assets) {
        if ($asset.name -eq $expected) {
            return [string]$asset.browser_download_url
        }
    }

    foreach ($asset in $release.assets) {
        if ($asset.name -match "windows_${arch}\.zip$") {
            return [string]$asset.browser_download_url
        }
    }

    throw "Error: no Windows zip asset found for arch '$arch' in release '$Version'."
}

function Install-Msi {
    param(
        [string]$Version,
        [string]$TempDir
    )

    $url = Find-MsiAssetUrl -Version $Version
    $fileName = Split-Path -Path $url -Leaf
    $msiPath = Join-Path -Path $TempDir -ChildPath $fileName

    if (-not $script:ReinstallConfirmed) {
        Confirm-Install "Install $BinName $Version from $fileName?"
    }
    Write-Step "Downloading $fileName"
    Download-File -Url $url -OutFile $msiPath

    Unblock-File -Path $msiPath -ErrorAction SilentlyContinue

    Write-Step "Starting Windows Installer"
    $proc = Start-Process -FilePath "msiexec.exe" -ArgumentList @("/i", "`"$msiPath`"") -Wait -PassThru -Verb RunAs
    if ($proc.ExitCode -ne 0) {
        throw "Error: installer failed with exit code $($proc.ExitCode)."
    }
}

function Ensure-UserPathContains {
    param([string]$Dir)

    $current = [Environment]::GetEnvironmentVariable("Path", "User")
    if ([string]::IsNullOrWhiteSpace($current)) {
        [Environment]::SetEnvironmentVariable("Path", $Dir, "User")
        return
    }

    $parts = $current.Split(";") | Where-Object { $_ -ne "" }
    if ($parts -contains $Dir) {
        return
    }

    Write-Step "Adding $Dir to user PATH"
    [Environment]::SetEnvironmentVariable("Path", "$current;$Dir", "User")
}

function Install-Zip {
    param(
        [string]$Version,
        [string]$TempDir
    )

    $url = Find-ZipAssetUrl -Version $Version
    $fileName = Split-Path -Path $url -Leaf
    $zipPath = Join-Path -Path $TempDir -ChildPath $fileName
    $extractDir = Join-Path -Path $TempDir -ChildPath "extract"
    $targetExe = Join-Path -Path $script:InstallDir -ChildPath "$BinName.exe"

    if (-not $script:ReinstallConfirmed) {
        Confirm-Install "Install $BinName $Version from $fileName to $($script:InstallDir)?"
    }
    Write-Step "Downloading $fileName"
    Download-File -Url $url -OutFile $zipPath
    Unblock-File -Path $zipPath -ErrorAction SilentlyContinue

    Write-Step "Extracting $fileName"
    Expand-Archive -Path $zipPath -DestinationPath $extractDir -Force
    $sourceExe = Join-Path -Path $extractDir -ChildPath "$BinName.exe"
    if (-not (Test-Path -Path $sourceExe -PathType Leaf)) {
        throw "Error: '$BinName.exe' was not found in archive '$fileName'."
    }

    Write-Step "Installing $BinName.exe to $($script:InstallDir)"
    New-Item -ItemType Directory -Path $script:InstallDir -Force | Out-Null
    Copy-Item -Path $sourceExe -Destination $targetExe -Force

    Ensure-UserPathContains -Dir $script:InstallDir
    Write-Host "If terminal was open, restart it to refresh PATH."
}

function Main {
    param([string[]]$CliArgs)

    Parse-Args -ArgsList $CliArgs

    if (-not $env:TEMP) {
        throw "Error: TEMP environment variable is not set."
    }

    Write-Step "Resolving target version"
    $version = Resolve-Version
    Write-Host "Target version: $version"
    Confirm-UpgradeIfNeeded -TargetVersion $version
    $tempDir = Join-Path -Path $env:TEMP -ChildPath ("mdx-install-" + [guid]::NewGuid().ToString("N"))
    New-Item -ItemType Directory -Path $tempDir -Force | Out-Null
    Write-Step "Created temporary directory $tempDir"

    try {
        if ($script:InstallMode -eq "msi") {
            Install-Msi -Version $version -TempDir $tempDir
        }
        else {
            try {
                Install-Zip -Version $version -TempDir $tempDir
            }
            catch {
                Write-Warning "Zip installation failed: $($_.Exception.Message)"
                if ($script:AutoYes) {
                    Write-Host "Trying MSI fallback..."
                    Install-Msi -Version $version -TempDir $tempDir
                }
                else {
                    $fallback = Read-Host "Try MSI installer instead? [y/N]"
                    if ($fallback -match "^(y|yes)$") {
                        Install-Msi -Version $version -TempDir $tempDir
                    }
                    else {
                        throw "Zip installation failed and MSI fallback was declined."
                    }
                }
            }
        }
    }
    finally {
        Remove-Item -Path $tempDir -Recurse -Force -ErrorAction SilentlyContinue
    }

    $invokeCmd = "$BinName --help"
    if (-not (Get-Command -Name $BinName -ErrorAction SilentlyContinue)) {
        $invokeCmd = (Join-Path -Path $script:InstallDir -ChildPath "$BinName.exe") + " --help"
    }

    Write-Host ""
    Write-Host "$BinName has been installed successfully."
    Write-Host "Run: $invokeCmd"
}

Main -CliArgs $args
