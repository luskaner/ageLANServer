#Requires -Version 5.1
<#
.SYNOPSIS
    Rebuilds every Windows amd64 binary of ageLANServer and drops it into a full
    install folder, keeping the user's existing configuration files untouched.

.DESCRIPTION
    The script mirrors the layout produced by the release "full" archive:

        <Target>\
            docs\
            launcher\launcher.exe + bin\*.exe + resources\ (configs) + docs\
            server\server.exe + bin\genCert.exe + resources\ + docs\
            battle-server-manager\battle-server-manager.exe + resources\ + docs\

    Rules:
      * Binaries are always rebuilt from the current branch with main.version
        set to -Version.
      * Code/data resources (responses, etc, per-game json) are refreshed.
      * Configuration files are NEVER overwritten if they already exist in the
        target, and are copied from -PreserveFrom when the target is new:
            launcher\resources\config.toml + config.age*.toml
            battle-server-manager\resources\config.age*.toml
            server\resources\config\config.toml
            server\resources\certificates\*.pem
            server\resources\userData\**   (profiles, match history, ...)
      * Nothing is deleted from the target.

.EXAMPLE
    # Refresh an existing install in place, keeping its configs
    .\build-windows.ps1 -Target 'D:\Juegos\AoE2DE Lan\ageLANServer_full_1.16.0_win10_x86-64'

.EXAMPLE
    # Create a new install from an older one (configs are inherited)
    .\build-windows.ps1 `
        -Target 'D:\Juegos\AoE2DE Lan\ageLANServer_full_1.16.0_win10_x86-64' `
        -PreserveFrom 'D:\Juegos\AoE2DE Lan\ageLANServer_full_1.14.0_win10_x86-64'
#>
[CmdletBinding()]
param(
    # Install folder to update/create. Defaults to a sibling of the newest
    # existing install under $InstallRoot, named after -Version.
    [string]$Target,
    # Existing install whose configuration must be inherited by a new Target.
    [string]$PreserveFrom,
    [string]$Version = '1.16.0',
    # Root that holds the release folders (only used for the default Target).
    [string]$InstallRoot = 'D:\Juegos\AoE2DE Lan',
    # Leave GOPROXY as configured (needed only if the module cache is incomplete).
    [switch]$AllowNetwork,
    # Do not write docs/ (only binaries, resources and configs).
    [switch]$SkipDocs
)

$ErrorActionPreference = 'Stop'

$root = $PSScriptRoot
$games = @('age1', 'age2', 'age3', 'age4', 'athens')
$stamp = Get-Date -Format 'yyyyMMdd-HHmmss'

function Write-Step([string]$Text) {
    Write-Host ''
    Write-Host "== $Text" -ForegroundColor Cyan
}

function Resolve-Go {
    if ($env:GOROOT) {
        $candidate = Join-Path $env:GOROOT 'bin\go.exe'
        if (Test-Path $candidate) { return $candidate }
    }
    $cmd = Get-Command go -ErrorAction SilentlyContinue
    if ($cmd) { return $cmd.Source }
    throw 'Go toolchain not found. Install it or set GOROOT.'
}

function Copy-FileIfMissing([string]$From, [string]$To) {
    if (-not (Test-Path $From)) { return $false }
    if (Test-Path $To) { return $false }
    $dir = Split-Path -Parent $To
    if (-not (Test-Path $dir)) { New-Item -ItemType Directory -Path $dir -Force | Out-Null }
    Copy-Item -LiteralPath $From -Destination $To -Force
    return $true
}

function Backup-And-Copy([string]$From, [string]$To) {
    $dir = Split-Path -Parent $To
    if (-not (Test-Path $dir)) { New-Item -ItemType Directory -Path $dir -Force | Out-Null }
    if (Test-Path $To) {
        $backup = "$To.bak"
        if (-not (Test-Path $backup)) {
            Copy-Item -LiteralPath $To -Destination $backup -Force
        }
    }
    Copy-Item -LiteralPath $From -Destination $To -Force
}

