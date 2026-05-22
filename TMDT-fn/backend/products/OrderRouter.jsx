const express = require('express');
const router = express.Router();
const sql = require('../db.jsx');
const { generateVietQR } = require('../services/generateQR.jsx');
const payos = require('../services/payosService.js');
const crypto = require('crypto');
const {
    reserveInventory,
    confirmCheckout,
    checkoutDirect,
    cancelReservation,
    checkoutFreeOrder,
} = require('../services/inventoryService.js');

function validateBody(requiredFields) {
    return (req, res, next) => {
        const missing = requiredFields.filter((f) => req.body[f] === undefined || req.body[f] === null);
        if (missing.length > 0) {
            return res.status(400).json({
                success: false,
                error: `Thiếu thông tin bắt buộc: ${missing.join(', ')}`,
            });
        }
        next();
    };
}

// ════════════════════════════════════════════════════════════════════════════════
// PHASE 1: GIỮ HÀNG (INVENTORY RESERVATION)
// ════════════════════════════════════════════════════════════════════════════════
// router.post('/reserve', validateBody(['user_id', 'items']), async (req, res) => {
//     const { user_id, items } = req.body;
//
//     if (!Array.isArray(items) || items.length === 0) {
//         return res.status(400).json({ success: false, error: 'Danh sách sản phẩm không hợp lệ.' });
//     }
//
//     for (const item of items) {
//         if (!item.variant_id || !item.quantity || item.quantity <= 0) {
//             return res.status(400).json({ success: false, error: 'Mỗi sản phẩm cần có loại và số lượng hợp lệ.' });
//         }
//     }
//
//     try {
//         const reservation = await reserveInventory(user_id, items);
//         return res.status(200).json({
//             success: true,
//             message: `Đã giữ hàng thành công. Bạn có 15 phút để hoàn tất thanh toán.`,
//             reservation_id: reservation.reservation_id,
//             expires_at: reservation.expires_at,
//             items: reservation.items,
//         });
//     } catch (error) {
//         console.error('[Reserve] Error:', error.message);
//         return res.status(400).json({ success: false, error: error.message });
//     }
// });
router.post('/reserve', validateBody(['user_id', 'items']), async (req, res) => {
    const { user_id, items, address_id, payment_method, note } = req.body;

    try {
        const reservation = await reserveInventory(user_id, items);

        // Tính tổng tiền từ giá thật DB + CỘNG tiền Ship - TRỪ Voucher (từ frontend truyền lên)
        const totalAmount = reservation.items.reduce((sum, item) => {
            // "items" là req.body.items truyền lên từ React
            const clientItem = items.find(i => Number(i.variant_id) === Number(item.variant_id));

            const shipFee = clientItem ? Number(clientItem.shipping_price || 0) : 0;
            const shipDiscount = clientItem ? Number(clientItem.shipping_support_price || 0) : 0;
            const prodDiscount = clientItem ? Number(clientItem.product_support_price || 0) : 0;
            // Thành tiền món hàng = (Giá x Số lượng) + Phí Ship - Voucher Ship - Voucher SP
            const lineTotal = (Number(item.unit_price) * item.quantity) + shipFee - shipDiscount - prodDiscount;
            return sum + lineTotal;
        }, 0);

        // ── TẠO PAYOS PAYMENT LINK ──────────────────────────────────────
        const CLIENT_URL = process.env.CLIENT_URL || 'http://localhost:5173';
        const orderCode = Math.floor(100000 + Math.random() * 900000000); // số nguyên dương

        // Lưu orderCode vào reservation để webhook tra cứu được
        await sql`
            UPDATE inventory_reservations
            SET payos_order_code = ${orderCode}
            WHERE reservation_id = ${reservation.reservation_id}
        `;

        if(totalAmount===0){
            return res.status(200).json({
                success: true,
                isFree: true,
                message: "Đơn hàng miễn phí, vui lòng xác nhận đơn hàng trực tiếp."
            });
        }

        const payosBody = {
            orderCode,
            amount: totalAmount,
            description: reservation.payment_code,
            items: reservation.items.map(i => ({
                name: i.product_name.substring(0, 50),
                quantity: i.quantity,
                price: Number(i.unit_price),
            })),
            returnUrl: `${CLIENT_URL}/checkout/payos-return`,
            cancelUrl: `${CLIENT_URL}/checkout/payos-cancel`,
        };

        const payosResponse = await payos.paymentRequests.create(payosBody);
        console.log(payosResponse);

        return res.status(200).json({
            success: true,
            message: `Giữ hàng thành công! Vui lòng thanh toán trong 15 phút.`,
            reservation_id: reservation.reservation_id,
            account_number: payosResponse.accountNumber,
            account_name: payosResponse.accountName,
            payos_checkout_url: payosResponse.checkoutUrl,
            qr_url: payosResponse.qrCode,
            payment_code: reservation.payment_code,
            total_amount: totalAmount,
            expires_at: reservation.expires_at,
        });
    } catch (error) {
        console.error('[Reserve Error]:', error.message);
        return res.status(400).json({ success: false, error: error.message });
    }
});

