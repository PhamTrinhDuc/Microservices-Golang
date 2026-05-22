const express = require('express');
const router = express.Router();
const sql = require("../db.jsx");

router.get('/:userId', async (req, res) => {
    try {
        const userId = req.params.userId;
        const addresses = await sql`
            SELECT id, id_user, full_name, num_phone, province, ward, detail_address, is_default, is_delete
            FROM address 
            WHERE id_user = ${userId}
            AND "is_delete" = false
            ORDER BY is_default DESC, id DESC
        `;
        res.json({ success: true, data: addresses });
    } catch (error) {
        res.status(500).json({ success: false, message: error.message });
    }
});

router.post('/', async (req, res) => {
    const { id_user, full_name, num_phone, province, ward, detail_address, is_default } = req.body;

    try {
        const result = await sql.begin(async (trx) => {
            if (is_default) {
                await trx`UPDATE address SET is_default = false WHERE id_user = ${id_user}`;
            }

            const [newAddr] = await trx`
                INSERT INTO address (id_user, full_name, num_phone, province, ward, detail_address, is_default)
                VALUES (${id_user}, ${full_name}, ${num_phone}, ${province}, ${ward}, ${detail_address}, ${is_default || false})
                RETURNING *
            `;
            return newAddr;
        });

        res.status(201).json({ success: true, data: result });
    } catch (error) {
        res.status(500).json({ success: false, message: error.message });
    }
});

router.put('/:id', async (req, res) => {
    const { id } = req.params;
    const { id_user, full_name, num_phone, province, ward, detail_address, is_default } = req.body;

    try {
        await sql.begin(async (trx) => {
            if (is_default) {
                await trx`UPDATE address SET is_default = false WHERE id_user = ${id_user}`;
            }

            await trx`
                UPDATE address SET 
                    full_name = ${full_name},
                    num_phone = ${num_phone},
                    province = ${province},
                    ward = ${ward},
                    detail_address = ${detail_address},
                    is_default = ${is_default}
                WHERE id = ${id}
            `;
        });
        res.json({ success: true, message: "Cập nhật thành công" });
    } catch (error) {
        res.status(500).json({ success: false, message: error.message });
    }
});

router.delete('/:id', async (req, res) => {
    try {
        const { id } = req.params;
        const result = await sql`
            UPDATE address 
            SET "is_delete" = true 
            WHERE id = ${id}
            RETURNING *
        `;

        if (result.length === 0) {
            return res.status(404).json({ success: false, message: "Không tìm thấy địa chỉ" });
        }

        res.json({ success: true, message: "Đã xóa địa chỉ thành công" });
    } catch (error) {
        console.log('error: ', error);
        res.status(500).json({ success: false, message: "Lỗi hệ thống khi xóa địa chỉ" });
    }
});

router.patch('/:id/default', async (req, res) => {
    const { id } = req.params;
    const { userId } = req.body;

    try {
        await sql.begin(async (trx) => {
            await trx`UPDATE address SET is_default = false WHERE id_user = ${userId}`;
            await trx`UPDATE address SET is_default = true WHERE id = ${id}`;
        });
        res.json({ success: true, message: "Đã đặt làm mặc định" });
    } catch (error) {
        res.status(500).json({ success: false, message: error.message });
    }
});

module.exports = router;