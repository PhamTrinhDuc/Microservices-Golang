/**
 * Dùng node-cron để chạy cron job mỗi phút, quét và rollback
 * tất cả reservation đã hết hạn trong PostgreSQL.
 */

const cron = require('node-cron');
const { rollbackExpiredReservations } = require('./inventoryService.js');

function startReservationCleanupJob() {
    const EVERY_MINUTE = '0 * * * * *';

    cron.schedule(EVERY_MINUTE, async () => {
        try {
            const count = await rollbackExpiredReservations();
            if (count > 0) {
                console.log(`[Cleanup] Rolled back ${count} expired reservation(s)`);
            }
        } catch (err) {
            console.error('[Cleanup] Error during reservation cleanup:', err.message);
        }
    });

    console.log('[Cleanup] Reservation cleanup cron job started (runs every minute)');
}

module.exports = { startReservationCleanupJob };
