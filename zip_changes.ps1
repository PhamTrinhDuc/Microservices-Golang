# 1. Tạm thời đưa các file mới (untracked) vào chỉ mục của git
git add -N .

# 2. Lấy danh sách các file thay đổi hoặc thêm mới
$changes = git diff --name-only | Where-Object { $_ -ne "" }

if ($changes) {
    Write-Host "=========================================="
    Write-Host "Đang chuẩn bị nén các file sau:"
    $changes | ForEach-Object { Write-Host "- $_" }
    
    # Tạo thư mục tạm để gom các file (giữ nguyên cấu trúc thư mục)
    $tempDir = "./temp_patch"
    if (Test-Path $tempDir) { Remove-Item -Recurse -Force $tempDir }
    New-Item -ItemType Directory -Path $tempDir -Force | Out-Null
    
    foreach ($file in $changes) {
        if (Test-Path $file) {
            $dest = Join-Path $tempDir $file
            $parent = Split-Path $dest
            if (!(Test-Path $parent)) { New-Item -Type Directory -Path $parent -Force | Out-Null }
            Copy-Item $file -Destination $dest -Force
        }
    }
    
    # Đợi 1 giây để hệ thống giải phóng toàn bộ file handles trước khi nén
    Start-Sleep -Seconds 1
    
    # Tiến hành nén thư mục tạm thành file patch.zip
    $zipPath = "./patch.zip"
    if (Test-Path $zipPath) { Remove-Item -Force $zipPath }
    
    # Di chuyển vào trong thư mục tạm để nén nhằm tránh bị đè lock từ folder cha
    Push-Location $tempDir
    try {
        Compress-Archive -Path * -DestinationPath "../$zipPath" -Force
    }
    finally {
        Pop-Location
    }
    
    # Dọn dẹp thư mục tạm
    Remove-Item -Recurse -Force $tempDir
    
    Write-Host "SUCCESS: Đã tạo file patch.zip thành công!"
    Write-Host "=========================================="
} else {
    Write-Host "=========================================="
    Write-Host "INFO: Không có file nào thay đổi hoặc thêm mới để nén."
    Write-Host "=========================================="
}

# 3. Trả trạng thái git về ban đầu
git reset . | Out-Null


# powershell -ExecutionPolicy Bypass -File .\zip_changes.ps1