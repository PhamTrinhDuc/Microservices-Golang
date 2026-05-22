const sql = require('../db.jsx');
const { v4: uuidv4 } = require('uuid');
const { generatePaymentCode } = require('./generateQR.jsx');
const RESERVATION_TTL_MINUTES = 15;

// ─── PHASE 1: GIỮ HÀNG ───────────────────────────────────────────────────────
async function reserveInventory(userId, items){
    return await sql.begin(async (trx)=>{
        const oldReservation = await trx`
            SELECT reservation_id, items
            FROM inventory_reservations
            WHERE user_id= ${userId}
              AND status = 'ACTIVE'
                FOR UPDATE
        `;
        for (const old of oldReservation){
            const parseItems = typeof old.items === 'string' ? JSON.parse(old.items) : old.items;

            for(const oldItem of parseItems){
                await trx`
                    UPDATE productinventory
                    SET reserved = GREATEST(0, reserved-${oldItem.quantity}),
                        last_updated = CURRENT_TIMESTAMP
                    WHERE variant_id = ${oldItem.variant_id}
                `;
                await trx`
                    UPDATE inventory_reservations
                    SET status = 'SUPERSEDED'
                    WHERE reservation_id = ${old.reservation_id}
                `;
            }
        }

        const itemsWithPrice = [];
        for(const item of items){
            const [inv] = await trx`
                SELECT
                    pi.variant_id,
                    pi.quantity,
                    pi.reserved,
                    pv.price_str,
                    p.name AS product_name,
                    p.product_id,
                    pv.price AS original_price,
                    COALESCE(
                            pv.price - (pv.price * pr.discount_percent / 100),
                            pv.price
                    ) AS price
                FROM productinventory pi
                         JOIN productvariant pv ON pi.variant_id = pv.variant_id
                         JOIN product p ON pv.product_id = p.product_id
                         LEFT JOIN promotions pr ON p.product_id = pr.product_id
                    AND pr.is_deleted = false
                    AND CURRENT_TIMESTAMP >= pr.start_date
                    AND CURRENT_TIMESTAMP <= pr.end_date
                WHERE pi.variant_id = ${item.variant_id}
                    FOR UPDATE OF pi 
            `;
            if(!inv){
                throw new Error(`Biến thể sản phẩm ID ${item.variant_id} không tồn tại.`);
            }

            const available = inv.quantity - inv.reserved;
            if (available < item.quantity) {
                throw new Error(
                    `"${inv.product_name}" không đủ hàng. Yêu cầu: ${item.quantity}, Khả dụng: ${Math.max(0, available)}.`
                );
            }

            const [updated] = await trx`
                UPDATE productinventory
                SET reserved     = reserved + ${item.quantity},
                    last_updated = CURRENT_TIMESTAMP
                WHERE variant_id = ${item.variant_id}
                  AND (quantity - reserved) >= ${item.quantity}
                    RETURNING quantity, reserved
            `;
            if (!updated) {
                throw new Error(`"${inv.product_name}" vừa hết hàng. Vui lòng thử lại.`);
            }

            itemsWithPrice.push({
                variant_id: item.variant_id,
                product_id: inv.product_id,
                product_name: inv.product_name,
                quantity: item.quantity,
                unit_price: inv.price,
                price_str: inv.price_str,
                shipping_price: item.shipping_price || 0,
                shipping_support_price: item.shipping_support_price || 0,
                product_support_price: item.product_support_price || 0
            });
        }

        // --- CƠ CHẾ CHỐNG TRÙNG MÃ THANH TOÁN ---
        let paymentCode;
        let isUnique = false;
        while (!isUnique) {
            paymentCode = generatePaymentCode();
            const [existing] = await trx`SELECT 1 FROM inventory_reservations WHERE payment_code = ${paymentCode}`;
            if (!existing) isUnique = true;
        }

        const reservationId = uuidv4();
        const expiresAt = new Date(Date.now() + 15 * 60 * 1000);

        // Lưu vào DB kèm payment_code
        await trx`
            INSERT INTO inventory_reservations
                (reservation_id, user_id, items, status, expires_at, payment_code)
            VALUES
                (${reservationId}, ${userId}, ${sql.json(itemsWithPrice)}, 'ACTIVE', ${expiresAt}, ${paymentCode})
        `;

        return {
            reservation_id: reservationId,
            payment_code: paymentCode,
            expires_at: expiresAt.toISOString(),
            items: itemsWithPrice
        };
    });
}

