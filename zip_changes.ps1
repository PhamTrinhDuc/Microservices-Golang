# 1. Tạm thời đưa các file mới (untracked) vào chỉ mục của git
git add -N .

# 2. Lấy danh sách các file thay đổi hoặc thêm mới
$changes = git diff --name-only | Where-Object { $_ -ne "" }

if ($changes) {
    Write-Host "=========================================="
    Write-Host "Đang chuẩn bị nén các file sau:"
    $changes | ForEach-Object { Write-Host "- $_" }
    
    $zipPath = "./patch.zip"
    if (Test-Path $zipPath) { Remove-Item -Force $zipPath }
    
    # Load thư viện nén của .NET
    Add-Type -AssemblyName System.IO.Compression
    Add-Type -AssemblyName System.IO.Compression.FileSystem
    
    # Lấy đường dẫn tuyệt đối của file zip đích
    $absoluteZipPath = [System.IO.Path]::GetFullPath($zipPath)
    
    # Khởi tạo Archive
    $zipArchive = [System.IO.Compression.ZipFile]::Open($absoluteZipPath, [System.IO.Compression.ZipArchiveMode]::Create)
    
    try {
        foreach ($file in $changes) {
            if (Test-Path $file) {
                $absoluteFilePath = [System.IO.Path]::GetFullPath($file)
                
                # Mở file stream với chế độ cho phép chia sẻ đọc/ghi (tránh lỗi file lock)
                $fileStream = New-Object System.IO.FileStream(
                    $absoluteFilePath,
                    [System.IO.FileMode]::Open,
                    [System.IO.FileAccess]::Read,
                    [System.IO.FileShare]::ReadWrite
                )
                
                try {
                    # Tạo entry trong file zip và giữ nguyên cấu trúc đường dẫn tương đối (dùng ký tự gạch chéo '/')
                    $entryName = $file.Replace("\", "/")
                    $entry = $zipArchive.CreateEntry($entryName)
                    $entryStream = $entry.Open()
                    
                    try {
                        $fileStream.CopyTo($entryStream)
                    } finally {
                        $entryStream.Close()
                        $entryStream.Dispose()
                    }
                } finally {
                    $fileStream.Close()
                    $fileStream.Dispose()
                }
            }
        }
        Write-Host "SUCCESS: Đã tạo file patch.zip thành công bằng .NET ZipArchive!"
    } catch {
        Write-Error "Lỗi khi nén file: $_"
    } finally {
        $zipArchive.Dispose()
    }
    Write-Host "=========================================="
} else {
    Write-Host "=========================================="
    Write-Host "INFO: Không có file nào thay đổi hoặc thêm mới để nén."
    Write-Host "=========================================="
}

# 3. Trả trạng thái git về ban đầu
git reset . | Out-Null

# powershell -ExecutionPolicy Bypass -File .\zip_changes.ps1