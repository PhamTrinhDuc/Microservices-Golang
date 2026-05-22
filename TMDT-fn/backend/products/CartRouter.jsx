const express = require('express');
const router = express.Router();
const sql = require('../db.jsx');

// Helper to extract capacity from product name (e.g. "OPPO A6t 4GB/64GB" -> "4GB/64GB")
function extractCapacityFromName(name) {
    if (!name) return '';
    const match = name.match(/\d+GB|\d+TB/g);
    return match ? match.join('/') : '';
}

// Helper to format price display
function formatPrice(price) {
    return price.toLocaleString('vi-VN') + '₫';
}

// 1. GET /api/cart/:userId - Get all cart items for a user
router.get('/:userId', async (req, res) => {
    const userId = Number(req.params.userId);
    if (isNaN(userId)) {
        return res.status(400).json({ success: false, error: 'User ID không hợp lệ.' });
    }

    try {
        const items = await sql`
            SELECT
                ci.variant_id,
                ci.quantity,
                pv.product_id,
                p.category_id,
                p.name AS product_name,
                p.img_thumb AS main_img_thumb, 
                pv.color_name,
                pv.color_code,
                pv.price_str,
                pv.price AS original_price,
                pv.local_gallery,
                COALESCE(
                        pv.price - (pv.price * pr.discount_percent / 100),
                        pv.price
                ) AS calculated_price,
                COALESCE(pi.quantity - pi.reserved, 0) AS stock
            FROM cart_items ci
                     JOIN productvariant pv ON ci.variant_id = pv.variant_id
                     JOIN product p ON pv.product_id = p.product_id
                     LEFT JOIN productinventory pi ON pv.variant_id = pi.variant_id
                     LEFT JOIN promotions pr ON p.product_id = pr.product_id
                AND pr.is_deleted = false
                AND CURRENT_TIMESTAMP >= pr.start_date
                AND CURRENT_TIMESTAMP <= pr.end_date
            WHERE ci.user_id = ${userId}
            ORDER BY ci.created_at DESC
        `;

        const formattedItems = items.map(item => {
            const finalPrice = Number(item.calculated_price);

            const imgPath = (item.local_gallery && item.local_gallery.length > 0)
                ? item.local_gallery[0]
                : (item.main_img_thumb || '');

            // ĐÃ SỬA: Chèn thêm item.category_id vào giữa
            const img_thumb = (imgPath.startsWith('http') || imgPath.startsWith('/'))
                ? imgPath
                : `/${item.category_id}/${item.product_id}/${imgPath}`;
            return{
                variant_id: item.variant_id,
                quantity: item.quantity,
                color_name: item.color_name,
                color_code: item.color_code,
                capacity: extractCapacityFromName(item.product_name),
                product: {
                    id: item.product_id,
                    product_id: item.product_id,
                    name: item.product_name,
                    img_thumb: img_thumb,
                    price: finalPrice,
                    calculated_price: finalPrice,
                    formatted_price: formatPrice(finalPrice),
                    price_str: item.price_str,
                    stock: Math.max(0, item.stock)
                }
            };
        });

        res.json({ success: true, cart: formattedItems });
    } catch (err) {
        console.error('Error fetching cart:', err);
        res.status(500).json({ success: false, error: 'Lỗi server khi lấy thông tin giỏ hàng.' });
    }
});

// 2. POST /api/cart/add - Add item to cart (or increment if already exists)
router.post('/add', async (req, res) => {
    const { user_id, variant_id, quantity } = req.body;
    const qty = Number(quantity) || 1;

    if (!user_id || !variant_id) {
        return res.status(400).json({ success: false, error: 'Thiếu thông tin user_id hoặc variant_id.' });
    }

    try {
        await sql`
            INSERT INTO cart_items (user_id, variant_id, quantity)
            VALUES (${user_id}, ${variant_id}, ${qty})
            ON CONFLICT (user_id, variant_id)
            DO UPDATE SET 
                quantity = cart_items.quantity + EXCLUDED.quantity,
                updated_at = CURRENT_TIMESTAMP
        `;
        res.json({ success: true, message: 'Đã thêm sản phẩm vào giỏ hàng.' });
    } catch (err) {
        console.error('Error adding to cart:', err);
        res.status(500).json({ success: false, error: 'Lỗi server khi thêm sản phẩm vào giỏ hàng.' });
    }
});

