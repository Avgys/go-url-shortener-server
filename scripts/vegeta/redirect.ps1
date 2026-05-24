# Load-test GET /{short} (307 redirect; -redirects=-1 stops at Location header).
#
# Usage:
#   .\scripts\vegeta\redirect.ps1
#   $env:SHORT_PATH="abc12xyz"; .\scripts\vegeta\redirect.ps1
#   $env:SEED_COUNT="10"; .\scripts\vegeta\redirect.ps1

param(
    [string]$BaseUrl = $(if ($env:BASE_URL) { $env:BASE_URL } else { "http://localhost:8080" }),
    [string]$Rate = $(if ($env:RATE) { $env:RATE } else { "50" }),
    [string]$Duration = $(if ($env:DURATION) { $env:DURATION } else { "30s" }),
    [int]$SeedCount = $(if ($env:SEED_COUNT) { [int]$env:SEED_COUNT } else { 1 }),
    [string]$ShortPath = $(if ($env:SHORT_PATH) { $env:SHORT_PATH } else { "" }),
    [string]$LongHost = $(if ($env:LONG_HOST) { $env:LONG_HOST } else { "http://ofdafnyylfqe.biz/page" }),
    [string]$Output = $(if ($env:OUTPUT) { $env:OUTPUT } else { "" })
)

function Get-ShortPath {
    $id = [guid]::NewGuid().ToString("n").Substring(0, 8)
    $long = "$LongHost/$id"
    $resp = Invoke-WebRequest -Method POST -Uri "$BaseUrl/" -ContentType "text/plain" -Body $long -UseBasicParsing
    $short = $resp.Content.Trim()
    if ($short.StartsWith($BaseUrl)) {
        return $short.Substring($BaseUrl.Length).TrimStart("/")
    }
    return $short.TrimStart("/")
}

$paths = @()
if ($ShortPath) {
    $paths = @($ShortPath.TrimStart("/"))
} else {
    Write-Host "seeding $SeedCount short URL(s)..."
    1..$SeedCount | ForEach-Object {
        $path = Get-ShortPath
        Write-Host "  -> $BaseUrl/$path"
        $paths += $path
    }
}

$targets = New-Object System.Text.StringBuilder
foreach ($path in $paths) {
    [void]$targets.Append("GET $BaseUrl/$path`n")
    [void]$targets.Append("`n")
}

Write-Host "redirect attack: base=$BaseUrl rate=$Rate duration=$Duration"

$attackArgs = @("-rate=${Rate}", "-duration=${Duration}", "-redirects=-1")
$binFile = if ($Output) { $Output } else { [System.IO.Path]::GetTempFileName() }
$attackArgs += "-output=${binFile}"

$targets.ToString() | vegeta attack @attackArgs
vegeta report $binFile

if (-not $Output) {
    Remove-Item -Force $binFile
}
