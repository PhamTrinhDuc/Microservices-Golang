const express = require('express');
const cors = require('cors');
const { createServer } = require('http'); // Thêm HTTP Server
const { Server } = require('socket.io');   // Thêm Socket.io
const swaggerUi = require('swagger-ui-express');
const swaggerDocument = require('./swagger.json');
require('dotenv').config();

// --- Import Routers ---
const productsRouter = require("./products/ProductsRouter.jsx");
const usersRouter = require("./users/UsersRouter.jsx");
const orderRouter = require("./products/OrderRouter.jsx");
const addressUserRouter = require("./users/AddressUserRouter.jsx");
const addressRouter = require("./address/AddressRouter.jsx");
const checkingRouter = require("./datatest/checking.jsx");
const storeAddressRouter = require("./address/StoreAddressRouter.jsx");
const shippingServiceRouter = require("./shipping/ShippingServiceRouter.jsx");
const StatisticRouter = require("./statistics/StatisticsRouter.jsx");
const OTPRouter = require("./otp/OTP.jsx");
const paginationRouter = require("./products/PaginationRouter.js");
const vouchersRouter = require("./products/VouchersRouter.jsx");
const promotionsRouter = require("./products/PromotionsRouter.jsx");
const reviewsRouter = require("./products/ReviewsRouter.jsx");
const clientURL = process.env.CLIENT_URL;

//tunel.exe tunnel --config config.yml run

// --- DB & Services ---
const sql = require("./db.jsx");
const { startReservationCleanupJob } = require('./services/reservationCleanup.js');

const app = express();
const PORT = 8080;

// 1. Khởi tạo HTTP Server và Socket.io
const httpServer = createServer(app);
const io = new Server(httpServer, {
    cors: {
        origin: clientURL,
        methods: ["GET", "POST"]
    }
});

// 2. Chia sẻ đối tượng 'io' để sử dụng trong các file Router khác thông qua
app.set('socketio', io);

// 3. Middlewares
app.use(express.json());
app.use(cors({
    origin: clientURL,
    methods: ['GET', 'POST', 'PUT', 'PATCH', 'DELETE'],
    credentials: true
}));

// Middleware Logging
app.use((req, res, next) => {
    console.log(`[${new Date().toLocaleString()}] Request: ${req.method} ${req.url}`);
    next();
});

// 4. Kiểm tra kết nối Database
console.log('Đang kết nối tới PostgreSQL (Supabase)...');
sql`SELECT 1`.then(() => {
    console.log('Kết nối PostgreSQL thành công!');
}).catch((err) => {
    console.error('Kết nối PostgreSQL thất bại:', err.message);
});

// 5. Cấu hình Socket.io Connection (Lắng nghe sự kiện toàn cục)
io.on('connection', (socket) => {
    console.log(`Client mới kết nối: ${socket.id}`);

    // Ví dụ: Lắng nghe sự kiện từ client
    socket.on('join_order_room', (orderId) => {
        socket.join(`order_${orderId}`);
        console.log(`Client tham gia vào room Order: ${orderId}`);
    });

    // Chat Sockets
    socket.on('join_chat_room', (roomId) => {
        socket.join(`chat_${roomId}`);
        console.log(`Client ${socket.id} tham gia vào room Chat: ${roomId}`);
    });

    socket.on('send_chat_message', async (data) => {
        const { room_id, sender_id, sender_role, message_text } = data;
        try {
            const [newMsg] = await sql`
                INSERT INTO chat_messages (room_id, sender_id, sender_role, message_text)
                VALUES (${room_id}, ${sender_id}, ${sender_role}, ${message_text})
                RETURNING 
                    message_id, 
                    room_id, 
                    sender_id, 
                    sender_role, 
                    message_text, 
                    is_read, 
                    created_at AT TIME ZONE 'UTC' AT TIME ZONE 'Asia/Ho_Chi_Minh' AS created_at
            `;

            await sql`
                UPDATE chat_rooms 
                SET updated_at = NOW() 
                WHERE room_id = ${room_id}
            `;

            console.log(newMsg);
            io.to(`chat_${room_id}`).emit('new_chat_message', newMsg);
            io.emit('chat_rooms_updated');
        } catch (error) {
            console.error('Lỗi truyền tin nhắn socket:', error);
        }
    });

    socket.on('mark_messages_read', async ({ room_id, role }) => {
        try {
            const targetRole = role === 'admin' ? 'customer' : 'admin';
            await sql`
                UPDATE chat_messages 
                SET is_read = true 
                WHERE room_id = ${room_id} AND sender_role = ${targetRole}
            `;
            io.emit('chat_rooms_updated');
        } catch (error) {
            console.error('Lỗi đọc tin nhắn socket:', error);
        }
    });

    socket.on('disconnect', () => {
        console.log(`Client ngắt kết nối: ${socket.id}`);
    });
});

// 6. Routes
app.use('/api-docs', swaggerUi.serve, swaggerUi.setup(swaggerDocument));

app.get('/', (req, res) => {
    res.json({ message: "Server and Socket are running on port 8080" });
});

// Gắn Routers
app.use('/api/product', productsRouter);
app.use('/api/product/reviews', reviewsRouter);
app.use('/api/chat', require('./products/ChatRouter.jsx'));
app.use('/api/cart', require('./products/CartRouter.jsx'));

app.use('/api/user', usersRouter);

app.use('/api/order', orderRouter);

app.use('/api/user/address', addressUserRouter);

app.use('/api/address',addressRouter);

app.use('/api/address/store',storeAddressRouter);

app.use('/api/otp/',OTPRouter);

app.use('/api/service/shipping', shippingServiceRouter);

app.use('/api/statistic',StatisticRouter);

// app.use('/api/shippingfee');

app.use('/api/checking',checkingRouter);

app.use('/api/pagination', paginationRouter);

app.use('/api/voucher', vouchersRouter);

app.use('/api/promotion', promotionsRouter);

httpServer.listen(PORT, () => {
    console.log(`Server is running on http://localhost:${PORT}`);
    console.log(`Swagger docs: http://localhost:${PORT}/api-docs`);

    // Kiểm tra DB sau khi server chạy
    sql`SELECT 1`.then(() => {
        console.log('PostgreSQL Connected!');
    }).catch((err) => {
        console.error('DB Connection Error:', err.message);
    });

    startReservationCleanupJob();
});