// 3. PUT /api/cart/update - Update quantity of an item
router.put('/update', async (req, res) => {
    const { user_id, variant_id, quantity } = req.body;

    if (!user_id || !variant_id || quantity === undefined) {
        return res.status(400).json({ success: false, error: 'Thiếu thông tin bắt buộc.' });
    }

    const qty = Number(quantity);
    if (qty < 1) {
        return res.status(400).json({ success: false, error: 'Số lượng phải lớn hơn hoặc bằng 1.' });
    }

    try {
        const [updated] = await sql`
            UPDATE cart_items
            SET quantity = ${qty}, updated_at = CURRENT_TIMESTAMP
            WHERE user_id = ${user_id} AND variant_id = ${variant_id}
            RETURNING id
        `;

        if (!updated) {
            return res.status(404).json({ success: false, error: 'Không tìm thấy sản phẩm trong giỏ hàng.' });
        }

        res.json({ success: true, message: 'Cập nhật số lượng thành công.' });
    } catch (err) {
        console.error('Error updating cart quantity:', err);
        res.status(500).json({ success: false, error: 'Lỗi server khi cập nhật giỏ hàng.' });
    }
});

// 4. DELETE /api/cart/remove - Remove an item from cart
router.delete('/remove', async (req, res) => {
    const { user_id, variant_id } = req.body;

    if (!user_id || !variant_id) {
        return res.status(400).json({ success: false, error: 'Thiếu thông tin user_id hoặc variant_id.' });
    }

    try {
        const [deleted] = await sql`
            DELETE FROM cart_items
            WHERE user_id = ${user_id} AND variant_id = ${variant_id}
            RETURNING id
        `;

        if (!deleted) {
            return res.status(404).json({ success: false, error: 'Không tìm thấy sản phẩm trong giỏ hàng.' });
        }

        res.json({ success: true, message: 'Đã xóa sản phẩm khỏi giỏ hàng.' });
    } catch (err) {
        console.error('Error removing from cart:', err);
        res.status(500).json({ success: false, error: 'Lỗi server khi xóa sản phẩm khỏi giỏ hàng.' });
    }
});

// 5. POST /api/cart/sync - Sync localStorage cart items to Database upon login
router.post('/sync', async (req, res) => {
    const { user_id, items } = req.body;

    if (!user_id) {
        return res.status(400).json({ success: false, error: 'Thiếu user_id.' });
    }

    if (!Array.isArray(items)) {
        return res.status(400).json({ success: false, error: 'Danh sách sản phẩm không hợp lệ.' });
    }

    try {
        await sql.begin(async (trx) => {
            for (const item of items) {
                const variant_id = item.variant_id || (item.product && item.product.variant_id);
                const quantity = Number(item.quantity) || 1;
                
                if (!variant_id) continue;

                await trx`
                    INSERT INTO cart_items (user_id, variant_id, quantity)
                    VALUES (${user_id}, ${variant_id}, ${quantity})
                    ON CONFLICT (user_id, variant_id)
                    DO UPDATE SET 
                        quantity = GREATEST(cart_items.quantity, EXCLUDED.quantity),
                        updated_at = CURRENT_TIMESTAMP
                `;
            }
        });
        res.json({ success: true, message: 'Đồng bộ giỏ hàng thành công.' });
    } catch (err) {
        console.error('Error syncing cart:', err);
        res.status(500).json({ success: false, error: 'Lỗi server khi đồng bộ giỏ hàng.' });
    }
});

// 6. DELETE /api/cart/clear/:userId - Clear entire cart for a user
router.delete('/clear/:userId', async (req, res) => {
    const userId = Number(req.params.userId);
    if (isNaN(userId)) {
        return res.status(400).json({ success: false, error: 'User ID không hợp lệ.' });
    }

    try {
        await sql`
            DELETE FROM cart_items
            WHERE user_id = ${userId}
        `;
        res.json({ success: true, message: 'Đã xóa toàn bộ giỏ hàng.' });
    } catch (err) {
        console.error('Error clearing cart:', err);
        res.status(500).json({ success: false, error: 'Lỗi server khi xóa giỏ hàng.' });
    }
});

module.exports = router;
