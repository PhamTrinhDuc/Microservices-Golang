const express = require('express');
const router = express.Router();
const sql = require("../db.jsx");

// 1. Thống kê doanh thu & Tỷ lệ chốt đơn (Dùng cho Chart và KPI)
router.get('/sales-overview', async (req, res) => {
    const { range = 'month' } = req.query;
    try {
        const stats = await sql`
            SELECT 
                DATE_TRUNC(${range}, order_date) as time_period,
                SUM(total_amount) as total_revenue,
                COUNT(order_id) as total_orders,
                COUNT(CASE WHEN status_id = 4 THEN 1 END) as completed_orders,
                COUNT(CASE WHEN status_id = 5 THEN 1 END) as cancelled_orders
            FROM orders
            GROUP BY 1 ORDER BY 1 DESC
        `;
        res.json(stats);
    } catch (err) { res.status(500).json({ message: err.message }); }
});

// 2. Top sản phẩm bán chạy (Dùng cho bảng Top Sellers)
router.get('/top-sellers', async (req, res) => {
    const { limit = 10, period = 'all' } = req.query;
    let timeFilter = sql``;
    if (period === 'week') timeFilter = sql`AND o.order_date >= NOW() - INTERVAL '7 days'`;
    else if (period === 'month') timeFilter = sql`AND o.order_date >= NOW() - INTERVAL '1 month'`;
    else if (period === 'year') timeFilter = sql`AND o.order_date >= NOW() - INTERVAL '1 year'`;

    try {
        const products = await sql`
            SELECT p.product_id, p.name, SUM(od.quantity) as total_sold, SUM(od.quantity * od.unit_price) as total_revenue
            FROM order_details od
                     JOIN product p ON od.product_id = p.product_id
                     JOIN orders o ON od.order_id = o.order_id
            WHERE 1=1 ${timeFilter}
            GROUP BY p.product_id, p.name ORDER BY total_sold DESC LIMIT ${limit}
        `;
        res.json(products);
    } catch (err) { res.status(500).json({ message: err.message }); }
});

// 3. Doanh thu theo phương thức thanh toán
router.get('/payment-revenue', async (req, res) => {
    try {
        const data = await sql`
            SELECT payment_method, COUNT(order_id) as total_orders, SUM(total_amount) as total_revenue
            FROM orders GROUP BY payment_method ORDER BY total_revenue DESC
        `;
        res.json(data);
    } catch (err) { res.status(500).json({ message: err.message }); }
});

// 4. Phân tích chiết khấu & Hỗ trợ (Dùng cho bảng phân tích thanh toán)
router.get('/discount-analysis', async (req, res) => {
    try {
        const data = await sql`
            SELECT
                payment_method,
                0 as total_shipping_support,
                SUM(COALESCE(voucher_discount, 0)) as total_discount_amount,
                ROUND((SUM(COALESCE(voucher_discount, 0)) / NULLIF(SUM(total_amount), 0)) * 100, 2) as discount_rate_percent
            FROM orders GROUP BY payment_method ORDER BY total_discount_amount DESC
        `;
        res.json(data);
    } catch (err) { res.status(500).json({ message: err.message }); }
});

// 5. Báo cáo nhập hàng
router.get('/procurement-report', async (req, res) => {
    try {
        const report = await sql`
            SELECT s.name as supplier_name, COUNT(ii.invoice_id) as total_invoices, SUM(iid.quantity * iid.price_import) as total_spend
            FROM suppliers s
                     JOIN import_invoices ii ON s.supplier_id = ii.supplier_id
                     JOIN import_invoice_details iid ON ii.invoice_id = iid.invoice_id
            GROUP BY s.supplier_id, s.name
        `;
        res.json(report);
    } catch (err) { res.status(500).json({ message: err.message }); }
});

// 6. Khách hàng thân thiết
router.get('/loyal-customers', async (req, res) => {
    try {
        const customers = await sql`
            SELECT u.id, u.full_name, u.email, COUNT(o.order_id) as total_orders, SUM(o.total_amount) as total_spent
            FROM users u JOIN orders o ON u.id = o.user_id
            WHERE o.status_id = 4 GROUP BY u.id, u.full_name, u.email ORDER BY total_spent DESC LIMIT 20
        `;
        res.json(customers);
    } catch (err) { res.status(500).json({ message: err.message }); }
});

// 7. Nhân khẩu học (Dùng cho bản đồ phân bố địa lý)
router.get('/customer-demographics', async (req, res) => {
    try {
        const geoDistribution = await sql`
            SELECT 
                a.province,
                COUNT(DISTINCT a.id_user) as count,
                COUNT(DISTINCT o.order_id) as total_orders,
                COUNT(DISTINCT CASE WHEN o.status_id = 4 THEN o.order_id END) as completed_orders
            FROM address a
            LEFT JOIN orders o ON o.user_id = a.id_user AND o.reciever_address ILIKE '%' || a.province || '%'
            WHERE a.province IS NOT NULL AND a.province != '' AND a.is_delete IS NOT TRUE
            GROUP BY a.province
            ORDER BY count DESC
        `;
        const genderDist = await sql`SELECT gender, COUNT(*) as count FROM users GROUP BY gender`;
        res.json({ geoDistribution, genderDist });
    } catch (err) { res.status(500).json({ message: err.message }); }
});

// 7.2 Lấy danh sách khách hàng theo tỉnh thành
router.get('/customers-by-province', async (req, res) => {
    const { province } = req.query;
    if (!province) {
        return res.status(400).json({ message: "Vui lòng cung cấp tỉnh thành!" });
    }
    try {
        const customers = await sql`
            SELECT DISTINCT u.id, u.full_name, u.email, u.num_phone, u.joined_date,
                (SELECT COUNT(*) FROM orders o WHERE o.user_id = u.id) as total_orders,
                (SELECT COUNT(*) FROM orders o WHERE o.user_id = u.id AND o.status_id = 4) as completed_orders
            FROM users u
            JOIN address a ON u.id = a.id_user
            WHERE a.province = ${province} AND a.is_delete IS NOT TRUE
            ORDER BY u.joined_date DESC
        `;
        res.json(customers);
    } catch (err) { 
        console.error("Lỗi lấy khách hàng theo tỉnh thành:", err);
        res.status(500).json({ message: err.message }); 
    }
});

// 8. KPI Dashboard (Dùng cho 3 thẻ to ở trên cùng)
router.get('/kpi-dashboard', async (req, res) => {
    try {
        const kpi = await sql`
            SELECT 
                SUM(total_amount) as total_revenue,
                SUM(total_amount) / NULLIF(COUNT(order_id), 0) as average_order_value,
                (SELECT COUNT(*) FROM orders) as total_orders,
                (SELECT COUNT(*) FROM orders WHERE status_id = 4) as completed_orders
            FROM orders
        `;
        res.json(kpi.length > 0 ? kpi[0] : { total_revenue: 0, average_order_value: 0 });
    } catch (err) { res.status(500).json({ message: err.message }); }
});

module.exports = router;
