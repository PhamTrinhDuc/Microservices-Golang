const express = require('express');
const router = express.Router();
const sql = require('../db.jsx');

// 1. GET ALL PROMOTIONS
router.get('/', async (req, res) => {
    try {
        const promotions = await sql`
            SELECT pr.*, p.name as product_name, p.price as product_price
            FROM promotions pr
            JOIN product p ON pr.product_id = p.product_id
            WHERE pr.is_deleted IS NOT TRUE
            ORDER BY pr.start_date DESC
        `;
        res.status(200).json({ success: true, data: promotions });
    } catch (err) {
        console.error("Lỗi lấy danh sách chương trình khuyến mãi:", err);
        res.status(500).json({ success: false, message: "Lỗi hệ thống khi lấy danh sách khuyến mãi!" });
    }
});

// 2. CREATE PROMOTION
router.post('/create', async (req, res) => {
    const { product_id, discount_percent, start_date, end_date } = req.body;

    if (!product_id || discount_percent === undefined || !start_date || !end_date) {
        return res.status(400).json({ success: false, message: "Vui lòng điền đầy đủ các thông tin bắt buộc!" });
    }

    const discount = parseInt(discount_percent, 10);
    if (isNaN(discount) || discount < 1 || discount > 99) {
        return res.status(400).json({ success: false, message: "Phần trăm giảm giá phải từ 1 đến 99%!" });
    }

    const start = new Date(start_date);
    const end = new Date(end_date);

    if (start >= end) {
        return res.status(400).json({ success: false, message: "Thời gian bắt đầu phải trước thời gian kết thúc!" });
    }

    try {
        // Check if there is already an active/overlapping promotion for this product
        const overlapping = await sql`
            SELECT promotion_id FROM promotions 
            WHERE product_id = ${product_id} 
              AND is_deleted IS NOT TRUE
              AND (
                (start_date <= ${start_date} AND end_date >= ${start_date}) OR
                (start_date <= ${end_date} AND end_date >= ${end_date}) OR
                (start_date >= ${start_date} AND end_date <= ${end_date})
              )
        `;

        if (overlapping.length > 0) {
            return res.status(400).json({ 
                success: false, 
                message: "Sản phẩm này đã có chương trình khuyến mãi đang diễn ra trong khoảng thời gian này!" 
            });
        }

        const newPromo = await sql`
            INSERT INTO promotions (product_id, discount_percent, start_date, end_date)
            VALUES (${product_id}, ${discount}, ${start_date}, ${end_date})
            RETURNING *
        `;

        res.status(201).json({ success: true, message: "Tạo chương trình khuyến mãi thành công!", data: newPromo[0] });
    } catch (err) {
        console.error("Lỗi tạo khuyến mãi:", err);
        res.status(500).json({ success: false, message: "Lỗi hệ thống khi tạo khuyến mãi!" });
    }
});

// 3. EDIT PROMOTION
router.put('/edit/:id', async (req, res) => {
    const { id } = req.params;
    const { product_id, discount_percent, start_date, end_date } = req.body;

    if (!product_id || discount_percent === undefined || !start_date || !end_date) {
        return res.status(400).json({ success: false, message: "Vui lòng điền đầy đủ các thông tin bắt buộc!" });
    }

    const discount = parseInt(discount_percent, 10);
    if (isNaN(discount) || discount < 1 || discount > 99) {
        return res.status(400).json({ success: false, message: "Phần trăm giảm giá phải từ 1 đến 99%!" });
    }

    const start = new Date(start_date);
    const end = new Date(end_date);

    if (start >= end) {
        return res.status(400).json({ success: false, message: "Thời gian bắt đầu phải trước thời gian kết thúc!" });
    }

    try {
        const overlapping = await sql`
            SELECT promotion_id FROM promotions 
            WHERE product_id = ${product_id} 
              AND promotion_id != ${id}
              AND is_deleted IS NOT TRUE
              AND (
                (start_date <= ${start_date} AND end_date >= ${start_date}) OR
                (start_date <= ${end_date} AND end_date >= ${end_date}) OR
                (start_date >= ${start_date} AND end_date <= ${end_date})
              )
        `;

        if (overlapping.length > 0) {
            return res.status(400).json({ 
                success: false, 
                message: "Sản phẩm này đã có chương trình khuyến mãi đang diễn ra trong khoảng thời gian này!" 
            });
        }

        const updated = await sql`
            UPDATE promotions SET
                product_id = ${product_id},
                discount_percent = ${discount},
                start_date = ${start_date},
                end_date = ${end_date}
            WHERE promotion_id = ${id}
            RETURNING *
        `;

        if (updated.length === 0) {
            return res.status(404).json({ success: false, message: "Không tìm thấy chương trình khuyến mãi!" });
        }

        res.status(200).json({ success: true, message: "Cập nhật khuyến mãi thành công!", data: updated[0] });
    } catch (err) {
        console.error("Lỗi cập nhật khuyến mãi:", err);
        res.status(500).json({ success: false, message: "Lỗi hệ thống khi cập nhật khuyến mãi!" });
    }
});

// 4. DELETE PROMOTION (SOFT DELETE)
router.delete('/delete/:id', async (req, res) => {
    const { id } = req.params;
    try {
        const deleted = await sql`
            UPDATE promotions 
            SET is_deleted = TRUE 
            WHERE promotion_id = ${id}
            RETURNING promotion_id
        `;
        if (deleted.length === 0) {
            return res.status(404).json({ success: false, message: "Không tìm thấy chương trình khuyến mãi để xóa!" });
        }
        res.status(200).json({ success: true, message: "Xóa chương trình khuyến mãi thành công!" });
    } catch (err) {
        console.error("Lỗi xóa khuyến mãi:", err);
        res.status(500).json({ success: false, message: "Lỗi hệ thống khi xóa khuyến mãi!" });
    }
});

module.exports = router;
