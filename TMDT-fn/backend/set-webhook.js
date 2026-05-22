require('dotenv').config();
const { PayOS } = require('@payos/node');

// Lệnh khởi tạo: node set-webhook.js

// Khởi tạo instance PayOS giống trong service
const payos = new PayOS(
    process.env.PAYOS_CLIENT_ID,
    process.env.PAYOS_API_KEY,
    process.env.PAYOS_CHECKSUM_KEY
);

// THAY ĐỔI ĐƯỜNG DẪN NÀY THÀNH ĐƯỜNG DẪN NGROK CỦA BẠN (GIỮ NGUYÊN PHẦN ĐUÔI)
SERVER_URL = 'https://api.thegioibatdong.site'
const WEBHOOK_URL = `${SERVER_URL}/api/order/payos/webhook`;

async function registerWebhook() {
    try {
        console.log(`Đang đăng ký webhook URL: ${WEBHOOK_URL}...`);

        // Gọi API của PayOS để xác nhận đăng ký webhook
        const result = await payos.webhooks.confirm(WEBHOOK_URL);

        console.log('ĐĂNG KÝ WEBHOOK THÀNH CÔNG!');
        console.log('Phản hồi từ PayOS:', result);
    } catch (error) {
        console.error('CÓ LỖI XẢY RA KHI ĐĂNG KÝ WEBHOOK:');
        console.error(error.message || error);
    }
}

registerWebhook();
