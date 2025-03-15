# https://docs.digitalocean.com/products/databases/postgresql/how-to/migrate/
# TODO: the curl script - migrate it to a powershell request equivalent with params

#curl -X PUT \
#-H "Content-Type: application/json" \
#-H "Authorization: Bearer $DIGITALOCEAN_TOKEN" \
#-d '{"source":{"host":"source-do-user-6607903-0.b.db.ondigitalocean.com","dbname":"defaultdb","port":25060,"username":"doadmin","password":"paakjnfe10rsrsmf"},"disable_ssl":false,"ignore_dbs":["db0","db1"]}' \
#"https://api.digitalocean.com/v2/databases/9cc10173-e9ea-4176-9dbc-a4cee4c4ff30/online-migration"

param (
    [Parameter(Mandatory = $true)]
    [string]$token,
    [Parameter(Mandatory = $true)]
    [string]$sourceHost,
    [Parameter(Mandatory = $true)]
    [string]$sourceDbName,
    [Parameter(Mandatory = $true)]
    [int]$sourcePort,
    [Parameter(Mandatory = $true)]
    [string]$sourceUserName,
    [Parameter(Mandatory = $true)]
    [SecureString]$sourcePassword,
    [Parameter(Mandatory = $true)]
    [string[]]$ignoreDbs,
    [Parameter(Mandatory = $true)]
    [string]$targetDbId
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
    disable_ssl = $false
    ignore_dbs = $ignoreDbs
} | ConvertTo-Json
$headers = @{
    "Content-Type"  = "application/json"
    "Authorization" = "Bearer $token"
}
$uri = "https://api.digitalocean.com/v2/databases/$targetDbId/online-migration"
Invoke-RestMethod -Uri $uri -Method Put -Headers $headers -Body $body -ErrorAction Stop
Write-Host "Migration started successfully."
Write-Host "Please monitor the migration status in the DigitalOcean Control Panel."
