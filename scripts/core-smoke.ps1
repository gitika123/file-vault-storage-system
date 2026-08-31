$ErrorActionPreference = 'Stop'
$root = Split-Path -Parent $PSScriptRoot
$envFile = Join-Path $root '.env'
$cfg = @{}
Get-Content $envFile | ForEach-Object { if ($_ -match '^([^#=]+)=(.*)$') { $cfg[$matches[1]] = $matches[2] } }
$base = if ($env:VAULT_BASE_URL) { $env:VAULT_BASE_URL } else { 'http://localhost:5173' }
$fixture = Join-Path $root 'tests/fixtures/sample.pdf'
$forged = Join-Path $root 'tests/fixtures/forged.docx'
$session = New-Object Microsoft.PowerShell.Commands.WebRequestSession
$passed = 0
function Assert-True([bool]$condition, [string]$message) { if (-not $condition) { throw "FAIL: $message" }; $script:passed++; Write-Host "PASS: $message" -ForegroundColor Green }
function Json-Request([string]$method, [string]$path, $body=$null, [hashtable]$headers=@{}) {
  Start-Sleep -Milliseconds 550
  $params = @{ Uri = "$base$path"; Method = $method; WebSession = $session; Headers = $headers; UseBasicParsing = $true }
  if ($null -ne $body) { $params.ContentType = 'application/json'; $params.Body = ($body | ConvertTo-Json -Compress) }
  $response = Invoke-WebRequest @params
  return ($response.Content | ConvertFrom-Json)
}
function Status-Request([string]$method, [string]$path, [hashtable]$headers=@{}) {
  Start-Sleep -Milliseconds 550
  $response = Invoke-WebRequest -Uri "$base$path" -Method $method -WebSession $session -Headers $headers -UseBasicParsing -SkipHttpErrorCheck
  if ([int]$response.StatusCode -ne 200) { Write-Host "HTTP $($response.StatusCode) for ${path}: $($response.Content)" -ForegroundColor Yellow }
  return [int]$response.StatusCode
}
Write-Host "Running core smoke tests against $base"
$alice = Json-Request POST '/api/auth/login' @{ email='alice@example.com'; password=$cfg['SEED_ALICE_PASSWORD'] }
Assert-True ($alice.email -eq 'alice@example.com') 'Alice authentication'
$csrf = (Json-Request GET '/api/auth/csrf').csrfToken
Assert-True ([string]::IsNullOrWhiteSpace($csrf) -eq $false) 'CSRF token issued'
Assert-True ((Json-Request GET '/api/auth/me').role -eq 'user') 'Current-user session'
$headers = @{ 'X-CSRF-Token' = $csrf }
$pdf = Invoke-WebRequest -Uri "$base/api/uploads" -Method POST -WebSession $session -Headers $headers -Form @{ files = Get-Item $fixture } -UseBasicParsing
$pdfResult = ($pdf.Content | ConvertFrom-Json).results[0]
Assert-True ($pdfResult.status -eq 'created') 'Valid PDF upload'
$fileId = $pdfResult.fileId
Start-Sleep -Milliseconds 550
$forgedResult = (Invoke-WebRequest -Uri "$base/api/uploads" -Method POST -WebSession $session -Headers $headers -Form @{ files = Get-Item $forged } -UseBasicParsing).Content | ConvertFrom-Json
Assert-True ($forgedResult.results[0].status -eq 'rejected' -and $forgedResult.results[0].error.code -eq 'INVALID_MIME') 'MIME mismatch rejection'
Start-Sleep -Milliseconds 550
$duplicate = (Invoke-WebRequest -Uri "$base/api/uploads" -Method POST -WebSession $session -Headers $headers -Form @{ files = Get-Item $fixture } -UseBasicParsing).Content | ConvertFrom-Json
Assert-True ($duplicate.results[0].deduplicated -eq $true) 'Duplicate content deduplication'
Assert-True ((Json-Request GET '/api/files').items.Count -ge 1) 'Owner file listing'
Assert-True ((Json-Request GET "/api/files/$fileId").uploaderEmail -eq 'alice@example.com') 'File details metadata'
$folder = Json-Request POST '/api/folders' @{ name = "Smoke-$([guid]::NewGuid().ToString('N').Substring(0,8))"; parentId='' } $headers
Assert-True ([string]::IsNullOrWhiteSpace($folder.id) -eq $false) 'Folder creation'
Json-Request PATCH "/api/files/$fileId/folder" @{ folderId=$folder.id } $headers | Out-Null
Assert-True ((Json-Request GET "/api/files?folderId=$($folder.id)").items.id -contains $fileId) 'Move file into folder'
$renamedFolder = "Renamed-$([guid]::NewGuid().ToString('N').Substring(0,8))"
Json-Request PATCH "/api/folders/$($folder.id)" @{ name=$renamedFolder } $headers | Out-Null
Assert-True ((Json-Request GET '/api/folders').name -contains $renamedFolder) 'Folder rename'
Assert-True ((Status-Request DELETE "/api/folders/$($folder.id)" $headers) -eq 409) 'Non-empty folder deletion protection'
Assert-True ((Json-Request GET "/api/files?filename=sample&mime=application/pdf&minSizeBytes=1&maxSizeBytes=100000").items.Count -ge 1) 'Combined search and filtering'
$stats = Json-Request GET "/api/stats/storage"
Assert-True ($stats.quotaBytes -gt 0 -and $stats.deduplicatedBytes -ge 0 -and $stats.originalBytes -ge $stats.deduplicatedBytes -and $stats.savingsBytes -ge 0) 'Storage statistics'
$sharedFolder = Json-Request POST '/api/folders' @{ name = "Shared-$([guid]::NewGuid().ToString('N').Substring(0,8))"; parentId='' } $headers
Json-Request PATCH "/api/files/$fileId/folder" @{ folderId=$sharedFolder.id } $headers | Out-Null
$folderShare = Json-Request POST '/api/shares/public' @{ folderId=$sharedFolder.id } $headers
Assert-True ((Status-Request GET "/public/$($folderShare.token)") -eq 200) 'Public folder landing page'
Assert-True ((Status-Request GET "/public/$($folderShare.token)/download?fileId=$fileId&preview=1") -eq 200) 'Public folder file preview'
Json-Request PATCH "/api/files/$fileId/folder" @{ folderId='' } $headers | Out-Null
Assert-True ((Status-Request DELETE "/api/folders/$($folder.id)" $headers) -eq 200) 'Empty folder deletion'
Assert-True ((Status-Request DELETE "/api/folders/$($sharedFolder.id)" $headers) -eq 200) 'Shared folder deletion after emptying'
$public = Json-Request POST '/api/shares/public' @{ fileId=$fileId } $headers
Assert-True ([string]::IsNullOrWhiteSpace($public.token) -eq $false) 'Public share creation'
$publicStatus = Status-Request GET "/public/$($public.token)/download"
Write-Host "Public download HTTP status: $publicStatus"
Assert-True ($publicStatus -eq 200) 'Public download'
Assert-True ((Status-Request GET "/api/files/$fileId/preview") -eq 200) 'Authenticated PDF preview'
Assert-True ((Status-Request GET "/api/files/$fileId/content") -eq 200) 'Authenticated download'
$adminSession = New-Object Microsoft.PowerShell.Commands.WebRequestSession
Invoke-WebRequest -Uri "$base/api/auth/login" -Method POST -WebSession $adminSession -ContentType 'application/json' -Body (@{email='admin@example.com';password=$cfg['SEED_ADMIN_PASSWORD']}|ConvertTo-Json) -UseBasicParsing | Out-Null
Start-Sleep -Milliseconds 550
Assert-True ((Invoke-WebRequest -Uri "$base/api/admin/stats" -WebSession $adminSession -UseBasicParsing).StatusCode -eq 200) 'Admin statistics authorization'
Start-Sleep -Milliseconds 550
$adminCsrf = ((Invoke-WebRequest -Uri "$base/api/auth/csrf" -Method GET -WebSession $adminSession -UseBasicParsing).Content | ConvertFrom-Json).csrfToken
Start-Sleep -Milliseconds 550
$adminUpload = (Invoke-WebRequest -Uri "$base/api/uploads" -Method POST -WebSession $adminSession -Headers @{ 'X-CSRF-Token' = $adminCsrf } -Form @{ files = Get-Item $fixture } -UseBasicParsing).Content | ConvertFrom-Json
Assert-True ($adminUpload.results[0].status -eq 'created') 'Admin file upload'
$adminFileId = $adminUpload.results[0].fileId
Start-Sleep -Milliseconds 550
$adminShare = Invoke-WebRequest -Uri "$base/api/shares/direct" -Method POST -WebSession $adminSession -Headers @{ 'X-CSRF-Token' = $adminCsrf } -ContentType 'application/json' -Body (@{ fileId=$adminFileId; recipientEmail='bob@example.com'; permission='view' }|ConvertTo-Json) -UseBasicParsing
Assert-True ([int]$adminShare.StatusCode -eq 201) 'Admin direct sharing'
Start-Sleep -Milliseconds 550
$adminInventory = (Invoke-WebRequest -Uri "$base/api/admin/files" -Method GET -WebSession $adminSession -UseBasicParsing).Content | ConvertFrom-Json
Assert-True ($adminInventory.items.uploaderEmail -contains 'admin@example.com') 'Admin file inventory details'
Assert-True ((Status-Request GET '/api/admin/stats') -eq 403) 'Normal-user admin denial'
Write-Host "Core smoke tests passed: $passed" -ForegroundColor Cyan
