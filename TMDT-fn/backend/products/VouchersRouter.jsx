const express = require('express');
const router = express.Router();
const sql = require('../db.jsx');

// 1. GET ALL VOUCHERS
router.get('/', async (req, res) => {
    try {
        const vouchers = await sql`
            SELECT * FROM vouchers 
            WHERE is_deleted IS NOT TRUE 
            ORDER BY start_date DESC
        `;
        res.status(200).json({ success: true, data: vouchers });
    } catch (err) {
        console.error("Lỗi lấy danh sách voucher:", err);
        res.status(500).json({ success: false, message: "Lỗi hệ thống khi lấy danh sách voucher!" });
    }
});

// 2. CREATE VOUCHER
router.post('/create', async (req, res) => {
    const { 
        code, name, start_date, end_date, discount_type, 
        discount_value, discount_target, min_order_value, max_discount_amount 
    } = req.body;

    if (!code || !name || !start_date || !end_date || !discount_type || !discount_value || !discount_target) {
        return res.status(400).json({ success: false, message: "Vui lòng nhập đầy đủ các trường bắt buộc!" });
    }

    const start = new Date(start_date);
    const end = new Date(end_date);
    const now = new Date();

    // Check time constraints: current_time < start_date < end_date
    if (start >= end) {
        return res.status(400).json({ success: false, message: "Thời gian bắt đầu phải trước thời gian kết thúc!" });
    }

    try {
        // Check uniqueness of code
        const existing = await sql`
            SELECT voucher_id FROM vouchers 
            WHERE code = ${code.toUpperCase()} AND is_deleted IS NOT TRUE
        `;
        if (existing.length > 0) {
            return res.status(400).json({ success: false, message: "Mã giảm giá này đã tồn tại!" });
        }

        const newVoucher = await sql`
            INSERT INTO vouchers (
                code, name, start_date, end_date, discount_type, 
                discount_value, discount_target, min_order_value, max_discount_amount
            ) VALUES (
                ${code.toUpperCase()}, ${name}, ${start_date}, ${end_date}, ${discount_type}, 
                ${discount_value}, ${discount_target}, ${min_order_value || 0}, ${max_discount_amount || null}
            )
            RETURNING *
        `;

        res.status(201).json({ success: true, message: "Thêm voucher thành công!", data: newVoucher[0] });
    } catch (err) {
        console.error("Lỗi tạo voucher:", err);
        res.status(500).json({ success: false, message: "Lỗi hệ thống khi tạo voucher!" });
    }
});

// 3. EDIT VOUCHER
router.put('/edit/:id', async (req, res) => {
    const { id } = req.params;
    const { 
        code, name, start_date, end_date, discount_type, 
        discount_value, discount_target, min_order_value, max_discount_amount 
    } = req.body;

    if (!code || !name || !start_date || !end_date || !discount_type || !discount_value || !discount_target) {
        return res.status(400).json({ success: false, message: "Vui lòng điền đầy đủ các thông tin bắt buộc!" });
    }

    const start = new Date(start_date);
    const end = new Date(end_date);

    if (start >= end) {
        return res.status(400).json({ success: false, message: "Thời gian bắt đầu phải trước thời gian kết thúc!" });
    }

    try {
        const existing = await sql`
            SELECT voucher_id FROM vouchers 
            WHERE code = ${code.toUpperCase()} AND voucher_id != ${id} AND is_deleted IS NOT TRUE
        `;
        if (existing.length > 0) {
            return res.status(400).json({ success: false, message: "Mã giảm giá này đã tồn tại ở voucher khác!" });
        }

        const updated = await sql`
            UPDATE vouchers SET
                code = ${code.toUpperCase()},
                name = ${name},
                start_date = ${start_date},
                end_date = ${end_date},
                discount_type = ${discount_type},
                discount_value = ${discount_value},
                discount_target = ${discount_target},
                min_order_value = ${min_order_value || 0},
                max_discount_amount = ${max_discount_amount || null}
            WHERE voucher_id = ${id}
            RETURNING *
        `;

        if (updated.length === 0) {
            return res.status(404).json({ success: false, message: "Không tìm thấy voucher!" });
        }

        res.status(200).json({ success: true, message: "Cập nhật voucher thành công!", data: updated[0] });
    } catch (err) {
        console.error("Lỗi cập nhật voucher:", err);
        res.status(500).json({ success: false, message: "Lỗi hệ thống khi cập nhật voucher!" });
    }
});

// 4. DELETE VOUCHER (SOFT DELETE)
router.delete('/delete/:id', async (req, res) => {
    const { id } = req.params;
    try {
        const deleted = await sql`
            UPDATE vouchers 
            SET is_deleted = TRUE 
            WHERE voucher_id = ${id}
            RETURNING voucher_id
        `;
        if (deleted.length === 0) {
            return res.status(404).json({ success: false, message: "Không tìm thấy voucher để xóa!" });
        }
        res.status(200).json({ success: true, message: "Xóa voucher khỏi hệ thống thành công!" });
    } catch (err) {
        console.error("Lỗi xóa voucher:", err);
        res.status(500).json({ success: false, message: "Lỗi hệ thống khi xóa voucher!" });
    }
});

// 5. APPLY VOUCHER
router.post('/apply', async (req, res) => {
    const { code, orderTotal } = req.body;

    if (!code) {
        return res.status(400).json({ success: false, message: "Vui lòng nhập mã giảm giá!" });
    }

    try {
        const vouchers = await sql`
            SELECT * FROM vouchers 
            WHERE code = ${code.toUpperCase()} AND is_deleted IS NOT TRUE
            LIMIT 1
        `;

        if (vouchers.length === 0) {
            return res.status(404).json({ success: false, message: "Mã giảm giá không hợp lệ!" });
        }

        const voucher = vouchers[0];
        const now = new Date();
        const start = new Date(voucher.start_date);
        const end = new Date(voucher.end_date);

        if (now < start) {
            return res.status(400).json({ success: false, message: "Mã giảm giá chưa đến thời gian áp dụng!" });
        }

        if (now > end) {
            return res.status(400).json({ success: false, message: "Mã giảm giá đã hết hạn!" });
        }

        const minOrder = Number(voucher.min_order_value || 0);
        const total = Number(orderTotal || 0);

        if (total < minOrder) {
            return res.status(400).json({ 
                success: false, 
                message: `Giá trị đơn hàng tối thiểu để áp dụng mã là ${minOrder.toLocaleString('vi-VN')}đ!` 
            });
        }

        let discountAmount = 0;
        const value = Number(voucher.discount_value);

        if (voucher.discount_type === 'percent') {
            discountAmount = Math.round(total * (value / 100));
            const maxDiscount = voucher.max_discount_amount ? Number(voucher.max_discount_amount) : null;
            if (maxDiscount && discountAmount > maxDiscount) {
                discountAmount = maxDiscount;
            }
        } else {
            discountAmount = value;
        }

        // Capped by orderTotal to prevent negative totals
        if (discountAmount > total) {
            discountAmount = total;
        }

        res.status(200).json({
            success: true,
            message: "Áp dụng mã giảm giá thành công!",
            discountAmount,
            discountTarget: voucher.discount_target,
            voucherCode: voucher.code
        });

    } catch (err) {
        console.error("Lỗi áp dụng voucher:", err);
        res.status(500).json({ success: false, message: "Lỗi hệ thống khi áp dụng voucher!" });
    }
});

module.exports = router;
