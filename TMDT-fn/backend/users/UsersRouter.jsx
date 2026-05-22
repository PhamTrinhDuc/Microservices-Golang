const express = require('express');
const router = express.Router();
const sql = require("../db.jsx");
const bcrypt = require('bcrypt');
const redis = require("redis");
require('dotenv').config();

router.get('/', async (req, res) => {
    try {
        const users = await sql`SELECT * FROM users ORDER BY joined_date ASC`;
        const usersData = users.map(user => {
            const { password: _, ...userData } = user;
            return userData;
        });
        res.status(200).json(usersData);

    } catch (err) {
        console.error('Lỗi khi lấy danh sách user:', err.message);
        res.status(500).json({ message: "Lỗi máy chủ nội bộ!" });
    }
});

// router.post('/login', async (req, res) => {
//     const { username, password } = req.body;
//
//     try {
//         const users = await sql`
//             SELECT * FROM users
//             WHERE numphone = ${username} OR email = ${username}
//                 LIMIT 1
//         `;
//
//         if (users.length === 0) {
//             return res.status(401).json({ message: "Tài khoản không tồn tại!" });
//         }
//
//         const user = users[0];
//
//         if (user.password !== password) {
//             return res.status(401).json({ message: "Mật khẩu không chính xác!" });
//         }
//
//         if (user.is_lock) {
//             return res.status(403).json({
//                 message: "Tài khoản này đã bị khóa bởi Admin!"
//             });
//         }
//
//         const { password: _, ...userData } = user;
//
//         res.status(200).json({
//             message: "Đăng nhập thành công!",
//             user: userData
//         });
//
//     } catch (err) {
//         console.error('Lỗi Login:', err.message);
//         res.status(500).json({ message: "Lỗi hệ thống!" });
//     }
// });

router.post('/login', async (req, res) => {
    const { username, password } = req.body;

    try {
        const users = await sql`
            SELECT * FROM users
            WHERE num_phone = ${username} OR email = ${username}
                LIMIT 1
        `;

        if (users.length === 0) {
            return res.status(401).json({ message: "Tài khoản không tồn tại!" });
        }

        const user = users[0];

        const isMatch = await bcrypt.compare(password, user.password);

        if (!isMatch) {
            return res.status(401).json({ message: "Mật khẩu không chính xác!" });
        }

        if (user.is_lock) {
            return res.status(403).json({ message: "Tài khoản này đã bị khóa bởi Admin!" });
        }

        const { password: _, ...userData } = user;

        res.status(200).json({
            message: "Đăng nhập thành công!",
            user: userData
        });

    } catch (err) {
        console.error('Lỗi Login:', err.message);
        res.status(500).json({ message: "Lỗi hệ thống!" });
    }
});

// router.post('/register', async (req, res) => {
//     const { fullName, email, numPhone, password} = req.body;
//
//     if (!fullName || !email || !numPhone || !password) {
//         return res.status(400).json({ message: "Vui lòng điền đầy đủ tất cả các trường!" });
//     }
//
//     try {
//         const existingUsers = await sql`
//             SELECT username, email, num_phone FROM users
//             WHERE email = ${email}
//                OR num_phone = ${numPhone}
//                 LIMIT 1
//         `;
//
//         if (existingUsers.length > 0) {
//             const user = existingUsers[0];
//             // if (user.username.toLowerCase() === username.toLowerCase()) return res.status(400).json({ message: "Tên tài khoản đã tồn tại!" });
//             if (user.email.toLowerCase() === email.toLowerCase()) return res.status(400).json({ message: "Email này đã được sử dụng!" });
//             if (user.num_phone === numPhone) return res.status(400).json({ message: "Số điện thoại đã được đăng ký!" });
//         }
//
//         const newUser = await sql`
//             INSERT INTO users (
//                 password,
//                 full_name,
//                 email,
//                 num_phone,
//                 role,
//                 is_lock,
//                 joined_date,
//                 gender,
//                 dob
//             ) VALUES (
//                 ${password},
//                 ${fullName},
//                 ${email},
//                 ${numPhone},
//                 'customer',
//                 FALSE,
//                 CURRENT_DATE,
//                 'none',
//                 NULL
//             )
//             RETURNING id, username, full_name, email, role
//         `;
//
//         res.status(201).json({
//             message: "Đăng ký thành công!",
//             user: newUser[0]
//         });
//
//     } catch (err) {
//         console.error('Lỗi Register:', err.message);
//         res.status(500).json({ message: "Lỗi hệ thống khi tạo tài khoản!" });
//     }
// });