// ════════════════════════════════════════════════════════════════════════════════
// PHASE 2: HOÀN TẤT THANH TOÁN (CHECKOUT)
// ════════════════════════════════════════════════════════════════════════════════
router.post(
    '/checkout',
    validateBody(['user_id', 'reservation_id', 'address_id', 'payment_method']),
    async (req, res) => {
        const { user_id, reservation_id, address_id, payment_method, note, voucherCode, voucher_code } = req.body;

        try {
            const result = await confirmCheckout({
                userId: user_id,
                reservationId: reservation_id,
                addressId: address_id,
                paymentMethod: payment_method,
                note,
                voucherCode: voucherCode || voucher_code
            });

            return res.status(201).json({
                success: true,
                message: 'Đặt hàng thành công! Cửa hàng đang chuẩn bị hàng gửi cho bạn.',
                order_id: result.order_id,
                total_amount: result.total_amount,
                status_id: result.status_id,
            });
        } catch (error) {
            console.error('[Checkout] Error:', error.message);
            const isReservationExpired = error.message.includes('hết hạn');
            const statusCode = isReservationExpired ? 410 : 400;
            return res.status(statusCode).json({ success: false, error: error.message });
        }
    }
);

router.post(
    '/checkout-free-order',
    validateBody(['user_id', 'items', 'address_id', 'payment_method']),
    async (req, res) => {
        const { user_id, items, address_id, payment_method, note } = req.body;

        try {
            // 1. Tính toán lại phải dùng Number() để tránh lỗi nối chuỗi
            const totalCheck = items.reduce((sum, item) => {
                const line = (Number(item.unit_price || 0) * item.quantity)
                    + Number(item.shipping_price || 0)
                    - Number(item.shipping_support_price || 0)
                    - Number(item.product_support_price || 0);
                return sum + line;
            }, 0);

            if (totalCheck > 0) {
                return res.status(400).json({
                    success: false,
                    error: `Đơn hàng vẫn còn ${totalCheck.toLocaleString()}đ, không thể thanh toán miễn phí!`
                });
            }

            if (!Array.isArray(items) || items.length === 0) {
                return res.status(400).json({ success: false, error: 'Danh sách sản phẩm không hợp lệ.' });
            }

            // 2. Gọi hàm service
            const result = await checkoutFreeOrder({
                userId: user_id,
                items: items,
                addressId: address_id,
                paymentMethod: payment_method,
                note: note || 'Thanh toán đơn hàng 0đ',
            });

            return res.status(201).json({
                success: true,
                message: 'Đặt hàng thành công!',
                order_id: result.order_id,
                total_amount: result.total_amount,
                status_id: result.status_id,
            });
        } catch (error) {
            console.error('[Checkout Free Order] Error:', error.message);
            return res.status(400).json({ success: false, error: error.message });
        }
    }
);

// ════════════════════════════════════════════════════════════════════════════════
// PHASE 2.5: ĐẶT HÀNG TRỰC TIẾP (COD BỎ QUA GIỮ HÀNG)
// ════════════════════════════════════════════════════════════════════════════════
router.post(
    '/checkout-direct',
    validateBody(['user_id', 'items', 'address_id', 'payment_method']),
    async (req, res) => {
        const { user_id, items, address_id, payment_method, note, voucherCode, voucher_code } = req.body;

        if (!Array.isArray(items) || items.length === 0) {
            return res.status(400).json({ success: false, error: 'Danh sách sản phẩm không hợp lệ.' });
        }

        try {
            const result = await checkoutDirect({
                userId: user_id,
                items: items,
                addressId: address_id,
                paymentMethod: payment_method,
                note,
                voucherCode: voucherCode || voucher_code
            });

            return res.status(201).json({
                success: true,
                message: 'Đặt hàng COD thành công! Đơn hàng đang chờ cửa hàng xác nhận.',
                order_id: result.order_id,
                total_amount: result.total_amount,
                status_id: result.status_id,
            });
        } catch (error) {
            console.error('[Checkout Direct] Error:', error.message);
            return res.status(400).json({ success: false, error: error.message });
        }
    }
);

