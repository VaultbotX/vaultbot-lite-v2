param (
    [Parameter(Mandatory=$true)]
    [string]$BucketName,
    [Parameter(Mandatory=$true)]
    [string]$TfStateKey,
    [Parameter(Mandatory=$true)]
    [string]$DoSpacesAccessKey,
    [Parameter(Mandatory=$true)]
    [string]$DoSpacesSecretKey
)

$initArgs = @(
    "init",
    "-backend-config=access_key=$DoSpacesAccessKey",
    "-backend-config=secret_key=$DoSpacesSecretKey",
    "-backend-config=bucket=$BucketName",
    "-backend-config=key=$TfStateKey"
)

terraform @initArgs