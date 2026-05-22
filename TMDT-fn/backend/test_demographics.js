require('dotenv').config();
const postgres = require('postgres');
const sql = postgres(process.env.DATABASE_URL);

async function test() {
    try {
        const query = await sql`
            SELECT 
                a.province,
                COUNT(DISTINCT a.id_user) as customer_count,
                COUNT(DISTINCT o.order_id) as total_orders,
                COUNT(DISTINCT CASE WHEN o.status_id = 4 THEN o.order_id END) as completed_orders
            FROM address a
            LEFT JOIN orders o ON o.reciever_address ILIKE '%' || a.province || '%'
            GROUP BY a.province
            ORDER BY customer_count DESC
        `;
        console.log("Demographics with success rate query results:", query);
    } catch (err) {
        console.error("Error executing test query:", err);
    }
    process.exit();
}
test();
