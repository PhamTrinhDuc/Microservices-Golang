const postgres = require('postgres');
require('dotenv').config();

const connectionString = process.env.DATABASE_URL;
const sql = postgres(connectionString, {
    onnotice: (notice) => console.log('Postgres Notice:', notice),
    prepare: false,
})

module.exports = sql;