// ─── UTILS ───────────────────────────────────────────────────────────────────
function resolvePaymentMeta(paymentMethod) {
    const isCOD = paymentMethod.toUpperCase() === 'COD';
    return {
        initialStatusId: isCOD ? 1 : 2,
        historyNote: isCOD
            ? 'Đơn hàng COD đang chờ xác nhận từ cửa hàng'
            : 'Thanh toán thành công, cửa hàng đang chuẩn bị hàng'
    };
}

async function getOrderAddresses(trx, addressId) {
    const [addr] = await trx`SELECT * FROM address WHERE id = ${addressId}`;
    if (!addr) throw new Error('Địa chỉ giao hàng không hợp lệ.');

    const [store] = await trx`SELECT * FROM store LIMIT 1`;

    return {
        recieverName: addr.full_name,
        recieverNumphone: addr.num_phone,
        recieverAddress: [addr.detail_address, addr.ward, addr.province].filter(Boolean).join(', '),

        senderName: store ? store.name : 'Hệ thống Cửa hàng',
        senderNumphone: store ? store.hotline : '',
        senderAddress: store ? [store.road, store.ward, store.province].filter(Boolean).join(', ') : ''
    };
}

// ─── PHASE 2: HOÀN TẤT THANH TOÁN (Cập nhật Bình quân gia quyền) ──────────────
async function confirmCheckout({ userId, reservationId, addressId, paymentMethod, note, voucherCode }) {
    const [reservation] = await sql`
        SELECT
            reservation_id,
            items,
            status,
            (expires_at < CURRENT_TIMESTAMP) as is_expired
        FROM inventory_reservations
        WHERE reservation_id = ${reservationId}
          AND user_id = ${userId}
    `;

    if (!reservation) throw new Error('Phiên giữ hàng không hợp lệ hoặc không tồn tại.');
    if (reservation.status !== 'ACTIVE') throw new Error(`Phiên giữ hàng đã "${reservation.status}".`);
    if (reservation.is_expired) throw new Error('Phiên giữ hàng đã hết hạn. Vui lòng quay lại giỏ hàng.');

    const { initialStatusId, historyNote } = resolvePaymentMeta(paymentMethod);
    let orderId, totalAmount;

    await sql.begin(async (trx) => {
        const [lockedRes] = await trx`
            SELECT status FROM inventory_reservations WHERE reservation_id = ${reservationId} FOR UPDATE
        `;
        if (lockedRes.status !== 'ACTIVE') throw new Error('Đơn hàng này đã được xử lý.');

        const {
            recieverName, recieverNumphone, recieverAddress,
            senderName, senderNumphone, senderAddress
        } = await getOrderAddresses(trx, addressId);

        if (typeof reservation.items === 'string') {
            reservation.items = JSON.parse(reservation.items);
        }

        totalAmount = reservation.items.reduce((sum, item) => {
            const itemTotalSales = Number(BigInt(item.unit_price) * BigInt(item.quantity));
            const shipping = Number(item.shipping_price || 0);
            const shipDiscount = Number(item.shipping_support_price || 0);
            const prodDiscount = Number(item.product_support_price || 0);

            return sum + (itemTotalSales + shipping - shipDiscount - prodDiscount);
        }, 0);

        // Lấy tổng tiền giảm giá từ Frontend truyền lên để lưu vào Database
        const voucherDiscount = reservation.items.reduce((sum, item) => {
            return sum + Number(item.product_support_price || 0) + Number(item.shipping_support_price || 0);
        }, 0);

        const [order] = await trx`
            INSERT INTO orders (
                user_id, total_amount, payment_method, status_id, note,
                sender_name, sender_address, sender_numphone,
                reciever_name, reciever_address, reciever_numphone,
                voucher_code, voucher_discount
            ) VALUES (
                         ${userId}, ${totalAmount}, ${paymentMethod.toUpperCase()}, ${initialStatusId}, ${note || null},
                         ${senderName}, ${senderAddress}, ${senderNumphone},
                         ${recieverName}, ${recieverAddress}, ${recieverNumphone},
                         ${voucherCode || null}, ${voucherDiscount}
                     )
                RETURNING order_id
        `;
        orderId = order.order_id;

        for (const item of reservation.items) {
            // 1. Lấy thông tin kho & cục tiền hiện tại
            const [inv] = await trx`
                SELECT quantity, total_value 
                FROM productinventory 
                WHERE variant_id = ${item.variant_id} 
                FOR UPDATE
            `;
            if (!inv) throw new Error(`Lỗi tồn kho với biến thể ${item.variant_id}.`);

            const currentQty = Number(inv.quantity);
            const currentTotalValue = Number(inv.total_value || 0);
            const qtySold = Number(item.quantity);

            // 2. TÍNH GIÁ VỐN CHO DÒNG NÀY (Cắt cục tiền)
            let itemTotalCost = 0;
            if (qtySold >= currentQty) {
                itemTotalCost = currentTotalValue; // Vét sạch
            } else if (currentQty > 0) {
                itemTotalCost = Math.round((currentTotalValue / currentQty) * qtySold);
            }

            // 3. Insert order_details kèm total_cost
            await trx`
                INSERT INTO order_details (
                    order_id, product_id, variant_id, quantity, unit_price,
                    shipping_price, shipping_support_price, product_support_price, total_cost
                )
                VALUES (
                           ${orderId}, ${item.product_id}, ${item.variant_id}, ${qtySold}, ${item.unit_price},
                           ${item.shipping_price || 0}, ${item.shipping_support_price || 0}, ${item.product_support_price || 0}, ${itemTotalCost}
                       )
            `;

            // 4. Trừ kho và trừ tiền vốn
            const [result] = await trx`
                UPDATE productinventory
                SET quantity     = quantity - ${qtySold},
                    reserved     = reserved - ${qtySold},
                    total_value  = total_value - ${itemTotalCost},
                    last_updated = CURRENT_TIMESTAMP
                WHERE variant_id = ${item.variant_id}
                  AND quantity >= ${qtySold}
                RETURNING quantity
            `;
            if (!result) throw new Error(`Kho hàng biến động với biến thể ${item.variant_id}. Giao dịch bị hủy.`);
        }

        await trx`
            INSERT INTO order_status_history (order_id, status_id, note, change_date)
            VALUES (${orderId}, ${initialStatusId}, ${historyNote}, CURRENT_TIMESTAMP)
        `;

        await trx`
            UPDATE inventory_reservations SET status = 'CONFIRMED' WHERE reservation_id = ${reservationId}
        `;
    });

    return { order_id: orderId, total_amount: totalAmount, status_id: initialStatusId };
}