// ════════════════════════════════════════════════════════════════════════════════
// HỦY RESERVATION (User rời khỏi trang Checkout)
// ════════════════════════════════════════════════════════════════════════════════
router.delete('/reserve/:userId', async (req, res) => {
    const { userId } = req.params;

    try {
        const cancelled = await cancelReservation(parseInt(userId));
        if (!cancelled) {
            return res.status(404).json({ success: false, message: 'Không tìm thấy phiên giữ hàng đang hoạt động.' });
        }
        return res.json({ success: true, message: 'Đã hủy giữ hàng và trả về kho.' });
    } catch (error) {
        console.error('[Cancel Reserve] Error:', error.message);
        return res.status(500).json({ success: false, error: 'Lỗi khi hủy giữ hàng.' });
    }
});

// ════════════════════════════════════════════════════════════════════════════════
// ADMIN: CẬP NHẬT TRẠNG THÁI ĐƠN HÀNG (Có hoàn/cắt tiền vốn khi xử lý Hủy Đơn)
// ════════════════════════════════════════════════════════════════════════════════
router.patch('/status/:order_id', async (req, res) => {
    const { order_id } = req.params;
    const { status_id, note, admin_id } = req.body;

    if (!status_id) {
        return res.status(400).json({ success: false, error: 'Thiếu status_id.' });
    }

    try {
        await sql.begin(async (trx) => {
            const [oldOrder] = await trx`
                SELECT status_id FROM orders WHERE order_id = ${order_id} FOR UPDATE
            `;

            if (!oldOrder) throw new Error("Không tìm thấy đơn hàng");

            const oldStatus = Number(oldOrder.status_id);
            const newStatus = Number(status_id);

            if (oldStatus === newStatus) return;

            await trx`UPDATE orders SET status_id = ${newStatus} WHERE order_id = ${order_id}`;

            await trx`
                INSERT INTO order_status_history (order_id, status_id, changed_by, note, change_date)
                VALUES (${order_id}, ${newStatus}, ${admin_id || null}, ${note || 'Cập nhật bởi hệ thống quản trị'}, CURRENT_TIMESTAMP)
            `;

            // 6 = Hủy Đơn; 5 = Trả hàng/Hoàn tiền
            const STATUSES_THAT_RESTORE_STOCK = [5, 6];

            const wasRestored = STATUSES_THAT_RESTORE_STOCK.includes(oldStatus);
            const isRestoring = STATUSES_THAT_RESTORE_STOCK.includes(newStatus);

            // TRƯỜNG HỢP 1: Từ Đơn sống chuyển sang Hủy -> TRẢ HÀNG & HOÀN TIỀN VỐN LẠI CHO KHO
            if (!wasRestored && isRestoring) {
                const items = await trx`
                    SELECT variant_id, quantity, total_cost 
                    FROM order_details 
                    WHERE order_id = ${order_id}
                `;
                for (const item of items) {
                    await trx`
                        UPDATE productinventory
                        SET quantity     = quantity + ${item.quantity},
                            total_value  = COALESCE(total_value, 0) + COALESCE(${item.total_cost}, 0),
                            last_updated = CURRENT_TIMESTAMP
                        WHERE variant_id = ${item.variant_id}
                    `;
                }
            }
            // TRƯỜNG HỢP 2: Chuyển từ Đã Hủy về Đơn Sống -> ADMIN KHÔNG HỦY NỮA, LẤY LẠI HÀNG & TRỪ LẠI TIỀN VỐN TỪ KHO
            else if (wasRestored && !isRestoring) {
                const items = await trx`
                    SELECT variant_id, quantity, total_cost 
                    FROM order_details 
                    WHERE order_id = ${order_id}
                `;
                for (const item of items) {
                    await trx`
                        UPDATE productinventory
                        SET quantity     = quantity - ${item.quantity},
                            total_value  = total_value - COALESCE(${item.total_cost}, 0),
                            last_updated = CURRENT_TIMESTAMP
                        WHERE variant_id = ${item.variant_id}
                    `;
                }
            }
        });

        return res.json({ success: true, message: 'Cập nhật trạng thái thành công' });
    } catch (error) {
        console.error('[UpdateStatus] Error:', error.message);
        return res.status(500).json({ success: false, error: 'Lỗi khi cập nhật trạng thái' });
    }
});

