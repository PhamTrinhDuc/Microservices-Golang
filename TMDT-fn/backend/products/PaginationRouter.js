const express = require('express');
const router = express.Router();
const sql = require('../db.jsx');

router.get('/', async (req, res) => {
    const page = await sql`
            SELECT *
            FROM pagenigation
        `;

    return res.status(201).json({
        success: true,
        data: page
    })
});

router.post('/', async (req, res) => {
    const { item_in_home, column_in_home, item_in_productlist, column_in_productlist } = req.body;
    try {
        const existing = await sql`SELECT id FROM pagenigation LIMIT 1`;
        if (existing.length === 0) {
            await sql`
                INSERT INTO pagenigation (item_in_home, column_in_home, item_in_productlist, column_in_productlist)
                VALUES (${item_in_home}, ${column_in_home}, ${item_in_productlist}, ${column_in_productlist})
            `;
        } else {
            await sql`
                UPDATE pagenigation
                SET item_in_home = ${item_in_home},
                    column_in_home = ${column_in_home},
                    item_in_productlist = ${item_in_productlist},
                    column_in_productlist = ${column_in_productlist}
                WHERE id = ${existing[0].id}
            `;
        }
        const updated = await sql`SELECT * FROM pagenigation WHERE id = ${existing[0]?.id ?? 0} LIMIT 1` || [];
        return res.status(200).json({ success: true, message: "Cập nhật cấu hình hiển thị thành công!", data: updated.length > 0 ? updated : [{ item_in_home, column_in_home, item_in_productlist, column_in_productlist }] });
    } catch (err) {
        console.error("Lỗi cập nhật cấu hình phân trang:", err);
        return res.status(500).json({ success: false, message: "Lỗi hệ thống khi cập nhật hiển thị!" });
    }
});

module.exports = router;