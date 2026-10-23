param([string]$Out = "smart-env.exe")
$ErrorActionPreference = "Stop"
go build -ldflags "-s -w -X main.version=dev" -o $Out ./cmd/smart-env
Write-Host "Built $Out"