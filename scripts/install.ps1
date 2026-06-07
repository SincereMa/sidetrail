# Install the sidetrail CLI binary.
#
# Usage:
#   .\install.ps1                          # install latest to $HOME\bin
#   .\install.ps1 -Version v0.1.0          # install a pinned version
#   .\install.ps1 -InstallDir C:\tools     # install to a custom dir
#   .\install.ps1 -Repo owner/name         # install from a fork
#
# Environment variables (overridden by parameters):
#   SIDETRAIL_VERSION                       # Same as -Version
#   SIDETRAIL_INSTALL_DIR                   # Same as -InstallDir
#   SIDETRAIL_REPO                          # Same as -Repo
#
# Requires PowerShell 5+ (Windows) or PowerShell Core 7+ (cross-platform).

[CmdletBinding()]
param(
  [string]$Version,
  [string]$InstallDir,
  [string]$Repo
)

$ErrorActionPreference = 'Stop'

if (-not $Version)    { $Version    = $env:SIDETRAIL_VERSION }
if (-not $InstallDir) { $InstallDir = $env:SIDETRAIL_INSTALL_DIR }
if (-not $Repo)       { $Repo       = $env:SIDETRAIL_REPO }

if (-not $Repo)       { $Repo = 'SincereMa/sidetrail' }
if (-not $Version)    { $Version = 'latest' }
if (-not $InstallDir) {
  if ($env:USERPROFILE) { $InstallDir = Join-Path $env:USERPROFILE 'bin' }
  else                  { $InstallDir = Join-Path $env:HOME       '.local/bin' }
}

$Project = 'sidetrail'
$IsWindows = ($IsWindows -or ($PSVersionTable.PSVersion.Major -lt 6 -and [System.Environment]::OSVersion.Platform -eq 'Win32NT'))

function Resolve-Latest {
  $url = "https://api.github.com/repos/$Repo/releases/latest"
  $body = Invoke-RestMethod -Uri $url -Method Get
  if (-not $body.tag_name) {
    throw "could not resolve latest version from $url"
  }
  return $body.tag_name
}

function Detect-Platform {
  $os = ''
  $arch = ''
  if ($IsWindows) {
    $os = 'windows'
  } elseif ($IsMacOS) {
    $os = 'darwin'
  } elseif ($IsLinux) {
    $os = 'linux'
  } else {
    throw "unsupported OS: $PSVersionTable.Platform"
  }
  switch ($env:PROCESSOR_ARCHITECTURE) {
    'AMD64' { $arch = 'amd64' }
    'X64'   { $arch = 'amd64' }
    'ARM64' { $arch = 'arm64' }
    'AARCH64' { $arch = 'arm64' }
    default {
      $m = (uname -m 2>$null)
      if     ($m -eq 'x86_64')  { $arch = 'amd64' }
      elseif ($m -eq 'aarch64' -or $m -eq 'arm64') { $arch = 'arm64' }
      else { throw "unsupported arch: $env:PROCESSOR_ARCHITECTURE / $m" }
    }
  }
  return @($os, $arch)
}

if ($Version -eq 'latest') {
  $Version = Resolve-Latest
}
if ($Version -notlike 'v*') {
  $Version = "v$Version"
}

$plat = Detect-Platform
$Os   = $plat[0]
$Arch = $plat[1]

$ext = 'tar.gz'
if ($Os -eq 'windows') { $ext = 'zip' }

$VerBare   = $Version.TrimStart('v')
$ArcName   = "${Project}_${VerBare}_${Os}_${Arch}.${ext}"
$BaseUrl   = "https://github.com/$Repo/releases/download/$Version"
$ArcUrl    = "$BaseUrl/$ArcName"
$CheckUrl  = "$BaseUrl/${Project}_${VerBare}_checksums.txt"

$Work = Join-Path ([System.IO.Path]::GetTempPath()) ("sidetrail-install-" + [System.Guid]::NewGuid().ToString('N'))
New-Item -ItemType Directory -Path $Work | Out-Null

try {
  Write-Host "install.ps1: downloading $ArcUrl"
  Invoke-WebRequest -Uri $ArcUrl -OutFile (Join-Path $Work $ArcName) -UseBasicParsing

  Write-Host "install.ps1: downloading $CheckUrl"
  Invoke-WebRequest -Uri $CheckUrl -OutFile (Join-Path $Work 'checksums.txt') -UseBasicParsing

  Write-Host "install.ps1: verifying checksum"
  $expected = (Get-Content (Join-Path $Work 'checksums.txt') |
    Where-Object { $_ -match ("\s" + [regex]::Escape($ArcName) + '$') } |
    ForEach-Object { ($_ -split '\s+')[0] } |
    Select-Object -First 1)
  if (-not $expected) { throw "checksum not found for $ArcName" }
  $actual = (Get-FileHash -Path (Join-Path $Work $ArcName) -Algorithm SHA256).Hash.ToLower()
  if ($expected.ToLower() -ne $actual) {
    throw "checksum mismatch`n  expected: $expected`n  actual:   $actual"
  }

  Write-Host "install.ps1: extracting"
  $ExtractDir = Join-Path $Work 'extract'
  New-Item -ItemType Directory -Path $ExtractDir | Out-Null
  if ($ext -eq 'zip') {
    Expand-Archive -Path (Join-Path $Work $ArcName) -DestinationPath $ExtractDir -Force
  } else {
    tar -xzf (Join-Path $Work $ArcName) -C $ExtractDir
  }

  if (-not (Test-Path $InstallDir)) { New-Item -ItemType Directory -Path $InstallDir | Out-Null }
  Write-Host "install.ps1: installing to $InstallDir/$Project"
  Copy-Item -Path (Join-Path $ExtractDir $Project) -Destination (Join-Path $InstallDir $Project) -Force

  Write-Host ""
  Write-Host "sidetrail installed at: $InstallDir/$Project"
  & (Join-Path $InstallDir $Project) --version
  Write-Host ""
  $pathDirs = ($env:PATH -split [System.IO.Path]::PathSeparator)
  if ($pathDirs -notcontains $InstallDir) {
    Write-Host "Note: $InstallDir is not on PATH. Add it to your shell profile."
  }
} finally {
  Remove-Item -Recurse -Force $Work
}
