$r = Get-Content reports\json\test-results.json -Raw | ConvertFrom-Json
function Get-Errors($suites) {
    foreach($s in $suites) {
        if($s.specs) {
            foreach($sp in $s.specs) {
                foreach($t in $sp.tests) {
                    if($t.results) {
                        foreach($res in $t.results) {
                            if($res.status -eq 'failed' -and $res.retry -eq 0 -and $t.projectName -eq 'chromium') {
                                $msg = $res.error.message
                                if($msg.Length -gt 300) { $msg = $msg.Substring(0,300) }
                                Write-Host ""
                                Write-Host "=== $($sp.file) | $($sp.title) ==="
                                Write-Host $msg
                            }
                        }
                    }
                }
            }
        }
        if($s.suites) { Get-Errors $s.suites }
    }
}
Get-Errors $r.suites