// ─── THANH TOÁN ĐƠN 0đ ────────────────
async function checkoutFreeOrder({ userId, items, addressId, paymentMethod, note }) {
    const sortedItems = [...items].sort((a, b) => a.variant_id - b.variant_id);
    const { initialStatusId, historyNote } = resolvePaymentMeta(paymentMethod);

    let orderId, totalAmount = 0;
    const calculatedItems = [];

    await sql.begin(async (trx) => {
        const {
            recieverName, recieverNumphone, recieverAddress,
            senderName, senderNumphone, senderAddress
        } = await getOrderAddresses(trx, addressId);

        for (const item of sortedItems) {
            const [inv] = await trx`
                SELECT
                    pi.quantity, pi.reserved, pi.total_value,
                    COALESCE(pv.price - (pv.price * pr.discount_percent / 100), pv.price) AS price,
                    p.product_id, p.name AS product_name
                FROM productinventory pi
                         JOIN productvariant pv ON pi.variant_id = pv.variant_id
                         JOIN product p ON pv.product_id = p.product_id
                         LEFT JOIN promotions pr ON p.product_id = pr.product_id
                    AND pr.is_deleted = false
                    AND CURRENT_TIMESTAMP >= pr.start_date
                    AND CURRENT_TIMESTAMP <= pr.end_date
                WHERE pi.variant_id = ${item.variant_id}
                    FOR UPDATE OF pi
            `;
            const available = inv.quantity - inv.reserved;
            if (available < item.quantity) {
                throw new Error(`Sản phẩm "${inv.product_name}" không đủ hàng. Khả dụng: ${Math.max(0, available)}`);
            }

            // Tính giá vốn ngay lúc hold khóa dòng
            const currentQty = Number(inv.quantity);
            const currentTotalValue = Number(inv.total_value || 0);
            const qtySold = Number(item.quantity);

            let itemTotalCost = 0;
            if (qtySold >= currentQty) {
                itemTotalCost = currentTotalValue;
            } else if (currentQty > 0) {
                itemTotalCost = Math.round((currentTotalValue / currentQty) * qtySold);
            }

            calculatedItems.push({
                product_id: inv.product_id, variant_id: item.variant_id,
                quantity: qtySold, unit_price: inv.price,
                shipping_price: item.shipping_price || 0,
                shipping_support_price: item.shipping_support_price || 0,
                product_support_price: item.product_support_price || 0,
                total_cost: itemTotalCost
            });
            const itemTotalSales = Number(BigInt(inv.price) * BigInt(qtySold));
            const shipping = Number(item.shipping_price || 0);
            const shipDiscount = Number(item.shipping_support_price || 0);
            const prodDiscount = Number(item.product_support_price || 0);

            totalAmount += (itemTotalSales + shipping - shipDiscount - prodDiscount);
        }

        const [order] = await trx`
            INSERT INTO orders (
                user_id, total_amount, payment_method, status_id, note,
                sender_name, sender_address, sender_numphone,
                reciever_name, reciever_address, reciever_numphone
            ) VALUES (
                         ${userId}, ${totalAmount}, ${paymentMethod.toUpperCase()}, ${initialStatusId}, ${note || null},
                         ${senderName}, ${senderAddress}, ${senderNumphone},
                         ${recieverName}, ${recieverAddress}, ${recieverNumphone}
                     )
                RETURNING order_id
        `;
        orderId = order.order_id;

        for (const cItem of calculatedItems) {
            await trx`
                INSERT INTO order_details (
                    order_id, product_id, variant_id, quantity, unit_price,
                    shipping_price, shipping_support_price, product_support_price, total_cost
                )
                VALUES (
                           ${orderId}, ${cItem.product_id}, ${cItem.variant_id}, ${cItem.quantity}, ${cItem.unit_price},
                           ${cItem.shipping_price}, ${cItem.shipping_support_price}, ${cItem.product_support_price}, ${cItem.total_cost}
                       )
            `;

            const [result] = await trx`
                UPDATE productinventory
                SET quantity     = quantity - ${cItem.quantity},
                    total_value  = total_value - ${cItem.total_cost},
                    last_updated = CURRENT_TIMESTAMP
                WHERE variant_id = ${cItem.variant_id}
                  AND (quantity - reserved) >= ${cItem.quantity}
                RETURNING quantity
            `;
            if (!result) throw new Error("Kho hàng biến động quá nhanh, giao dịch thất bại.");
        }

        await trx`
            INSERT INTO order_status_history (order_id, status_id, note, change_date)
            VALUES (${orderId}, ${initialStatusId}, ${historyNote}, CURRENT_TIMESTAMP)
        `;
    });

    return { order_id: orderId, total_amount: totalAmount, status_id: initialStatusId };
}

