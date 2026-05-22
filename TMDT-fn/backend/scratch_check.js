require('dotenv').config();
const postgres = require('postgres');
const sql = postgres(process.env.DATABASE_URL);

async function run() {
    try {
        console.log("Adding price_base column to productvariant table...");
        await sql`ALTER TABLE productvariant ADD COLUMN IF NOT EXISTS price_base BIGINT DEFAULT 0;`;
        console.log("Successfully added price_base column!");

        console.log("Migrating existing base_price_numeric data to price_base in productvariant...");
        await sql`
            UPDATE productvariant pv 
            SET price_base = COALESCE(
                (SELECT p.base_price_numeric FROM product p WHERE p.product_id = pv.product_id), 
                pv.price,
                0
            )
            WHERE pv.price_base IS NULL OR pv.price_base = 0;
        `;
        console.log("Successfully migrated existing cost price data!");

        // Let's verify the columns again
        const cols = await sql`
            SELECT column_name, data_type 
            FROM information_schema.columns 
            WHERE table_name = 'productvariant'
        `;
        console.log("Updated productvariant Schema:", cols.map(c => `${c.column_name} (${c.data_type})`));
    } catch (err) {
        console.error("Migration failed:", err);
    }
    process.exit();
}
run();
