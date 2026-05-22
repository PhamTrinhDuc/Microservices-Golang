const express = require('express');
const cors = require('cors');
const dbConnect = require('./DBconnect.jsx');
const swaggerUi = require('swagger-ui-express');
const swaggerDocument = require('./swagger.json');
const productsRouter = require("./products/ProductsRouter.jsx")
const usersRouter = require("./users/UsersRouter.jsx")
const orderRouter = require("./products/OrderRouter.jsx");
const addressUserRouter = require("./users/AddressUserRouter.jsx");
const addressRouter = require("./address/AddressRouter.jsx");
const checkingRouter = require("./datatest/checking.jsx");
const storeAddressRouter = require("./address/StoreAddressRouter.jsx");
const shippingServiceRouter = require("./shipping/ShippingServiceRouter.jsx");
const OTPRouter = require("./otp/OTP.jsx");
const sql = require("./db.jsx");
const { startReservationCleanupJob } = require('./services/reservationCleanup.js');
const app = express();
app.use(express.json());
// dbConnect();
//cloudflared-windows-amd64.exe tunnel --url http://localhost:8080

console.log('Đang kết nối tới PostgreSQL (Supabase)...');
sql`SELECT 1`.then(() => {
    console.log('Kết nối PostgreSQL thành công!');
}).catch((err) => {
    console.error('Kết nối PostgreSQL thất bại:', err.message);
});

app.use(cors());

app.use('/api-docs', swaggerUi.serve, swaggerUi.setup(swaggerDocument));

app.get('/', (req, res) => {
    res.json("Server is running in port 8080");
})
app.use((req, res, next) => {
    console.log(`🔍 Request đi qua: ${req.method} ${req.url}`);
    next();
});

app.use('/api/product', productsRouter);

app.use('/api/user', usersRouter);

app.use('/api/order', orderRouter);

app.use('/api/user/address', addressUserRouter);

app.use('/api/address',addressRouter);

app.use('/api/address/store',storeAddressRouter);

app.use('/api/otp/',OTPRouter);

app.use('/api/service/shipping', shippingServiceRouter);

// app.use('/api/shippingfee');

app.use('/api/checking',checkingRouter);

app.listen(8080, () => {
    console.log("Listen on 8080");
    startReservationCleanupJob();
})