// ─── COD/CHUYỂN KHOẢN TRỰC TIẾP (Cập nhật Bình quân gia quyền) ────────────────
async function checkoutDirect({ userId, items, addressId, paymentMethod, note, voucherCode }) {
    const sortedItems = [...items].sort((a, b) => a.variant_id - b.variant_id);
    const { initialStatusId, historyNote } = resolvePaymentMeta(paymentMethod);

    let orderId, totalAmount = 0;
    const calculatedItems = [];

    await sql.begin(async (trx) => {
        const {
            recieverName, recieverNumphone, recieverAddress,
            senderName, senderNumphone, senderAddress
        } = await getOrderAddresses(trx, addressId);

        for (const item of sortedItems) {
            const [inv] = await trx`
                SELECT
                    pi.quantity, pi.reserved, pi.total_value,
                    COALESCE(pv.price - (pv.price * pr.discount_percent / 100), pv.price) AS price,
                    p.product_id, p.name AS product_name
                FROM productinventory pi
                         JOIN productvariant pv ON pi.variant_id = pv.variant_id
                         JOIN product p ON pv.product_id = p.product_id
                         LEFT JOIN promotions pr ON p.product_id = pr.product_id
                    AND pr.is_deleted = false
                    AND CURRENT_TIMESTAMP >= pr.start_date
                    AND CURRENT_TIMESTAMP <= pr.end_date
                WHERE pi.variant_id = ${item.variant_id}
                    FOR UPDATE OF pi
            `;

            if (!inv) throw new Error(`Sản phẩm (Variant ID: ${item.variant_id}) không tồn tại.`);
            const available = inv.quantity - inv.reserved;
            if (available < item.quantity) {
                throw new Error(`Sản phẩm "${inv.product_name}" không đủ hàng. Khả dụng: ${Math.max(0, available)}`);
            }

            // Tính giá vốn ngay lúc hold khóa dòng
            const currentQty = Number(inv.quantity);
            const currentTotalValue = Number(inv.total_value || 0);
            const qtySold = Number(item.quantity);

            let itemTotalCost = 0;
            if (qtySold >= currentQty) {
                itemTotalCost = currentTotalValue;
            } else if (currentQty > 0) {
                itemTotalCost = Math.round((currentTotalValue / currentQty) * qtySold);
            }

            calculatedItems.push({
                product_id: inv.product_id, variant_id: item.variant_id,
                quantity: qtySold, unit_price: inv.price,
                shipping_price: item.shipping_price || 0,
                shipping_support_price: item.shipping_support_price || 0,
                product_support_price: item.product_support_price || 0,
                total_cost: itemTotalCost
            });
            const itemTotalSales = Number(BigInt(inv.price) * BigInt(qtySold));
            const shipping = Number(item.shipping_price || 0);
            const shipDiscount = Number(item.shipping_support_price || 0);
            const prodDiscount = Number(item.product_support_price || 0);

            totalAmount += (itemTotalSales + shipping - shipDiscount - prodDiscount);
        }

        // Lấy tổng tiền giảm giá từ Frontend truyền lên để lưu vào Database
        const voucherDiscount = calculatedItems.reduce((sum, item) => {
            return sum + Number(item.product_support_price || 0) + Number(item.shipping_support_price || 0);
        }, 0);

        const [order] = await trx`
            INSERT INTO orders (
                user_id, total_amount, payment_method, status_id, note,
                sender_name, sender_address, sender_numphone,
                reciever_name, reciever_address, reciever_numphone,
                voucher_code, voucher_discount
            ) VALUES (
                         ${userId}, ${totalAmount}, ${paymentMethod.toUpperCase()}, ${initialStatusId}, ${note || null},
                         ${senderName}, ${senderAddress}, ${senderNumphone},
                         ${recieverName}, ${recieverAddress}, ${recieverNumphone},
                         ${voucherCode || null}, ${voucherDiscount}
                     )
                RETURNING order_id
        `;
        orderId = order.order_id;

        for (const cItem of calculatedItems) {
            await trx`
                INSERT INTO order_details (
                    order_id, product_id, variant_id, quantity, unit_price,
                    shipping_price, shipping_support_price, product_support_price, total_cost
                )
                VALUES (
                           ${orderId}, ${cItem.product_id}, ${cItem.variant_id}, ${cItem.quantity}, ${cItem.unit_price},
                           ${cItem.shipping_price}, ${cItem.shipping_support_price}, ${cItem.product_support_price}, ${cItem.total_cost}
                       )
            `;

            const [result] = await trx`
                UPDATE productinventory
                SET quantity     = quantity - ${cItem.quantity},
                    total_value  = total_value - ${cItem.total_cost},
                    last_updated = CURRENT_TIMESTAMP
                WHERE variant_id = ${cItem.variant_id}
                  AND (quantity - reserved) >= ${cItem.quantity}
                RETURNING quantity
            `;
            if (!result) throw new Error("Kho hàng biến động quá nhanh, giao dịch thất bại.");
        }

        await trx`
            INSERT INTO order_status_history (order_id, status_id, note, change_date)
            VALUES (${orderId}, ${initialStatusId}, ${historyNote}, CURRENT_TIMESTAMP)
        `;
    });

    return { order_id: orderId, total_amount: totalAmount, status_id: initialStatusId };
}