// ════════════════════════════════════════════════════════════════════════════════
// TRACKING ĐƠN HÀNG
// ════════════════════════════════════════════════════════════════════════════════
router.get('/tracking/:order_id', async (req, res) => {
    const { order_id } = req.params;

    try {
        const history = await sql`
            SELECT
                os.status_name,
                osh.change_date,
                osh.note,
                u.full_name AS actor_name
            FROM order_status_history osh
                     JOIN order_status os ON osh.status_id = os.status_id
                     LEFT JOIN users u ON osh.changed_by = u.id
            WHERE osh.order_id = ${order_id}
            ORDER BY osh.change_date ASC
        `;

        if (history.length === 0) {
            return res.status(404).json({ success: false, message: 'Không tìm thấy lịch sử đơn hàng' });
        }

        return res.json({ success: true, data: history });
    } catch (error) {
        console.error('[Tracking] Error:', error.message);
        return res.status(500).json({ success: false, error: 'Lỗi lấy thông tin tracking' });
    }
});

// ════════════════════════════════════════════════════════════════════════════════
// LỊCH SỬ MUA HÀNG CỦA USER
// ════════════════════════════════════════════════════════════════════════════════
router.get('/purchase-history/:userId', async (req, res) => {
    const { userId } = req.params;

    try {
        const orders = await sql`
            SELECT
                o.order_id,
                o.order_date,
                o.total_amount,
                o.payment_method,
                o.payment_status,
                o.note AS order_note,
                os.status_name,
                o.status_id,
                o.reciever_name AS receiver_name,
                o.reciever_numphone AS receiver_phone,
                o.reciever_address AS detail_address,
                EXISTS (
                    SELECT 1 FROM reviews r WHERE r.order_id = o.order_id
                ) AS is_reviewed
            FROM orders o
                     JOIN order_status os ON o.status_id = os.status_id
            WHERE o.user_id = ${userId}
            ORDER BY o.order_date DESC
        `;

        const ordersWithDetails = await Promise.all(
            orders.map(async (order) => {
                const items = await sql`
                    SELECT
                        od.order_detail_id,
                        od.quantity,
                        od.unit_price,
                        od.shipping_price,
                        od.shipping_support_price,
                        od.product_support_price,
                        p.product_id,
                        p.category_id,
                        p.name AS product_name,
                        p.img_thumb,
                        pv.variant_id,
                        pv.color_name,
                        pv.color_code
                    FROM order_details od
                             JOIN product p ON od.product_id = p.product_id
                             LEFT JOIN productvariant pv ON od.variant_id = pv.variant_id
                    WHERE od.order_id = ${order.order_id}
                `;

                const itemsWithUrl = items.map(({ product_id, category_id, img_thumb, ...rest }) => ({
                    ...rest,
                    product_id,
                    img_thumb: `/${category_id}/${product_id}/${img_thumb}`,
                }));

                return { ...order, items: itemsWithUrl };
            })
        );

        return res.json({ success: true, data: ordersWithDetails });
    } catch (error) {
        console.error('[PurchaseHistory] Error:', error.message);
        return res.status(500).json({
            success: false,
            message: 'Không thể lấy lịch sử mua hàng',
            error: error.message,
        });
    }
});

// ════════════════════════════════════════════════════════════════════════════════
// ADMIN: LẤY TẤT CẢ ĐƠN HÀNG (PHÂN TRANG)
// ════════════════════════════════════════════════════════════════════════════════
router.get('/all', async (req, res) => {
    const page = parseInt(req.query.page) || 1;
    const limit = 7;
    const offset = (page - 1) * limit;

    try {
        const [{ count }] = await sql`SELECT COUNT(*) FROM orders`;
        const totalPages = Math.ceil(parseInt(count) / limit);

        const orders = await sql`
            SELECT
                o.order_id,
                o.order_date,
                o.total_amount,
                o.payment_method,
                o.payment_status,
                o.status_id,
                os.status_name,
                u.full_name AS customer_name,
                o.reciever_name AS receiver_name,
                o.reciever_numphone AS customer_phone,
                o.reciever_address AS customer_address,
                o.sender_name AS shop_name,
                o.sender_address AS shop_address,
                o.sender_numphone AS shop_phone
            FROM orders o
                     JOIN order_status os ON o.status_id = os.status_id
                     JOIN users u ON o.user_id = u.id
            ORDER BY o.order_date DESC
                LIMIT ${limit}
            OFFSET ${offset}
        `;

        const ordersWithDetails = await Promise.all(
            orders.map(async (order) => {
                const items = await sql`
                    SELECT
                        od.quantity,
                        od.unit_price,
                        od.shipping_price,
                        od.shipping_support_price,
                        od.product_support_price,
                        p.name AS product_name,
                        p.img_thumb,
                        p.product_id,
                        p.category_id
                    FROM order_details od
                             JOIN product p ON od.product_id = p.product_id
                    WHERE od.order_id = ${order.order_id}
                `;

                const itemsWithUrl = items.map(({ product_id, category_id, img_thumb, ...rest }) => ({
                    ...rest,
                    img_thumb: `/${category_id}/${product_id}/${img_thumb}`,
                }));

                return { ...order, items: itemsWithUrl };
            })
        );

        return res.json({
            success: true,
            pagination: {
                total_items: parseInt(count),
                total_pages: totalPages,
                current_page: page,
                limit,
            },
            data: ordersWithDetails,
        });
    } catch (error) {
        console.error('[OrderAll] Pagination Error:', error.message);
        return res.status(500).json({
            success: false,
            message: 'Lỗi phân trang đơn hàng',
            error: error.message,
        });
    }
});

