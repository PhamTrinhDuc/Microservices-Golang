const express = require("express");
const redis = require('redis');
const bcrypt = require('bcrypt');
const transporter = require("./Transporter.jsx");
const getOTPTemplate = require("./emailTemplate.js");
const crypto = require('crypto');
require('dotenv').config();
const router = express.Router();

// --- KẾT NỐI REDIS CLOUD (Dán URL Upstash vào đây) ---
const REDIS_URL = process.env.REDIS_URL;

const client = redis.createClient({
    url: REDIS_URL
});

client.on('error', err => console.log('Redis Client Error', err));
client.connect().then(() => console.log("Đã kết nối Redis Cloud thành công!"));

const OTP_EXPIRY = 300;
const LIMIT_TIME = 60;

// API Gửi OTP
// router.post('/send-otp', async (req, res) => {
//     try {
//         const { email } = req.body;
//         if (!email) return res.status(400).json({ message: "Vui lòng nhập email" });
//
//         const rateKey = `limit:${email}`;
//         const requestCount = await client.get(rateKey);
//
//         if (requestCount && parseInt(requestCount) >= 3) {
//             return res.status(429).json({ message: "Thao tác quá nhanh. Thử lại sau 1 phút." });
//         }
//
//         const otp = Math.floor(100000 + Math.random() * 900000).toString();
//         const hashedOtp = await bcrypt.hash(otp, 10);
//         await client.setEx(`otp:${email}`, OTP_EXPIRY, hashedOtp);
//
//         await transporter.sendMail({
//             from: '"Hệ thống CellphoneX" <your-email@gmail.com>',
//             to: email,
//             subject: `[OTP] ${otp} là mã xác nhận của bạn`,
//             html: getOTPTemplate(otp)
//         });
//
//         if (!requestCount) {
//             await client.setEx(rateKey, LIMIT_TIME, "1");
//         } else {
//             await client.incr(rateKey);
//         }
//
//         console.log(`OTP của ${email} là: ${otp}`);
//         res.status(200).json({ message: "Mã OTP đã được gửi về email của bạn!" });
//
//     } catch (error) {
//         console.error("Lỗi:", error);
//         res.status(500).json({ message: "Lỗi hệ thống, vui lòng thử lại sau." });
//     }
// });

router.post('/send-otp', async (req, res) => {
    try {
        const { email } = req.body;
        if (!email) return res.status(400).json({ message: "Vui lòng nhập email" });

        const otp = Math.floor(100000 + Math.random() * 900000).toString();
        const hashedOtp = await bcrypt.hash(otp, 10);

        // Lưu OTP vào Redis (5 phút)
        await client.setEx(`otp:${email}`, 300, hashedOtp);

        await transporter.sendMail({
            from: '"Hệ thống CellphoneX" <your-email@gmail.com>',
            to: email,
            subject: `[OTP] ${otp} là mã xác nhận của bạn`,
            html: getOTPTemplate(otp)
        });

        res.status(200).json({ message: "Mã OTP đã được gửi!" });
    } catch (error) {
        res.status(500).json({ message: "Lỗi gửi OTP" });
    }
});

// router.post('/verify-otp', async (req, res) => {
//     try {
//         const { email, otp } = req.body;
//         if (!email || !otp) return res.status(400).json({ message: "Thiếu email hoặc OTP" });
//
//         const storedHash = await client.get(`otp:${email}`);
//         if (!storedHash) return res.status(400).json({ message: "Mã OTP đã hết hạn hoặc không tồn tại" });
//
//         const isMatch = await bcrypt.compare(otp, storedHash);
//         if (isMatch) {
//             await client.del(`otp:${email}`); // Xóa mã sau khi dùng
//             res.status(200).json({ success: true, message: "Xác thực thành công!" });
//         } else {
//             res.status(400).json({ success: false, message: "Mã OTP không chính xác" });
//         }
//     } catch (error) {
//         res.status(500).json({ message: "Lỗi xử lý xác thực" });
//     }
// });

router.post('/verify-otp', async (req, res) => {
    try {
        const { email, otp } = req.body;
        const storedHash = await client.get(`otp:${email}`);

        if (!storedHash) return res.status(400).json({ message: "Mã OTP hết hạn" });

        const isMatch = await bcrypt.compare(otp, storedHash);
        if (isMatch) {
            const resetToken = crypto.randomBytes(32).toString('hex');
            await client.setEx(`resetToken:${email}`, 600, resetToken);
            await client.del(`otp:${email}`);

            res.status(200).json({
                success: true,
                message: "Xác thực thành công!",
                resetToken: resetToken
            });
        } else {
            res.status(400).json({ message: "Mã OTP không chính xác" });
        }
    } catch (error) {
        res.status(500).json({ message: "Lỗi xác thực" });
    }
});

module.exports = router;