const nodemailer = require('nodemailer');
require('dotenv').config();

const transporter = nodemailer.createTransport({
    //service: 'gmail', // Hoặc sử dụng host/port của dịch vụ khác
    service: 'gmail',
    host: 'smtp.gmail.com',
    port: 587,
    secure: false,
    auth: {
        user: process.env.OTP_GMAIL,
        pass: process.env.OTP_PASSWORD
    },
    tls: {
        rejectUnauthorized: false
    }
});

transporter.verify((error, success) => {
    if (error) {
        console.log("Lỗi kết nối Mail Server:", error);
    } else {
        console.log("Sẵn sàng gửi Email!");
    }
});

module.exports = transporter;