router.post('/register', async (req, res) => {
    const { fullName, email, numPhone, password } = req.body;

    if (!fullName || !email || !numPhone || !password) {
        return res.status(400).json({ message: "Vui lòng điền đầy đủ tất cả các trường!" });
    }

    try {
        const existingUsers = await sql`
            SELECT email, num_phone FROM users
            WHERE email = ${email} OR num_phone = ${numPhone}
            LIMIT 1
        `;

        if (existingUsers.length > 0) {
            const user = existingUsers[0];
            if (user.email.toLowerCase() === email.toLowerCase())
                return res.status(400).json({ message: "Email này đã được sử dụng!" });
            if (user.num_phone === numPhone)
                return res.status(400).json({ message: "Số điện thoại đã được đăng ký!" });
        }

        const salt = await bcrypt.genSalt(10);
        const hashedPassword = await bcrypt.hash(password, salt);

        const newUser = await sql`
            INSERT INTO users (
                password, 
                full_name, 
                email, 
                num_phone, 
                role, 
                is_lock, 
                joined_date, 
                gender
            ) VALUES (
                ${hashedPassword}, -- Dùng mật khẩu đã mã hóa ở đây
                ${fullName}, 
                ${email}, 
                ${numPhone}, 
                'Customer',   
                FALSE,          
                CURRENT_DATE,   
                'none'          
            )
            RETURNING id, full_name, email, role
        `;

        res.status(201).json({
            message: "Đăng ký thành công!",
            user: newUser[0]
        });

    } catch (err) {
        console.error('Lỗi Register:', err.message);
        res.status(500).json({ message: "Lỗi hệ thống khi tạo tài khoản!" });
    }
});

router.get('/information/:id', async (req, res) => {
    const { id } = req.params;

    try {
        const users = await sql`
            SELECT * FROM users WHERE id = ${id} LIMIT 1`
        if (users.length === 0) {
            return res.status(404).json({ message: "Không tìm thấy người dùng!" });
        }
        const user = users[0];
        const { password: _, ...userData } = user;
        res.status(200).json(userData);

    } catch (err) {
        console.error('Lỗi khi lấy thông tin user:', err.message);
        res.status(500).json({ message: "Lỗi máy chủ!" });
    }
});

router.get('/role',async(req,res)=>{
    const roles = await sql`SELECT DISTINCT role FROM users;`;
    const roleNum = await sql`SELECT COUNT(DISTINCT role) FROM users;`;
    if(roles.length === 0||roleNum.length === 0) {
        return res.status(404).json({ message: "Không tìm thấy người dùng!" });
    }
    res.status(200).json({
        "roleNum": roleNum,
        "roles": roles
    })
});

router.put('/information/:id', async (req, res) => {
    const { id } = req.params;
    const { fullName, email, numPhone, dob, gender } = req.body;

    try {
        const existing = await sql`
            SELECT id FROM users 
            WHERE (email = ${email} OR num_phone = ${numPhone}) 
            AND id != ${id}
            LIMIT 1
        `;

        if (existing.length > 0) {
            return res.status(400).json({
                message: "Email hoặc Số điện thoại đã được sử dụng bởi tài khoản khác!"
            });
        }
        const updatedUser = await sql`
            UPDATE users SET
                full_name = ${fullName},
                email = ${email},
                num_phone = ${numPhone},
                dob = ${dob},       -- Giá trị từ input date (YYYY-MM-DD)
                gender = ${gender}  -- Giá trị từ select (Nam/Nữ)
            WHERE id = ${id}
            RETURNING id, full_name, email, num_phone, dob, gender, joined_date
        `;
        //username
        if (updatedUser.length === 0) {
            return res.status(404).json({ message: "Không tìm thấy người dùng!" });
        }
        res.status(200).json({
            message: "Đã lưu thay đổi thành công!",
            user: updatedUser[0]
        });

    } catch (err) {
        console.error('Lỗi Update:', err.message);
        res.status(500).json({ message: "Lỗi hệ thống khi cập nhật!" });
    }
});

