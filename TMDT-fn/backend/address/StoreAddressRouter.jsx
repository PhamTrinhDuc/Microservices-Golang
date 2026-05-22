const express = require('express');
const router = express.Router();
const sql = require('../db.jsx');

router.post('/', async (req, res) => {
    try {
        const { id, name, hotline, province, ward, road, mail } = req.body;

        if (!name || !province || !ward || !road) {
            return res.status(400).json({
                success: false,
                message: "Vui lòng cung cấp đủ thông tin bắt buộc: Tên cửa hàng, Tỉnh/Thành, Phường/Xã và Đường."
            });
        }

        if (id) {
            const updatedStore = await sql`
                UPDATE store 
                SET name = ${name}, 
                    hotline = ${hotline || null}, 
                    province = ${province}, 
                    ward = ${ward}, 
                    road = ${road}, 
                    mail = ${mail || null}
                WHERE id = ${id}
                RETURNING *
            `;

            if (updatedStore.length === 0) {
                return res.status(404).json({ success: false, message: "Không tìm thấy cửa hàng để cập nhật." });
            }

            return res.status(200).json({
                success: true,
                message: "Cập nhật thông tin cửa hàng thành công!",
                data: updatedStore[0]
            });

        } else {
            const newStore = await sql`
                INSERT INTO store (name, hotline, province, ward, road, mail)
                VALUES (${name}, ${hotline || null}, ${province}, ${ward}, ${road}, ${mail || null})
                RETURNING *
            `;

            return res.status(201).json({
                success: true,
                message: "Thêm cửa hàng mới thành công!",
                data: newStore[0]
            });
        }

    } catch (error) {
        console.error("Lỗi API POST Store:", error.message);
        res.status(500).json({ success: false, message: "Lỗi hệ thống khi lưu cửa hàng." });
    }
});

router.get('/', async (req, res) => {
    try {
        const filterProvince = req.query.province;

        let stores;

        if (filterProvince) {
            stores = await sql`
                SELECT * FROM store 
                WHERE province ILIKE ${'%' + filterProvince + '%'}
                ORDER BY created_at DESC
            `;
        } else {
            stores = await sql`
                SELECT * FROM store 
                ORDER BY created_at DESC
            `;
        }

        res.status(200).json({
            success: true,
            total: stores.length,
            data: stores
        });

    } catch (error) {
        console.error("Lỗi API GET Stores:", error.message);
        res.status(500).json({ success: false, message: "Lỗi hệ thống khi lấy danh sách cửa hàng." });
    }
});

module.exports = router;