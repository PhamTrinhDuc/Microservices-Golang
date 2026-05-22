const express = require('express');
const router = express.Router();
const sql = require('../db.jsx');

// 1. Get or create a chat room for a customer
router.post('/room', async (req, res) => {
    const { customer_id } = req.body;

    if (!customer_id) {
        return res.status(400).json({ success: false, message: 'Thiếu thông tin customer_id' });
    }

    try {
        // Check if room already exists
        let [room] = await sql`
            SELECT * FROM chat_rooms WHERE customer_id = ${customer_id} LIMIT 1
        `;

        if (!room) {
            // Create a new room
            [room] = await sql`
                INSERT INTO chat_rooms (customer_id)
                VALUES (${customer_id})
                RETURNING *
            `;
        }

        return res.status(200).json({ success: true, data: room });
    } catch (error) {
        console.error('Error in /room:', error);
        return res.status(500).json({ success: false, message: 'Lỗi máy chủ', error: error.message });
    }
});

// 2. Admin: Get all active chat rooms
router.get('/rooms', async (req, res) => {
    try {
        const rooms = await sql`
            SELECT 
                r.room_id, 
                r.customer_id,
                r.created_at AT TIME ZONE 'UTC' AT TIME ZONE 'Asia/Ho_Chi_Minh' AS created_at,
                r.updated_at AT TIME ZONE 'UTC' AT TIME ZONE 'Asia/Ho_Chi_Minh' AS updated_at,
                u.full_name AS customer_name, 
                u.email AS customer_email,
                (
                    SELECT message_text 
                    FROM chat_messages 
                    WHERE room_id = r.room_id 
                    ORDER BY created_at DESC 
                    LIMIT 1
                ) AS last_message,
                (
                    SELECT created_at 
                    FROM chat_messages 
                    WHERE room_id = r.room_id 
                    ORDER BY created_at DESC 
                    LIMIT 1
                ) AS last_message_time,
                (
                    SELECT COUNT(*)::int 
                    FROM chat_messages 
                    WHERE room_id = r.room_id AND is_read = false AND sender_role = 'customer'
                ) AS unread_count
            FROM chat_rooms r
            JOIN users u ON r.customer_id = u.id
            ORDER BY r.updated_at DESC
        `;

        return res.json({ success: true, data: rooms });
    } catch (error) {
        console.error('Error in /rooms:', error);
        return res.status(500).json({ success: false, message: 'Lỗi máy chủ', error: error.message });
    }
});

// 3. Get message history for a specific room
router.get('/messages/:roomId', async (req, res) => {
    const { roomId } = req.params;

    try {
        const messages = await sql`
            SELECT
                message_id,
                room_id,
                sender_id,
                sender_role,
                message_text,
                is_read,
                created_at AT TIME ZONE 'UTC' AT TIME ZONE 'Asia/Ho_Chi_Minh' AS created_at
            FROM chat_messages
            WHERE room_id = ${roomId}
            ORDER BY created_at ASC
            
        `;

        return res.json({ success: true, data: messages });
    } catch (error) {
        console.error('Error fetching messages:', error);
        return res.status(500).json({ success: false, message: 'Lỗi máy chủ', error: error.message });
    }
});

// 4. Mark all messages in a room as read
router.patch('/messages/:roomId/read', async (req, res) => {
    const { roomId } = req.params;
    const { role } = req.body; // 'customer' or 'admin'

    if (!role) {
        return res.status(400).json({ success: false, message: 'Thiếu thông tin vai trò (role)' });
    }

    try {
        // If customer read, mark admin messages as read. If admin read, mark customer messages as read.
        const targetRole = role === 'admin' ? 'customer' : 'admin';

        await sql`
            UPDATE chat_messages 
            SET is_read = true 
            WHERE room_id = ${roomId} AND sender_role = ${targetRole}
        `;

        return res.json({ success: true, message: 'Đã đánh dấu đã đọc thành công!' });
    } catch (error) {
        console.error('Error marking messages as read:', error);
        return res.status(500).json({ success: false, message: 'Lỗi máy chủ', error: error.message });
    }
});

module.exports = router;
