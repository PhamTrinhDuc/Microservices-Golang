const express = require('express');
const router = express.Router();
const { OAuth2Client } = require('google-auth-library');
const sql = require("../db.jsx");
const bcrypt = require('bcrypt');

const clientGoogle = new OAuth2Client(process.env.GOOGLE_CLIENT_ID);

router.post('/login', async (req, res) => {
    const { idToken } = req.body;
    try {
        const ticket = await clientGoogle.verifyIdToken({
            idToken,
            audience: process.env.GOOGLE_CLIENT_ID,
        });

        const payload = ticket.getPayload();
        const { email, name, sub } = payload; // 'sub' là ID duy nhất từ Google

        let users = await sql`SELECT * FROM users WHERE email = ${email} LIMIT 1`;

        if (users.length === 0) {
            // Mã hóa 'sub' để đồng nhất với logic bcrypt của hệ thống
            const salt = await bcrypt.genSalt(10);
            const hashedSub = await bcrypt.hash(sub, salt);

            const newUser = await sql`
                INSERT INTO users (full_name, email, password, role, is_lock, joined_date)
                VALUES (${name}, ${email}, ${hashedSub}, 'Customer', FALSE, CURRENT_DATE)
                RETURNING *
            `;
            users = newUser;
        }

        const user = users[0];
        if (user.is_lock) {
            return res.status(403).json({ message: "Tài khoản Google này đã bị khóa!" });
        }

        const { password: _, ...userData } = user;
        res.status(200).json({ message: "Đăng nhập Google thành công!", user: userData });

    } catch (error) {
        console.error("Lỗi Google Auth:", error.message);
        res.status(401).json({ message: "Xác thực Google thất bại!" });
    }
});

module.exports = router;