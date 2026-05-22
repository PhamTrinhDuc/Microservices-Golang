require('dotenv').config();
const sql = require('../db.jsx');

async function migrate() {
    try {
        console.log("Starting chat database migrations...");

        // Create chat_rooms
        console.log("Creating 'chat_rooms' table...");
        await sql`
            CREATE TABLE IF NOT EXISTS chat_rooms (
                room_id SERIAL PRIMARY KEY,
                customer_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
                created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
            )
        `;

        // Create chat_messages
        console.log("Creating 'chat_messages' table...");
        await sql`
            CREATE TABLE IF NOT EXISTS chat_messages (
                message_id SERIAL PRIMARY KEY,
                room_id INT NOT NULL REFERENCES chat_rooms(room_id) ON DELETE CASCADE,
                sender_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
                sender_role VARCHAR(20) NOT NULL, -- 'customer' or 'admin'
                message_text TEXT NOT NULL,
                is_read BOOLEAN DEFAULT FALSE,
                created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
            )
        `;

        // Check if indexes are needed
        console.log("Creating indexes...");
        await sql`
            CREATE INDEX IF NOT EXISTS idx_chat_rooms_customer_id ON chat_rooms(customer_id)
        `;
        await sql`
            CREATE INDEX IF NOT EXISTS idx_chat_messages_room_id ON chat_messages(room_id)
        `;

        console.log("Chat migrations applied successfully!");
    } catch (err) {
        console.error("Chat migration error:", err);
    } finally {
        await sql.end();
    }
}

migrate();
