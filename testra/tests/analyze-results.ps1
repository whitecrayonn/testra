$json = Get-Content 'c:\Private\project\testra\tests\reports\json\test-results.json' -Raw | ConvertFrom-Json

Write-Host "=== OVERALL STATS ==="
$json.stats | Format-List

Write-Host "`n=== FAILURES BY BROWSER ==="
$failed = @()
foreach($suite in $json.suites) {
    foreach($spec in $suite.specs) {
        foreach($test in $spec.tests) {
            $hasFailed = $false
            foreach($result in $test.results) {
                if($result.status -eq 'failed') {
                    $hasFailed = $true
                }
            }
            if($hasFailed) {
                $failed += [PSCustomObject]@{
                    Title = $spec.title
                    Project = $test.projectName
                    File = $suite.title
                }
            }
        }
    }
}

$failed | Group-Object Project | Select-Object Name, Count | Sort-Object Count -Descending | Format-Table -AutoSize

Write-Host "`n=== FAILURES BY TEST FILE ==="
$failed | Group-Object File | Select-Object Name, Count | Sort-Object Count -Descending | Format-Table -AutoSize

Write-Host "`n=== UNIQUE FAILING TEST TITLES ==="
$failed | Select-Object -Unique Title | Sort-Object Title