router.put('/role/:id', async (req, res) => {
    const { id } = req.params;
    const { role } = req.body;

    if (!role) {
        return res.status(400).json({ message: "Vui lòng cung cấp quyền (role) mới!" });
    }

    try {
        const updatedUser = await sql`
            UPDATE users 
            SET role = ${role}
            WHERE id = ${id}
            RETURNING id, full_name, role
        `;

        if (updatedUser.length === 0) {
            return res.status(404).json({ message: "Không tìm thấy người dùng để cập nhật!" });
        }

        res.status(200).json({
            message: "Cập nhật quyền thành công!",
            user: updatedUser[0]
        });

    } catch (err) {
        console.error('Lỗi Update Role:', err.message);
        res.status(500).json({ message: "Lỗi hệ thống khi cập nhật quyền!" });
    }
});

router.put('/status/:id', async (req, res) => {
    const { id } = req.params;
    const { is_lock } = req.body;

    if (is_lock === undefined) {
        return res.status(400).json({ message: "Vui lòng cung cấp trạng thái khóa (is_lock)!" });
    }

    try {
        const updatedUser = await sql`
            UPDATE users 
            SET is_lock = ${is_lock}
            WHERE id = ${id}
            RETURNING id, full_name, email, is_lock
        `;

        if (updatedUser.length === 0) {
            return res.status(404).json({ message: "Không tìm thấy người dùng để cập nhật trạng thái!" });
        }

        res.status(200).json({
            message: `Tiến hành ${is_lock ? 'khóa' : 'mở khóa'} tài khoản thành công!`,
            user: updatedUser[0]
        });

    } catch (err) {
        console.error('Lỗi Update Status (Lock/Unlock):', err.message);
        res.status(500).json({ message: "Lỗi hệ thống khi cập nhật trạng thái tải khoản!" });
    }
});

const REDIS_URL = process.env.REDIS_URL;

const client = redis.createClient({
    url: REDIS_URL
});

client.on('error', err => console.log('Redis Client Error', err));
client.connect().then(() => console.log("Đã kết nối Redis Cloud thành công ở UserRouter!"));

router.post('/check-password-validity', async (req, res) => {
    const { email, newPassword } = req.body;

    if (!email || !newPassword) {
        return res.status(400).json({ message: "Thiếu Email hoặc Mật khẩu mới!" });
    }

    try {
        // Lấy hash mật khẩu hiện tại từ DB
        const user = await sql`SELECT password FROM users WHERE email = ${email}`;

        if (user.length === 0) {
            return res.status(404).json({ message: "Người dùng không tồn tại!" });
        }

        // So sánh
        const isSame = await bcrypt.compare(newPassword, user[0].password);

        if (isSame) {
            return res.status(400).json({
                isValid: false,
                message: "Mật khẩu mới không được trùng với mật khẩu hiện tại!"
            });
        }

        return res.status(200).json({
            isValid: true,
            message: "Mật khẩu hợp lệ."
        });

    } catch (err) {
        console.error(err);
        res.status(500).json({ message: "Lỗi kiểm tra mật khẩu!" });
    }
});