# --- Defaults -------------------------------------------------------------

if (-not $Target) {
    $suffix = "ageLANServer_full_${Version}_win10_x86-64"
    if (Test-Path $InstallRoot) {
        $Target = Join-Path $InstallRoot $suffix
    } else {
        $Target = Join-Path $root "build\package\$suffix"
    }
}
$Target = [IO.Path]::GetFullPath($Target)

if (-not $PreserveFrom -and (Test-Path $InstallRoot)) {
    $PreserveFrom = Get-ChildItem $InstallRoot -Directory -Filter 'ageLANServer_full_*' -ErrorAction SilentlyContinue |
        Where-Object { $_.FullName -ne $Target -and (Test-Path (Join-Path $_.FullName 'server\server.exe')) } |
        Sort-Object Name -Descending |
        Select-Object -First 1 -ExpandProperty FullName
}

$go = Resolve-Go
Write-Step "Build configuration"
Write-Host "repo          : $root"
Write-Host "target        : $Target"
Write-Host "preserve from : $(if ($PreserveFrom) { $PreserveFrom } else { '<none>' })"
Write-Host "version       : $Version"
Write-Host "go            : $go"

if (-not $AllowNetwork) {
    # The module cache is complete in this machine; avoid sumdb/proxy lookups.
    $env:GOPROXY = 'off'
}
$env:GOOS = 'windows'
$env:GOARCH = 'amd64'
$env:CGO_ENABLED = '0'

# --- Refuse to overwrite a running install --------------------------------

$running = Get-Process -ErrorAction SilentlyContinue | Where-Object {
    try {
        $_.Path -and $_.Path.StartsWith($Target, [StringComparison]::OrdinalIgnoreCase)
    } catch {
        $false
    }
}
if ($running) {
    $list = ($running | ForEach-Object { "$($_.ProcessName) (PID $($_.Id))" }) -join ', '
    throw "The target install is in use: $list. Close those processes and run the script again."
}

# --- Build ----------------------------------------------------------------

$stage = Join-Path $root "build\package\stage-$stamp"
$builds = @(
    @{ Pkg = './server';                     Out = 'server\server.exe' },
    @{ Pkg = './server-genCert';             Out = 'server\bin\genCert.exe' },
    @{ Pkg = './launcher';                   Out = 'launcher\launcher.exe' },
    @{ Pkg = './launcher-agent';             Out = 'launcher\bin\agent.exe' },
    @{ Pkg = './launcher-config';            Out = 'launcher\bin\config.exe' },
    @{ Pkg = './launcher-config-admin';      Out = 'launcher\bin\config-admin.exe' },
    @{ Pkg = './launcher-config-admin-agent';Out = 'launcher\bin\config-admin-agent.exe' },
    @{ Pkg = './battle-server-manager';      Out = 'battle-server-manager\battle-server-manager.exe' }
)

$ldflags = "-s -w -X main.version=$Version"

Write-Step "Building binaries (version $Version)"
Push-Location $root
try {
    foreach ($b in $builds) {
        $out = Join-Path $stage $b.Out
        $outDir = Split-Path -Parent $out
        if (-not (Test-Path $outDir)) { New-Item -ItemType Directory -Path $outDir -Force | Out-Null }
        Write-Host "  go build $($b.Pkg) -> $($b.Out)"
        & $go build -trimpath -ldflags $ldflags -o $out $b.Pkg
        if ($LASTEXITCODE -ne 0) {
            throw "go build failed for $($b.Pkg) (exit $LASTEXITCODE)"
        }
    }
} finally {
    Pop-Location
}

# --- Target skeleton -------------------------------------------------------

Write-Step "Preparing target folders"
@(
    'docs',
    'server\bin', 'server\resources\config', 'server\resources\responses', 'server\resources\etc',
    'server\resources\certificates', 'server\resources\userData',
    'launcher\bin', 'launcher\resources', 'launcher\docs',
    'battle-server-manager\resources', 'battle-server-manager\docs'
) | ForEach-Object {
    $p = Join-Path $Target $_
    if (-not (Test-Path $p)) { New-Item -ItemType Directory -Path $p -Force | Out-Null }
}

