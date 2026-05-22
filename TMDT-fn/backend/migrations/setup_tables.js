require('dotenv').config();
const sql = require('../db.jsx');

async function migrate() {
    try {
        console.log("Starting database migrations...");

        // 1. Alter product table
        console.log("Checking and altering 'product' table...");
        await sql`
            ALTER TABLE product 
            ADD COLUMN IF NOT EXISTS low_stock_threshold INT DEFAULT 5
        `;
        await sql`
            ALTER TABLE product 
            ADD COLUMN IF NOT EXISTS base_price_numeric NUMERIC
        `;

        // 2. Alter orders table
        console.log("Checking and altering 'orders' table...");
        await sql`
            ALTER TABLE orders 
            ADD COLUMN IF NOT EXISTS voucher_code VARCHAR(100)
        `;
        await sql`
            ALTER TABLE orders 
            ADD COLUMN IF NOT EXISTS voucher_discount NUMERIC DEFAULT 0
        `;

        // 3. Create vouchers table
        console.log("Checking and creating 'vouchers' table...");
        await sql`
            CREATE TABLE IF NOT EXISTS vouchers (
                voucher_id SERIAL PRIMARY KEY,
                code VARCHAR(100) UNIQUE NOT NULL,
                name VARCHAR(255) NOT NULL,
                start_date TIMESTAMP NOT NULL,
                end_date TIMESTAMP NOT NULL,
                discount_type VARCHAR(50) NOT NULL,
                discount_value NUMERIC NOT NULL,
                discount_target VARCHAR(50) NOT NULL,
                min_order_value NUMERIC NOT NULL DEFAULT 0,
                max_discount_amount NUMERIC,
                is_deleted BOOLEAN DEFAULT FALSE
            )
        `;

        // 4. Create promotions table
        console.log("Checking and creating 'promotions' table...");
        await sql`
            CREATE TABLE IF NOT EXISTS promotions (
                promotion_id SERIAL PRIMARY KEY,
                product_id VARCHAR(100) NOT NULL,
                discount_percent INT NOT NULL,
                start_date TIMESTAMP NOT NULL,
                end_date TIMESTAMP NOT NULL,
                is_deleted BOOLEAN DEFAULT FALSE
            )
        `;

        console.log("Migrations applied successfully!");
    } catch (err) {
        console.error("Migration error:", err);
    } finally {
        await sql.end();
    }
}

migrate();
