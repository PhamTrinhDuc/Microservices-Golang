const express = require('express');
const router = express.Router();
const sql = require('../db.jsx');

// 1. Submit review for an order (creates reviews for all products in the order)
router.post('/', async (req, res) => {
    const { user_id, order_id, rating, comment } = req.body;

    if (!user_id || !order_id || !rating) {
        return res.status(400).json({ success: false, message: 'Thiếu thông tin đánh giá bắt buộc' });
    }

    try {
        // Check if order has already been reviewed
        const [existingReview] = await sql`
            SELECT 1 FROM reviews WHERE order_id = ${order_id} LIMIT 1
        `;

        if (existingReview) {
            return res.status(400).json({ success: false, message: 'Đơn hàng này đã được đánh giá trước đó' });
        }

        // Get all products in the order
        const items = await sql`
            SELECT DISTINCT product_id 
            FROM order_details 
            WHERE order_id = ${order_id}
        `;

        if (items.length === 0) {
            return res.status(404).json({ success: false, message: 'Không tìm thấy sản phẩm nào trong đơn hàng' });
        }

        // Insert review for each product
        await sql.begin(async (trx) => {
            for (const item of items) {
                await trx`
                    INSERT INTO reviews (user_id, product_id, order_id, rating, comment, status)
                    VALUES (${user_id}, ${item.product_id}, ${order_id}, ${rating}, ${comment || ''}, 'pending')
                `;
            }
        });

        return res.status(201).json({ success: true, message: 'Đánh giá đã được gửi và đang chờ phê duyệt!' });
    } catch (error) {
        console.error('Error submitting review:', error);
        return res.status(500).json({ success: false, message: 'Lỗi máy chủ khi gửi đánh giá', error: error.message });
    }
});

// 2. Get approved reviews for a specific product
router.get('/:productId', async (req, res) => {
    const { productId } = req.params;

    try {
        const reviews = await sql`
            SELECT 
                r.review_id, 
                r.rating, 
                r.comment, 
                r.created_at, 
                u.full_name
            FROM reviews r
            JOIN users u ON r.user_id = u.id
            WHERE r.product_id = ${productId} AND r.status = 'approved'
            ORDER BY r.created_at DESC
        `;

        return res.json({ success: true, data: reviews });
    } catch (error) {
        console.error('Error fetching product reviews:', error);
        return res.status(500).json({ success: false, message: 'Lỗi máy chủ khi lấy danh sách đánh giá', error: error.message });
    }
});

// 3. Admin: Get all reviews
router.get('/admin/all', async (req, res) => {
    try {
        const reviews = await sql`
            SELECT 
                r.review_id, 
                r.rating, 
                r.comment, 
                r.status, 
                r.created_at, 
                r.order_id,
                r.product_id,
                u.full_name AS customer_name, 
                p.name AS product_name
            FROM reviews r
            JOIN users u ON r.user_id = u.id
            JOIN product p ON r.product_id = p.product_id
            ORDER BY r.created_at DESC
        `;

        return res.json({ success: true, data: reviews });
    } catch (error) {
        console.error('Error fetching admin reviews:', error);
        return res.status(500).json({ success: false, message: 'Lỗi máy chủ khi lấy tất cả đánh giá', error: error.message });
    }
});

// 4. Admin: Approve a review
router.patch('/:reviewId/approve', async (req, res) => {
    const { reviewId } = req.params;

    try {
        const result = await sql`
            UPDATE reviews 
            SET status = 'approved' 
            WHERE review_id = ${reviewId}
            RETURNING review_id
        `;

        if (result.length === 0) {
            return res.status(404).json({ success: false, message: 'Không tìm thấy đánh giá' });
        }

        return res.json({ success: true, message: 'Đã duyệt đánh giá thành công!' });
    } catch (error) {
        console.error('Error approving review:', error);
        return res.status(500).json({ success: false, message: 'Lỗi máy chủ khi duyệt đánh giá', error: error.message });
    }
});

// 5. Admin: Delete a review
router.delete('/:reviewId', async (req, res) => {
    const { reviewId } = req.params;

    try {
        const result = await sql`
            DELETE FROM reviews 
            WHERE review_id = ${reviewId}
            RETURNING review_id
        `;

        if (result.length === 0) {
            return res.status(404).json({ success: false, message: 'Không tìm thấy đánh giá' });
        }

        return res.json({ success: true, message: 'Đã xóa đánh giá thành công!' });
    } catch (error) {
        console.error('Error deleting review:', error);
        return res.status(500).json({ success: false, message: 'Lỗi máy chủ khi xóa đánh giá', error: error.message });
    }
});

module.exports = router;
