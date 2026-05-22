require('dotenv').config();
const postgres = require('postgres');
const sql = postgres(process.env.DATABASE_URL);

async function run() {
    try {
        for (const tbl of ['productinventory', 'productvariant']) {
            const cols = await sql`
                SELECT column_name, data_type 
                FROM information_schema.columns 
                WHERE table_name = ${tbl}
            `;
            console.log(`Schema for table ${tbl}:`, cols.map(c => `${c.column_name} (${c.data_type})`));
        }
    } catch (err) {
        console.error("Error executing query:", err);
    } finally {
        await sql.end();
    }
}

run();
