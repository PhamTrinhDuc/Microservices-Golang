require('dotenv').config();
const postgres = require('postgres');
const sql = postgres(process.env.DATABASE_URL);

async function check() {
    try {
        for (const tbl of ['orders', 'address']) {
            const cols = await sql`
                SELECT column_name, data_type 
                FROM information_schema.columns 
                WHERE table_name = ${tbl}
            `;
            console.log(`Schema for table ${tbl}:`, cols.map(c => `${c.column_name} (${c.data_type})`));
        }

        // Print sample from orders
        const sampleOrder = await sql`SELECT * FROM orders LIMIT 1`;
        console.log("Sample order:", sampleOrder[0]);

        // Print sample from address
        const sampleAddress = await sql`SELECT * FROM address LIMIT 1`;
        console.log("Sample address:", sampleAddress[0]);

    } catch (err) {
        console.error("Error fetching schema:", err);
    }
    process.exit();
}
check();
