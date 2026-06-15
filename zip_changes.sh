#!/bin/bash

# 1. Tạm thời đưa các file mới tạo (untracked) vào chỉ mục của git để so khớp
git add -N .

# 2. Lấy danh sách tất cả các file có sự thay đổi hoặc thêm mới
changes=$(git diff --name-only)

if [ -n "$changes" ]; then
    # 3. Tạo file zip chứa các file thay đổi
    # Sử dụng git archive để zip đúng cấu trúc thư mục
    git archive -o patch.zip HEAD $changes
    echo "=========================================="
    echo "SUCCESS: Đã tạo file patch.zip thành công!"
    echo "Các file đã được nén vào zip:"
    echo "$changes"
    echo "=========================================="
else
    echo "=========================================="
    echo "INFO: Không có file nào thay đổi hoặc thêm mới để nén."
    echo "=========================================="
fi

# 4. Trả các file untracked về trạng thái cũ
git reset . > /dev/null 2>&1


# bash ./zip_changes.sh