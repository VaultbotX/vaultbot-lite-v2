# https://docs.digitalocean.com/products/databases/postgresql/how-to/migrate/

param (
    [Parameter(Mandatory = $true)]
    [string]$digitalOceanAccessToken,
    [Parameter(Mandatory = $true)]
    [string]$sourceHost,
    [string]$sourceDbName = "vaultbot",
    [int]$sourcePort = 5432,
    [Parameter(Mandatory = $true)]
    [string]$sourceUserName,
    [Parameter(Mandatory = $true)]
    [SecureString]$sourcePassword,
    [string[]]$ignoreDbs = @("postgres"),
    [Parameter(Mandatory = $true)]
    [string]$targetDbId,
    [Boolean]$disableSsl = $false
)

$sourcePasswordPlainText = [System.Runtime.InteropServices.Marshal]::PtrToStringAuto([System.Runtime.InteropServices.Marshal]::SecureStringToBSTR($sourcePassword))
$body = @{
    source = @{
        host = $sourceHost
        dbname = $sourceDbName
        port = $sourcePort
        username = $sourceUserName
        password = $sourcePasswordPlainText
    }
    disable_ssl = $disableSsl
    ignore_dbs = $ignoreDbs
} | ConvertTo-Json
$headers = @{
    "Content-Type"  = "application/json"
    "Authorization" = "Bearer $digitalOceanAccessToken"
}

Write-Host "Starting migration..."
$uri = "https://api.digitalocean.com/v2/databases/$targetDbId/online-migration"
Invoke-RestMethod -Uri $uri -Method Put -Headers $headers -Body $body -ErrorAction Stop
Write-Host "Migration started successfully."
Write-Host "Please monitor the migration status in the DigitalOcean Control Panel."
