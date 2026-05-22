const sql = require("./db.jsx"); // Đường dẫn tới file db của bạn
const bcrypt = require('bcrypt');

async function migrate() {
    try {
        console.log("🚀 Bắt đầu quá trình mã hóa mật khẩu...");

        // 1. Lấy 6 bản ghi đầu tiên (ID từ 1 đến 6)
        const users = await sql`
            SELECT id, password FROM users 
            WHERE id IN (1, 2, 3, 4, 5, 6)
        `;

        for (let user of users) {
            // Kiểm tra xem mật khẩu đã được mã hóa chưa (bcrypt luôn bắt đầu bằng $2b$)
            if (user.password.startsWith('$2b$')) {
                console.log(`- User ID ${user.id}: Đã mã hóa trước đó, bỏ qua.`);
                continue;
            }

            // 2. Tiến hành mã hóa mật khẩu cũ
            const salt = await bcrypt.genSalt(10);
            const hashedPassword = await bcrypt.hash(user.password, salt);

            // 3. Update ngược lại vào Database
            await sql`
                UPDATE users 
                SET password = ${hashedPassword} 
                WHERE id = ${user.id}
            `;
            console.log(`✅ Đã mã hóa xong cho User ID: ${user.id}`);
        }

        console.log("✨ Hoàn thành cập nhật 6 bản ghi!");
        process.exit();

    } catch (err) {
        console.error("❌ Lỗi khi migrate:", err.message);
        process.exit(1);
    }
}

migrate();