// ════════════════════════════════════════════════════════════════════════════════
// PAYOS WEBHOOK – Xác nhận thanh toán tự động
// PayOS POST đến endpoint này mỗi khi trạng thái thanh toán thay đổi
// ════════════════════════════════════════════════════════════════════════════════
router.post('/payos/webhook', express.json(), async (req, res) => {
    try {
        // 1. Xác minh chữ ký webhook từ PayOS
        // verify() trả về WebhookData: { orderCode, amount, description, code, desc, ... }
        const webhookData = await payos.webhooks.verify(req.body);

        const { orderCode, code, desc } = webhookData;
        const isPaid = code === '00';   // '00' = thanh toán thành công

        console.log(`[PayOS Webhook] orderCode=${orderCode}, code=${code}, desc=${desc}, isPaid=${isPaid}`);

        if (!isPaid) {
            // Các sự kiện khác (tạo link, link hết hạn…) – bỏ qua
            return res.json({ success: true });
        }

        // 2. Tìm reservation theo orderCode
        const [reservation] = await sql`
            SELECT reservation_id, user_id, items, status, expires_at
            FROM inventory_reservations
            WHERE payos_order_code = ${orderCode}
              AND status = 'ACTIVE'
        `;

        if (!reservation) {
            console.warn(`[PayOS Webhook] Không tìm thấy reservation cho orderCode=${orderCode}`);
            return res.json({ success: true });   // Trả 200 để PayOS không retry
        }

        // 3. Lấy address_id mặc định của user (fallback nếu chưa pass qua)
        const [defaultAddr] = await sql`
            SELECT id FROM address
            WHERE id_user = ${reservation.user_id}
            ORDER BY is_default DESC, id ASC
            LIMIT 1
        `;

        if (!defaultAddr) {
            console.error(`[PayOS Webhook] User ${reservation.user_id} không có địa chỉ.`);
            return res.json({ success: true });
        }

        // 4. Gọi confirmCheckout để tạo đơn hàng
        const result = await confirmCheckout({
            userId: reservation.user_id,
            reservationId: reservation.reservation_id,
            addressId: defaultAddr.id,
            paymentMethod: 'QR',
            note: `Thanh toán qua PayOS – Mã GD: ${orderCode}`,
        });

        console.log(`[PayOS Webhook] ✅ Đơn hàng #${result.order_id} tạo thành công!`);

        // 5. Emit Socket.io để client tự redirect sang trang thành công
        // (req.app.get('socketio') được gán trong index.js)
        const io = req.app.get('socketio');
        if (io) {
            io.emit(`payos_paid_${reservation.user_id}`, {
                order_id: result.order_id,
                total_amount: result.total_amount,
            });
        }

        return res.json({ success: true });
    } catch (err) {
        console.error('[PayOS Webhook Error]:', err.message);
        // Trả 200 để PayOS không retry vô tận
        return res.status(200).json({ success: false, error: err.message });
    }
});

// ════════════════════════════════════════════════════════════════════════════════
// PAYOS: Kiểm tra trạng thái link thanh toán (client poll)
// ════════════════════════════════════════════════════════════════════════════════
router.get('/payos/status/:orderCode', async (req, res) => {
    try {
        const { orderCode } = req.params;
        const info = await payos.paymentRequests.get(orderCode);
        return res.json({ success: true, data: info });
    } catch (err) {
        console.error('[PayOS Status Error]:', err.message);
        return res.status(500).json({ success: false, error: err.message });
    }
});

module.exports = router;