router.post('/reset-password', async (req, res) => {
    const { email, resetToken, newPassword } = req.body;

    // 1. Kiểm tra đầu vào
    if (!email || !resetToken || !newPassword) {
        return res.status(400).json({
            message: "Vui lòng cung cấp đầy đủ Email, Reset Token và Mật khẩu mới!"
        });
    }

    // Kiểm tra độ dài mật khẩu (Ví dụ tối thiểu 6 ký tự)
    if (newPassword.length < 6) {
        return res.status(400).json({
            message: "Mật khẩu mới phải có ít nhất 6 ký tự!"
        });
    }

    try {
        // 2. Kiểm tra "Vé thông hành" (Reset Token) trong Redis
        const storedToken = await client.get(`resetToken:${email}`);

        // Nếu không có token hoặc token không khớp
        if (!storedToken || storedToken !== resetToken) {
            return res.status(403).json({
                message: "Yêu cầu không hợp lệ hoặc phiên làm việc đã hết hạn. Vui lòng xác thực lại OTP!"
            });
        }
        // ------------------------------------------

        // 3. Mã hóa mật khẩu mới trước khi lưu vào DB
        const salt = await bcrypt.genSalt(10);
        const hashedPassword = await bcrypt.hash(newPassword, salt);

        // 4. Cập nhật mật khẩu vào bảng public.users bằng thư viện sql (Theo mẫu của bạn)
        const updatedUser = await sql`
            UPDATE users 
            SET password = ${hashedPassword}
            WHERE email = ${email}
            RETURNING id, full_name, email, role
        `;

        // Nếu không tìm thấy user (Email sai hoặc đã bị xóa)
        if (updatedUser.length === 0) {
            return res.status(404).json({
                message: "Không tìm thấy người dùng với email này để cập nhật mật khẩu!"
            });
        }

        // 5. Xóa Token trong Redis sau khi đổi mật khẩu thành công để bảo mật tuyệt đối
        await client.del(`resetToken:${email}`);

        // 6. Trả về kết quả thành công
        res.status(200).json({
            success: true,
            message: "Mật khẩu của bạn đã được đặt lại thành công!",
            user: {
                id: updatedUser[0].id,
                full_name: updatedUser[0].full_name,
                email: updatedUser[0].email
            }
        });

    } catch (err) {
        console.error('Lỗi Reset Password:', err.message);
        res.status(500).json({
            message: "Đã xảy ra lỗi hệ thống trong quá trình đặt lại mật khẩu!"
        });
    }
});

router.post('/check-email', async (req, res) => {
    const { email } = req.body;
    console.log(email);
    // 1. Kiểm tra đầu vào
    if (!email) {
        return res.status(400).json({ message: "Vui lòng cung cấp địa chỉ email!" });
    }

    try {
        // 2. Truy vấn kiểm tra email trong bảng users
        const user = await sql`
            SELECT id, email, full_name, is_lock 
            FROM users 
            WHERE email = ${email}
        `;

        // 3. Nếu không tìm thấy người dùng
        if (user.length === 0) {
            return res.status(404).json({
                exists: false,
                message: "Email này chưa được đăng ký trong hệ thống!"
            });
        }

        // 4. Kiểm tra xem tài khoản có đang bị khóa không (Dựa trên cột is_lock của bạn)
        if (user[0].is_lock) {
            return res.status(403).json({
                exists: true,
                is_lock: true,
                message: "Tài khoản này hiện đang bị khóa. Vui lòng liên hệ quản trị viên!"
            });
        }

        // 5. Trả về kết quả thành công để Frontend chuyển sang bước gửi OTP
        res.status(200).json({
            exists: true,
            is_lock: false,
            message: "Email hợp lệ, chuẩn bị gửi mã xác thực!",
            user: {
                full_name: user[0].full_name,
                email: user[0].email
            }
        });

    } catch (err) {
        console.error('Lỗi Check Email:', err.message);
        res.status(500).json({ message: "Lỗi hệ thống khi kiểm tra email!" });
    }
});

router.put('/admin-reset-password/:id', async (req, res) => {
    const { id } = req.params;
    const { password } = req.body;

    if (!password || password.length < 6) {
        return res.status(400).json({ message: "Mật khẩu phải có ít nhất 6 ký tự!" });
    }

    try {
        const salt = await bcrypt.genSalt(10);
        const hashedPassword = await bcrypt.hash(password, salt);

        const updatedUser = await sql`
            UPDATE users 
            SET password = ${hashedPassword}
            WHERE id = ${id}
            RETURNING id, full_name, email
        `;

        if (updatedUser.length === 0) {
            return res.status(404).json({ message: "Không tìm thấy người dùng!" });
        }

        res.status(200).json({
            message: "Đặt lại mật khẩu người dùng thành công!",
            user: updatedUser[0]
        });

    } catch (err) {
        console.error('Lỗi Admin Reset Password:', err.message);
        res.status(500).json({ message: "Lỗi hệ thống khi đặt lại mật khẩu!" });
    }
});

const GoogleAuthRouter = require('./GoogleAuthRouter.jsx')
router.use('/google', GoogleAuthRouter);

module.exports = router;