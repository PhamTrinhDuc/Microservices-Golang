require('dotenv').config();
const postgres = require('postgres');
const sql = postgres(process.env.DATABASE_URL);

async function run() {
    try {
        console.log("Listing tables:");
        const tables = await sql`
            SELECT table_name 
            FROM information_schema.tables 
            WHERE table_schema = 'public'
        `;
        console.log("Tables:", tables.map(t => t.table_name));

        const targetTables = ['pagenigation', 'product', 'users', 'orders', 'address', 'vouchers', 'promotions'];
        for (const tbl of targetTables) {
            const cols = await sql`
                SELECT column_name, data_type 
                FROM information_schema.columns 
                WHERE table_name = ${tbl}
            `;
            if (cols.length > 0) {
                console.log(`Schema for table ${tbl}:`, cols.map(c => `${c.column_name} (${c.data_type})`));
            } else {
                console.log(`Table ${tbl} does not exist.`);
            }
        }
    } catch (err) {
        console.error("Error executing query:", err);
    } finally {
        await sql.end();
    }
}

run();
