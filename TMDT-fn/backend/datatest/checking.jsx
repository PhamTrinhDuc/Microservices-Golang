const express = require('express');
const router = express.Router();
const sql = require('../db.jsx');

router.get('/products/missing-weight', async (req, res) => {
    try {
        const data = await sql.unsafe(`
            SELECT
                p.product_id,
                p.name,
                p.category_id,
                p.price,
                p.specs
            FROM product p
            WHERE
                p.specs IS NOT NULL
              AND (
                -- ĐIỀU KIỆN 1: Không có trường khối lượng VÀ không có trường kích thước
                (
                    NOT EXISTS (
                        SELECT 1
                        FROM jsonb_each(p.specs) section
                        JOIN jsonb_each(section.value) kv ON true
                        WHERE kv.key ILIKE '%khối lượng%'
                        OR kv.key ILIKE '%khoi luong%'
                        )
                        AND NOT EXISTS (
                        SELECT 1
                        FROM jsonb_each(p.specs) section
                        JOIN jsonb_each(section.value) kv ON true
                        WHERE kv.key ILIKE '%kích thước%'
                            OR kv.key ILIKE '%kich thuoc%'
                        )
                    )

                    OR

                    -- ĐIỀU KIỆN 2: Có trường khối lượng nhưng giá trị không chứa g/gr/kg
                EXISTS (
                    SELECT 1
                    FROM jsonb_each(p.specs) section
                    JOIN jsonb_each(section.value) kv ON true
                    WHERE (
                              kv.key ILIKE '%khối lượng%'
                            OR kv.key ILIKE '%khoi luong%'
                              )
                    AND kv.value::text !~* '[0-9]\\s*(kg|gr|g)\\y'
                    )

                    OR

                    -- ĐIỀU KIỆN 3: Có trường kích thước nhưng giá trị không chứa g/gr/kg
                EXISTS (
                    SELECT 1
                    FROM jsonb_each(p.specs) section
                    JOIN jsonb_each(section.value) kv ON true
                    WHERE (
                              kv.key ILIKE '%kích thước%'
                            OR kv.key ILIKE '%kich thuoc%'
                              )
                    AND kv.value::text !~* '[0-9]\\s*(kg|gr|g)\\y'
                    )
                )
        `);

        res.json({
            success: true,
            count: data.length,
            message: "Danh sách sản phẩm chưa có thông tin cân nặng (gram/kg)",
            data
        });
    } catch (error) {
        console.error("Lỗi truy vấn:", error);
        res.status(500).json({ success: false, message: error.message });
    }
});

module.exports = router;