// ─── HỦY RESERVATION & ROLLBACK (Giữ nguyên) ────────────────────────────────
async function cancelReservation(userId) {
    return await sql.begin(async (trx) => {
        const [reservation] = await trx`
            SELECT reservation_id, items FROM inventory_reservations
            WHERE user_id = ${userId} AND status = 'ACTIVE'
            ORDER BY created_at DESC LIMIT 1 FOR UPDATE
        `;

        if (!reservation) return false;
        if (typeof reservation.items === 'string') reservation.items = JSON.parse(reservation.items);

        for (const item of reservation.items) {
            await trx`
                UPDATE productinventory
                SET reserved     = GREATEST(0, reserved - ${item.quantity}),
                    last_updated = CURRENT_TIMESTAMP
                WHERE variant_id = ${item.variant_id}
            `;
        }
        await trx`UPDATE inventory_reservations SET status = 'CANCELLED' WHERE reservation_id = ${reservation.reservation_id}`;
        return true;
    });
}

async function rollbackExpiredReservations() {
    const expired = await sql`
        SELECT reservation_id, user_id, items
        FROM inventory_reservations
        WHERE status = 'ACTIVE' AND expires_at < CURRENT_TIMESTAMP
    `;

    if (expired.length === 0) return 0;

    for (const res of expired) {
        await sql.begin(async (trx) => {
            const [locked] = await trx`
                SELECT status FROM inventory_reservations WHERE reservation_id = ${res.reservation_id} FOR UPDATE
            `;
            if (locked.status !== 'ACTIVE') return;

            if (typeof res.items === 'string') res.items = JSON.parse(res.items);

            for (const item of res.items) {
                await trx`
                    UPDATE productinventory
                    SET reserved     = GREATEST(0, reserved - ${item.quantity}),
                        last_updated = CURRENT_TIMESTAMP
                    WHERE variant_id = ${item.variant_id}
                `;
            }

            await trx`UPDATE inventory_reservations SET status = 'EXPIRED' WHERE reservation_id = ${res.reservation_id}`;
        });
    }
    return expired.length;
}

module.exports = {
    reserveInventory,
    confirmCheckout,
    checkoutFreeOrder,
    checkoutDirect,
    cancelReservation,
    rollbackExpiredReservations,
};