# --- Binaries --------------------------------------------------------------

Write-Step "Installing binaries"
foreach ($b in $builds) {
    $src = Join-Path $stage $b.Out
    $dst = Join-Path $Target $b.Out
    Backup-And-Copy $src $dst
    Write-Host ("  {0}  ({1:N0} bytes)" -f $b.Out, (Get-Item $dst).Length)
}
# Older full installs also shipped server\ageLANServer.exe; keep it in sync.
$altServer = Join-Path $Target 'server\ageLANServer.exe'
if (Test-Path $altServer) {
    Backup-And-Copy (Join-Path $stage 'server\server.exe') $altServer
    Write-Host '  server\ageLANServer.exe (updated in place)'
}

# --- Code/data resources (always refreshed) --------------------------------

Write-Step "Refreshing server resources"
Copy-Item "$root\server\resources\responses\*" (Join-Path $Target 'server\resources\responses') -Recurse -Force
Copy-Item "$root\server\resources\etc\*" (Join-Path $Target 'server\resources\etc') -Recurse -Force
Get-ChildItem "$root\server\resources\config" -Directory | ForEach-Object {
    Copy-Item $_.FullName (Join-Path $Target 'server\resources\config') -Recurse -Force
}

# --- Configuration (never overwrite) --------------------------------------

Write-Step "Preserving configuration"
$restored = 0

if ($PreserveFrom -and (Test-Path $PreserveFrom)) {
    $preserved = @(
        'launcher\resources\config.toml',
        'battle-server-manager\resources\config.toml',
        'server\resources\config\config.toml'
    )
    foreach ($g in $games) {
        $preserved += "launcher\resources\config.$g.toml"
        $preserved += "battle-server-manager\resources\config.$g.toml"
    }
    foreach ($rel in $preserved) {
        $src = Join-Path $PreserveFrom $rel
        $dst = Join-Path $Target $rel
        if (Copy-FileIfMissing $src $dst) {
            Write-Host "  inherited $rel"
            $restored++
        }
    }

    $certFrom = Join-Path $PreserveFrom 'server\resources\certificates'
    if (Test-Path $certFrom) {
        Get-ChildItem $certFrom -File | ForEach-Object {
            $dst = Join-Path $Target "server\resources\certificates\$($_.Name)"
            if (Copy-FileIfMissing $_.FullName $dst) {
                Write-Host "  inherited server\resources\certificates\$($_.Name)"
                $restored++
            }
        }
    }

    $userDataFrom = Join-Path $PreserveFrom 'server\resources\userData'
    $userDataTo = Join-Path $Target 'server\resources\userData'
    if ((Test-Path $userDataFrom) -and -not (Test-Path (Join-Path $userDataTo 'age2'))) {
        Copy-Item "$userDataFrom\*" $userDataTo -Recurse -Force
        Write-Host '  inherited server\resources\userData'
        $restored++
    }
}

# Defaults for anything still missing (fresh install, no PreserveFrom).
if (-not (Test-Path (Join-Path $Target 'launcher\resources\config.toml'))) {
    Copy-Item "$root\launcher\resources\config.toml" (Join-Path $Target 'launcher\resources\config.toml') -Force
    Write-Host '  default launcher\resources\config.toml'
}
foreach ($g in $games) {
    $launcherCfg = Join-Path $Target "launcher\resources\config.$g.toml"
    if (-not (Test-Path $launcherCfg)) {
        Copy-Item "$root\launcher\resources\config.game.toml" $launcherCfg -Force
        Write-Host "  default launcher\resources\config.$g.toml"
    }
    $bsmCfg = Join-Path $Target "battle-server-manager\resources\config.$g.toml"
    if (-not (Test-Path $bsmCfg)) {
        Copy-Item "$root\battle-server-manager\resources\config.game.toml" $bsmCfg -Force
        Write-Host "  default battle-server-manager\resources\config.$g.toml"
    }
}
if (-not (Test-Path (Join-Path $Target 'server\resources\config\config.toml'))) {
    Copy-Item "$root\server\resources\config\config.toml" (Join-Path $Target 'server\resources\config\config.toml') -Force
    Write-Host '  default server\resources\config\config.toml'
}
Write-Host "  configuration files inherited: $restored"

