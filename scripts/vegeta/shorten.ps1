# Load-test POST / (plain-text shorten).
#
# Usage:
#   .\scripts\vegeta\shorten.ps1
#   .\scripts\vegeta\shorten.ps1 -Rate 100 -Duration 30s -Count 1000

param(
    [string]$BaseUrl = $(if ($env:BASE_URL) { $env:BASE_URL } else { "http://localhost:8080" }),
    [string]$Rate = $(if ($env:RATE) { $env:RATE } else { "100" }),
    [string]$Duration = $(if ($env:DURATION) { $env:DURATION } else { "10s" }),
    [int]$Count = $(if ($env:COUNT) { [int]$env:COUNT } else { 500 }),
    [string]$LongHost = $(if ($env:LONG_HOST) { $env:LONG_HOST } else { "http://ofdafnyylfqe.biz/page" }),
    [string]$Output = $(if ($env:OUTPUT) { $env:OUTPUT } else { "" })
)

$lines = New-Object System.Collections.Generic.List[string]
1..$Count | ForEach-Object {
    $id = [guid]::NewGuid().ToString("n").Substring(0, 8)
    $body = "$LongHost/$id"
    $payload = @{
        method = "POST"
        url    = "$BaseUrl/"
        body   = [Convert]::ToBase64String([Text.Encoding]::UTF8.GetBytes($body))
        header = @{ "Content-Type" = @("text/plain") }
    } | ConvertTo-Json -Compress
    $lines.Add($payload)
}

Write-Host "shorten attack: base=$BaseUrl rate=$Rate duration=$Duration targets=$Count"

$attackArgs = @("-format", "json", "-rate=${Rate}", "-duration=${Duration}")
$binFile = if ($Output) { $Output } else { [System.IO.Path]::GetTempFileName() }
$attackArgs += "-output=${binFile}"

($lines -join "`n") | vegeta attack @attackArgs
vegeta report $binFile

if (-not $Output) {
    Remove-Item -Force $binFile
}