# --- Launcher scripts and docs (only when missing) -------------------------

Write-Step "Launcher scripts"
$scriptMap = @{
    'server'                 = "$root\server\resources\windows"
    'launcher'               = "$root\launcher\resources\windows"
    'battle-server-manager'  = "$root\battle-server-manager\resources\windows"
}
foreach ($component in $scriptMap.Keys) {
    # The generated start_age*.bat / clean.bat wrappers call start.bat, which
    # must live next to them in the component folder.
    if (Copy-FileIfMissing "$root\common\resources\start.bat" (Join-Path $Target "$component\start.bat")) {
        Write-Host "  added $component\start.bat"
    }
    Get-ChildItem $scriptMap[$component] -Filter *.bat -ErrorAction SilentlyContinue | ForEach-Object {
        $dst = Join-Path $Target "$component\$($_.Name)"
        if (Copy-FileIfMissing $_.FullName $dst) {
            Write-Host "  added $component\$($_.Name)"
        }
    }
}

if (-not $SkipDocs) {
    Write-Step "Docs"
    $docPairs = @(
        @{ From = 'README.md';                      To = 'docs\README.txt' },
        @{ From = 'LICENSE';                        To = 'docs\LICENSE.txt' },
        @{ From = 'server\README.md';               To = 'server\docs\README.txt' },
        @{ From = 'server\BattleServers.md';        To = 'server\docs\BattleServers.txt' },
        @{ From = 'launcher\README.md';             To = 'launcher\docs\README.txt' },
        @{ From = 'launcher-config\README.md';      To = 'launcher\docs\README-config.txt' },
        @{ From = 'battle-server-manager\README.md';To = 'battle-server-manager\docs\README.txt' }
    )
    foreach ($d in $docPairs) {
        $src = Join-Path $root $d.From
        $dst = Join-Path $Target $d.To
        if (Copy-FileIfMissing $src $dst) {
            Write-Host "  added $($d.To)"
        }
    }
}

# --- Certificates ----------------------------------------------------------

$certDir = Join-Path $Target 'server\resources\certificates'
$hasCerts = (Get-ChildItem $certDir -Filter *.pem -ErrorAction SilentlyContinue | Measure-Object).Count -gt 0
if (-not $hasCerts) {
    Write-Step "Generating certificates"
    Push-Location (Join-Path $Target 'server')
    try {
        & '.\bin\genCert.exe' -r
        if ($LASTEXITCODE -ne 0) {
            Write-Warning "genCert.exe exited with $LASTEXITCODE; generate the certificate pairs manually."
        }
    } finally {
        Pop-Location
    }
}

# --- Summary ---------------------------------------------------------------

Write-Step "Done"
$exes = @(
    'server\server.exe',
    'server\bin\genCert.exe',
    'launcher\launcher.exe',
    'launcher\bin\agent.exe',
    'launcher\bin\config.exe',
    'launcher\bin\config-admin.exe',
    'launcher\bin\config-admin-agent.exe',
    'battle-server-manager\battle-server-manager.exe'
)
$exes | ForEach-Object {
    $p = Join-Path $Target $_
    if (Test-Path $p) {
        $h = (Get-FileHash $p -Algorithm SHA256).Hash.Substring(0, 12)
        Write-Host ("  {0,-52} {1,12:N0} B  {2}" -f $_, (Get-Item $p).Length, $h)
    }
}
Write-Host ''
Write-Host "Install ready: $Target"
Write-Host 'Config files were not overwritten.'
if ($PreserveFrom) { Write-Host "Inherited from: $PreserveFrom" }
