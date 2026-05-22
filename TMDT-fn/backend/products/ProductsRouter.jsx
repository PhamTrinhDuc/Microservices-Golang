const express = require('express');
const router = express.Router();
const sql = require("../db.jsx");
const axios = require("axios");
const SERVER_URL = process.env.SERVER_URL;


const formatProductPaths = (product) => {
    if (!product) return null;
    const getFullUrl = (img) => `/${product.category_id}/${product.product_id}/${img}`;

    return {
        ...product,
        img_thumb: product.img_thumb ? getFullUrl(product.img_thumb) : null,
        local_desc_images: (product.local_desc_images || []).map(getFullUrl),
        variants: (product.variants || []).map(variant => ({
            ...variant,
            local_gallery: (variant.local_gallery || []).map(getFullUrl),
            amount: variant.quantity || 0
        }))
    };
};

// ════════════════════════════════════════════════════════════════════════════════
// KIỂM TRA TỒN KHO CỦA MỘT DANH SÁCH SẢN PHẨM
// Body: { items: [ { variant_id: 1, quantity: 2 }, { variant_id: 2, quantity: 1 } ] }
// ════════════════════════════════════════════════════════════════════════════════
router.post('/check-inventory', async (req, res) => {
    const { items } = req.body;

    if (!items || !Array.isArray(items) || items.length === 0) {
        return res.status(400).json({
            success: false,
            message: "Danh sách sản phẩm kiểm tra không hợp lệ."
        });
    }

    try {
        const results = [];
        let allInStock = true;

        for (const item of items) {
            const { variant_id, quantity: requestedQty } = item;

            if (!variant_id || requestedQty <= 0) continue;

            const [inv] = await sql`
                SELECT 
                    pi.quantity, 
                    pi.reserved,
                    p.name as product_name,
                    pv.color_name,
                    COALESCE(
                        pv.price - (pv.price * pr.discount_percent / 100),
                        pv.price
                    ) AS current_price
                FROM productinventory pi
                JOIN productvariant pv ON pi.variant_id = pv.variant_id
                JOIN product p ON pv.product_id = p.product_id
                LEFT JOIN promotions pr ON p.product_id = pr.product_id 
                    AND pr.is_deleted = false 
                    AND CURRENT_TIMESTAMP >= pr.start_date 
                    AND CURRENT_TIMESTAMP <= pr.end_date
                WHERE pi.variant_id = ${variant_id}
            `;

            if (!inv) {
                results.push({
                    variant_id,
                    product_name: "Sản phẩm không tồn tại",
                    requested: requestedQty,
                    available: 0,
                    is_enough: false
                });
                allInStock = false;
                continue;
            }
            const availableQty = inv.quantity - inv.reserved;
            const isEnough = availableQty >= requestedQty;

            if (!isEnough) allInStock = false;

            results.push({
                variant_id,
                product_name: inv.product_name,
                color_name: inv.color_name,
                requested: requestedQty,
                available: Math.max(0, availableQty),
                is_enough: isEnough,
                current_price: Number(inv.current_price)
            });
        }

        res.status(200).json({
            success: true,
            all_in_stock: allInStock,
            data: results
        });

    } catch (error) {
        console.error('Lỗi API Check Inventory:', error.message);
        res.status(500).json({
            success: false,
            message: "Lỗi hệ thống khi kiểm tra tồn kho."
        });
    }
});

router.get('/category', async (req, res) => {
    try {
        const allCategories = await sql`SELECT * FROM category ORDER BY menu_id ASC`;
        const buildMenu = (categories) => {
            const rootCategories = categories.filter(c => c.parent_id === null);
            return rootCategories.map(root => {
                const subGroups = categories.filter(c => c.parent_id === root.category_id);

                const result = {
                    id: root.menu_id,
                    label: root.category_name,
                    icon: root.icon,
                    slug: root.slug
                };
                if (subGroups.length > 0) {
                    result.hasSubmenu = true;
                    result.submenu = subGroups.map(group => {
                        const items = categories.filter(c => c.parent_id === group.category_id);

                        return {
                            title: group.category_name,
                            items: items.map(item => ({
                                label: item.category_name,
                                slug: item.slug
                            }))
                        };
                    });
                }

                return result;
            });
        };

        const formattedMenu = buildMenu(allCategories);

        res.status(200).json({
            success: true,
            data: formattedMenu
        });

    } catch (error) {
        console.error("Lỗi format menu:", error.message);
        res.status(500).json({ success: false, message: error.message });
    }
});

router.get('/form-templates', async (req, res) => {
    try {
        const configs = await sql`SELECT * FROM categoryformconfig`;
        const form_category = configs.reduce((acc, curr) => {
            acc[curr.category_id] = { fields: curr.fields };
            return acc;
        }, {});
        res.json(form_category);
    } catch (error) {
        res.status(500).json({ message: "Lỗi DB" });
    }
});

router.get('/all-products', async (req, res) => {
    try {
        const pageRes = await axios.get(`${SERVER_URL}/api/pagination`);
        const pageData = pageRes.data.data[0];

        const item_in_home = pageData.item_in_home;
        const column_in_home = pageData.column_in_home;

        const limitPerCategory = item_in_home;

        const products = await sql`
            WITH RECURSIVE CategoryTree AS (
                SELECT category_id, category_name, category_id as root_id, category_name as root_name, menu_id
                FROM category
                WHERE parent_id IS NULL
                UNION ALL
                SELECT c.category_id, c.category_name, ct.root_id, ct.root_name, ct.menu_id
                FROM category c
                         INNER JOIN CategoryTree ct ON c.parent_id = ct.category_id
            ),
                           BaseProducts AS (
                               SELECT
                                   COALESCE(p.base_id, p.product_id) as group_id,
                                   MAX(p.product_id) as max_product_id,
                                   ct.root_id,
                                   ct.root_name,
                                   ct.menu_id
                               FROM product p
                                        JOIN CategoryTree ct ON p.category_id = ct.category_id
                               GROUP BY COALESCE(p.base_id, p.product_id), ct.root_id, ct.root_name, ct.menu_id
                           ),
                           RankedBases AS (
                               SELECT
                                   group_id, root_id, root_name, menu_id,
                                   ROW_NUMBER() OVER(PARTITION BY root_id ORDER BY max_product_id DESC) as rn
                               FROM BaseProducts
                           ),
                           TopBases AS (
                               SELECT group_id, root_id, root_name, menu_id, rn
                               FROM RankedBases
                               WHERE rn <= ${limitPerCategory}
                           )
            SELECT
                p.*,
                (
                    SELECT ROUND(COALESCE(AVG(r.rating), 4.9)::numeric, 1)
                    FROM reviews r
                    WHERE r.product_id = p.product_id AND r.status = 'approved'
                ) as rating,
                tb.root_id,
                tb.root_name,
                tb.menu_id,
                tb.group_id,
                tb.rn, 
                MAX(prom.discount_percent) as discount_percent,
                COALESCE(
                        json_agg(
                                json_build_object(
                                        'variant_id', v.variant_id,
                                        'color_name', v.color_name,
                                        'color_code', v.color_code,
                                        'price', v.price,
                                        'price_str', v.price_str,
                                        'local_gallery', v.local_gallery,
                                        'quantity', i.quantity,
                                        'reserved', i.reserved,
                                        'price_base', v.price_base
                                )
                        ) FILTER (WHERE v.variant_id IS NOT NULL),
                        '[]'
                ) as variants
            FROM product p
                     JOIN TopBases tb ON COALESCE(p.base_id, p.product_id) = tb.group_id
                     LEFT JOIN productvariant v ON p.product_id = v.product_id
                     LEFT JOIN productinventory i ON v.variant_id = i.variant_id
                     LEFT JOIN promotions prom ON p.product_id = prom.product_id AND NOW() BETWEEN prom.start_date AND prom.end_date AND prom.is_deleted IS NOT TRUE
            GROUP BY p.product_id, tb.root_id, tb.root_name, tb.menu_id, tb.group_id, tb.rn
            ORDER BY tb.menu_id ASC, tb.rn ASC, p.price ASC
        `;

        const formattedProducts = products.map(formatProductPaths);
        const categoryMap = new Map();

        formattedProducts.forEach(prod => {
            const { root_id, root_name, menu_id, group_id, rn, ...productData } = prod;

            if (!categoryMap.has(root_id)) {
                categoryMap.set(root_id, {
                    category_id: root_id,
                    category_name: root_name,
                    menu_id: menu_id,
                    products: []
                });
            }

            const categoryObj = categoryMap.get(root_id);
            let card = categoryObj.products.find(p => p.group_id === group_id);

            if (!card) {
                card = {
                    ...productData,
                    group_id: group_id,
                    versions: [productData]
                };
                categoryObj.products.push(card);
            } else {
                card.versions.push(productData);
            }
        });

        const finalData = Array.from(categoryMap.values()).sort((a, b) => {
            return (a.menu_id || 99) - (b.menu_id || 99);
        });

        res.status(200).json({
            success: true,
            data: finalData,
            column: column_in_home
        });

    } catch (error) {
        console.error("Lỗi format all-products:", error.message);
        res.status(500).json({ success: false, message: error.message });
    }
});

// ════════════════════════════════════════════════════════════════════════════════
// XEM CHI TIẾT 1 SẢN PHẨM KÈM CÁC PHIÊN BẢN CẤU HÌNH (CÙNG BASE_ID) VÀ BIẾN THỂ MÀU SẮC
// ════════════════════════════════════════════════════════════════════════════════
router.get('/product/:id', async (req, res) => {
    try {
        const identifier = req.params.id;

        const [mainProduct] = await sql`
            SELECT p.*, c.category_name, c.slug as category_slug,
                   (
                       SELECT ROUND(COALESCE(AVG(r.rating), 4.9)::numeric, 1)
                       FROM reviews r
                       WHERE r.product_id = p.product_id AND r.status = 'approved'
                   ) as rating
            FROM product p
            LEFT JOIN category c ON p.category_id = c.category_id
            WHERE p.product_id = ${identifier} 
               OR p.name_id = ${identifier} 
            LIMIT 1
        `;

        if (!mainProduct) {
            return res.status(404).json({
                success: false,
                message: "Không tìm thấy sản phẩm."
            });
        }

        const groupId = mainProduct.base_id || mainProduct.product_id;

        const productsInGroup = await sql`
            SELECT 
                p.*,
                (
                    SELECT ROUND(COALESCE(AVG(r.rating), 4.9)::numeric, 1)
                    FROM reviews r
                    WHERE r.product_id = p.product_id AND r.status = 'approved'
                ) as rating,
                MAX(prom.discount_percent) as discount_percent,
                COALESCE(
                    json_agg(
                        json_build_object(
                            'variant_id', v.variant_id,
                            'color_name', v.color_name,
                            'color_code', v.color_code,
                            'price', v.price,
                            'price_str', v.price_str,
                            'local_gallery', v.local_gallery,
                            'quantity', i.quantity,
                            'reserved', i.reserved,
                            'price_base', v.price_base
                        )
                    ) FILTER (WHERE v.variant_id IS NOT NULL), 
                    '[]'
                ) as variants
            FROM product p
            LEFT JOIN productvariant v ON p.product_id = v.product_id
            LEFT JOIN productinventory i ON v.variant_id = i.variant_id
            LEFT JOIN promotions prom ON p.product_id = prom.product_id AND NOW() BETWEEN prom.start_date AND prom.end_date AND prom.is_deleted IS NOT TRUE
            WHERE p.base_id = ${groupId} OR p.product_id = ${groupId}
            GROUP BY p.product_id
            ORDER BY p.price ASC
        `;

        const formattedGroup = productsInGroup.map(formatProductPaths);
        let targetProduct = formattedGroup.find(p => p.product_id === mainProduct.product_id);
        if (!targetProduct) targetProduct = formattedGroup[0];
        targetProduct.category_name = mainProduct.category_name;
        targetProduct.category_slug = mainProduct.category_slug;

        targetProduct.versions = formattedGroup.map(prod => {
            const { variants, local_desc_images, description, specs, ...basicInfo } = prod;
            return basicInfo;
        });

        res.status(200).json({
            success: true,
            data: targetProduct
        });

    } catch (error) {
        console.error("Lỗi API Product Detail:", error.message);
        res.status(500).json({ success: false, message: error.message });
    }
});

const findValueInSpecs = (specs, targetField) => {
    if (!specs) return null;
    for (const group in specs) {
        if (specs[group] && specs[group][targetField]) {
            return specs[group][targetField];
        }
    }
    return null;
};

const getAllAvailableSpecFields = (products) => {
    const allFields = new Set();
    const forbidden = ['hang', 'brand', 'thuong hieu', 'nha san xuat'];
    products.forEach(product => {
        if (product.specs) {
            Object.values(product.specs).forEach(group => {
                Object.keys(group).forEach(key => {
                    const cleanKey = key.toLowerCase().normalize("NFD").replace(/[\u0300-\u036f]/g, "").trim();
                    if (!forbidden.some(word => cleanKey.includes(word))) {
                        allFields.add(key);
                    }
                });
            });
        }
    });
    return Array.from(allFields);
};

const isEligibleFilter = (products, targetField) => {
    let hasValue = false;
    for (const p of products) {
        const val = findValueInSpecs(p.specs, targetField);
        if (val) {
            hasValue = true;
            const checkVal = Array.isArray(val) ? String(val[0]) : String(val);
            if (checkVal.length > 50) return false;
        }
    }
    return hasValue;
};

const getDynamicUniqueValues = (products, targetField) => {
    const values = new Set();
    products.forEach(p => {
        let val = findValueInSpecs(p.specs, targetField);
        if (val) {
            if (Array.isArray(val)) val = val[0];
            let cleanVal = String(val).trim();
            if (cleanVal && cleanVal !== "Hãng không công bố" && cleanVal !== "Đang cập nhật") {
                values.add(cleanVal);
            }
        }
    });
    return Array.from(values).sort();
};

router.get('/filters/:category', async (req, res) => {
    try {
        const type = req.params.category;
        const productsToScan = await sql`
            WITH RECURSIVE cat_tree AS (
                SELECT category_id FROM category WHERE category_id = ${type}
                UNION ALL
                SELECT c.category_id FROM category c 
                INNER JOIN cat_tree ct ON c.parent_id = ct.category_id
            )
            SELECT specs, name FROM product 
            WHERE category_id IN (SELECT category_id FROM cat_tree)
        `;

        if (!productsToScan || productsToScan.length === 0) {
            return res.status(404).json({ success: false, message: "Không tìm thấy sản phẩm để tạo bộ lọc" });
        }

        const filterData = {};
        const brands = [...new Set(productsToScan.map(product => {
            const specs = product.specs || {};
            const rawBrandArray = findValueInSpecs(specs, 'Hãng') || findValueInSpecs(specs, 'Thương hiệu');
            let brand = 'Other';

            if (Array.isArray(rawBrandArray) && rawBrandArray.length > 0) {
                brand = rawBrandArray[0];
            } else if (typeof rawBrandArray === 'string') {
                brand = rawBrandArray.split('.')[0];
            } else if (product.name) {
                brand = product.name.trim().split(' ')[0];
            }

            if (brand) {
                let cleanBrand = brand.toString().replace(/\s*\(.*?\)/g, '').replace(/[.,]/g, '').trim();
                if (cleanBrand.toLowerCase() === 'macbook') return 'Apple';
                return cleanBrand.charAt(0).toUpperCase() + cleanBrand.slice(1).toLowerCase();
            }
            return 'Other';
        }))].filter(b => b !== 'Other').sort();

        filterData.brands = { displayName: "Thương hiệu", options: brands };
        const allPotentialFields = getAllAvailableSpecFields(productsToScan);

        allPotentialFields.forEach(field => {
            if (isEligibleFilter(productsToScan, field)) {
                const values = getDynamicUniqueValues(productsToScan, field);
                if (values.length > 1 && values.length < productsToScan.length) {
                    const responseKey = field.toLowerCase()
                        .replace(/\s+/g,'_')
                        .normalize("NFD")
                        .replace(/[\u0300-\u036f]/g, "");

                    filterData[responseKey] = {
                        displayName: field,
                        options: values
                    };
                }
            }
        });

        res.status(200).json({
            success: true,
            filters: filterData
        });

    } catch (error) {
        console.error("Internal Server Error:", error);
        res.status(500).json({ success: false, message: error.message });
    }
});

router.post('/products/:category/filter', async (req, res) => {
    try {
        const pageRes = await axios.get(`${SERVER_URL}/api/pagination`);
        const pageData = pageRes.data.data[0];
        const limit = pageData.item_in_productlist;
        const column = pageData.column_in_productlist;

        const type = req.params.category;
        const page = parseInt(req.body.page) || 1;
        const offset = (page - 1) * limit;

        const sort = req.body.sort || 'newest';
        const filters = req.body.filters || {};
        const filterParams = [];
        const conditions = [];
        let pIdx = 2;

        if (filters.brands && Array.isArray(filters.brands) && filters.brands.length > 0) {
            const brandConds = filters.brands.map(b => {
                filterParams.push(`%${b}%`, `%${b}%`);
                const p1 = pIdx++;
                const p2 = pIdx++;
                return `(p.name ILIKE $${p1} OR p.specs::text ILIKE $${p2})`;
            });
            conditions.push(`(${brandConds.join(' OR ')})`);
        }

        Object.keys(filters).forEach(key => {
            if (key !== 'brands' && Array.isArray(filters[key]) && filters[key].length > 0) {
                const specConds = filters[key].map(val => {
                    filterParams.push(`%${val}%`);
                    const p = pIdx++;
                    return `p.specs::text ILIKE $${p}`;
                });
                conditions.push(`(${specConds.join(' OR ')})`);
            }
        });

        const filterWhere = conditions.length > 0 ? 'AND ' + conditions.join(' AND ') : '';
        const countParams = [type, ...filterParams];
        const countQuery = `
            WITH RECURSIVE cat_tree AS (
                SELECT category_id FROM category WHERE category_id = $1
                UNION ALL
                SELECT c.category_id FROM category c 
                INNER JOIN cat_tree ct ON c.parent_id = ct.category_id
            )
            SELECT COUNT(DISTINCT COALESCE(p.base_id, p.product_id)) as count
            FROM product p
            WHERE p.category_id IN (SELECT category_id FROM cat_tree)
            ${filterWhere}
        `;
        const countResult = await sql.unsafe(countQuery, countParams);
        const totalItems = parseInt(countResult[0].count, 10);
        const totalPages = Math.ceil(totalItems / limit);
        let sortOrderBy = 'MAX(p.product_id) DESC';
        if (sort === 'price_asc') sortOrderBy = 'MIN(p.price) ASC';
        else if (sort === 'price_desc') sortOrderBy = 'MIN(p.price) DESC';
        const limitIdx = pIdx++;
        const offsetIdx = pIdx++;
        const dataParams = [type, ...filterParams, limit, offset];

        const dataQuery = `
            WITH RECURSIVE cat_tree AS (
                SELECT category_id FROM category WHERE category_id = $1
                UNION ALL
                SELECT c.category_id FROM category c 
                INNER JOIN cat_tree ct ON c.parent_id = ct.category_id
            ),
            PaginatedBases AS (
                SELECT COALESCE(p.base_id, p.product_id) as group_id,
                       ROW_NUMBER() OVER(ORDER BY ${sortOrderBy}) as sort_index
                FROM product p
                WHERE p.category_id IN (SELECT category_id FROM cat_tree)
                ${filterWhere}
                GROUP BY COALESCE(p.base_id, p.product_id)
                ORDER BY sort_index
                LIMIT $${limitIdx}::int OFFSET $${offsetIdx}::int
            )
            SELECT p.*, 
                (
                    SELECT ROUND(COALESCE(AVG(r.rating), 4.9)::numeric, 1)
                    FROM reviews r
                    WHERE r.product_id = p.product_id AND r.status = 'approved'
                ) as rating,
                COALESCE(p.base_id, p.product_id) as group_id,
                pb.sort_index,
                MAX(prom.discount_percent) as discount_percent,
                COALESCE(
                    json_agg(
                        json_build_object(
                            'variant_id', v.variant_id,
                            'color_name', v.color_name,
                            'color_code', v.color_code,
                            'price', v.price,
                            'price_str', v.price_str,
                            'local_gallery', v.local_gallery,
                            'quantity', i.quantity,
                            'reserved', i.reserved,
                            'price_base', v.price_base
                        )
                    ) FILTER (WHERE v.variant_id IS NOT NULL), 
                    '[]'
                ) as variants
            FROM product p
            JOIN PaginatedBases pb ON COALESCE(p.base_id, p.product_id) = pb.group_id
            LEFT JOIN productvariant v ON p.product_id = v.product_id
            LEFT JOIN productinventory i ON v.variant_id = i.variant_id
            LEFT JOIN promotions prom ON p.product_id = prom.product_id AND NOW() BETWEEN prom.start_date AND prom.end_date AND prom.is_deleted IS NOT TRUE
            GROUP BY p.product_id, pb.group_id, pb.sort_index
            ORDER BY pb.sort_index ASC, p.price ASC
        `;

        const products = await sql.unsafe(dataQuery, dataParams);
        const formattedProducts = products.map(formatProductPaths);
        const groupedData = [];
        const groupMap = new Map();

        formattedProducts.forEach(prod => {
            const groupId = prod.group_id;
            const { group_id, sort_index, ...productData } = prod;

            if (!groupMap.has(groupId)) {
                groupMap.set(groupId, {
                    ...productData,
                    versions: [productData]
                });
                groupedData.push(groupMap.get(groupId));
            } else {
                groupMap.get(groupId).versions.push(productData);
            }
        });

        res.status(200).json({
            success: true,
            data: groupedData,
            pagination: {
                totalItems,
                totalPages,
                currentPage: page,
                limit,
                column:column
            }
        });
    } catch (error) {
        console.error("Lỗi API Lọc Sản Phẩm:", error.message);
        res.status(500).json({ success: false, message: error.message });
    }
});

router.post('/import', async (req, res) => {
    const { supplier, items, note } = req.body;
    if (!supplier || !supplier.name) {
        return res.status(400).json({ message: "Vui lòng cung cấp tên nhà cung cấp!" });
    }
    if (!items || items.length === 0) {
        return res.status(400).json({ message: "Danh sách sản phẩm nhập không được trống!" });
    }

    try {
        await sql.begin(async (sql) => {
            let supplierId;
            const existingSuppliers = await sql`
                SELECT supplier_id FROM suppliers
                WHERE phone = ${supplier.phone || ''} AND phone != ''
                   OR name = ${supplier.name}
                    LIMIT 1
            `;

            if (existingSuppliers.length > 0) {
                supplierId = existingSuppliers[0].supplier_id;
            } else {
                const newSupplier = await sql`
                    INSERT INTO suppliers (name, address, phone)
                    VALUES (${supplier.name}, ${supplier.address || ''}, ${supplier.phone || ''})
                        RETURNING supplier_id
                `;
                supplierId = newSupplier[0].supplier_id;
            }

            const totalItemsQty = items.reduce((sum, item) => sum + (parseInt(item.quantity) || 0), 0);

            const newInvoice = await sql`
                INSERT INTO import_invoices (supplier_id, total_items, note)
                VALUES (${supplierId}, ${totalItemsQty}, ${note || 'Nhập kho từ hệ thống quản trị'})
                    RETURNING invoice_id
            `;
            const invoiceId = newInvoice[0].invoice_id;

            for (const item of items) {
                const importQty = parseInt(item.quantity) || 0;
                const importPrice = parseInt(item.priceImport) || 0;
                const totalImportValue = importQty * importPrice; // Tính tổng tiền lô này

                await sql`
                    INSERT INTO import_invoice_details
                    (invoice_id, product_id, variant_id, quantity, current_stock_before, price_import)
                    VALUES
                        (${invoiceId}, ${item.productId}, ${item.variantId || null}, ${importQty}, ${item.currentStock || 0}, ${importPrice})
                `;

                // Tìm variantId tương ứng để tính toán và cập nhật giá vốn (price_base)
                let targetVariantId = item.variantId;
                if (!targetVariantId) {
                    const existingVariants = await sql`
                        SELECT variant_id FROM productvariant
                        WHERE product_id = ${item.productId}
                        LIMIT 1
                    `;
                    if (existingVariants.length > 0) {
                        targetVariantId = existingVariants[0].variant_id;
                    }
                }

                if (targetVariantId) {
                    // 1. Lấy giá vốn (price_base) hiện tại của variant
                    const [variantRecord] = await sql`
                        SELECT price_base FROM productvariant
                        WHERE variant_id = ${targetVariantId}
                    `;
                    const oldPriceBase = variantRecord ? (parseInt(variantRecord.price_base) || 0) : 0;

                    // 2. Lấy số lượng tồn kho hiện tại của variant
                    const [inventoryRecord] = await sql`
                        SELECT quantity FROM productinventory
                        WHERE variant_id = ${targetVariantId}
                    `;
                    const oldQty = inventoryRecord ? (parseInt(inventoryRecord.quantity) || 0) : 0;

                    // 3. Tính giá vốn bình quân gia quyền mới
                    const newQty = oldQty + importQty;
                    let newPriceBase = 0;
                    if (newQty > 0) {
                        newPriceBase = Math.round(((oldPriceBase * oldQty) + (importPrice * importQty)) / newQty);
                    } else {
                        newPriceBase = importPrice;
                    }

                    // 4. Cập nhật price_base của variant
                    await sql`
                        UPDATE productvariant
                        SET price_base = ${newPriceBase}
                        WHERE variant_id = ${targetVariantId}
                    `;

                    // 5. Cập nhật kho
                    await sql`
                        INSERT INTO productinventory (variant_id, quantity, total_value, last_updated)
                        VALUES (${targetVariantId}, ${importQty}, ${totalImportValue}, CURRENT_TIMESTAMP)
                            ON CONFLICT (variant_id) 
                        DO UPDATE SET
                            quantity = productinventory.quantity + EXCLUDED.quantity,
                            total_value = COALESCE(productinventory.total_value, 0) + EXCLUDED.total_value,
                            last_updated = CURRENT_TIMESTAMP
                    `;
                }
            }
        });

        res.status(200).json({ message: "Nhập kho thành công và đã cập nhật giá vốn!" });
    } catch (err) {
        console.error('Lỗi Import Stock:', err);
        res.status(500).json({ message: "Lỗi hệ thống khi nhập kho!", error: err.message });
    }
});

router.put('/update-price', async (req, res) => {
    const { updates } = req.body;
    if (!updates || !Array.isArray(updates) || updates.length === 0) {
        return res.status(400).json({ message: "Dữ liệu cập nhật giá không hợp lệ!" });
    }

    try {
        await sql.begin(async (sql) => {
            for (const update of updates) {
                const { productId, variantId, newPrice } = update;

                if (variantId) {
                    // Cập nhật giá bán biến thể
                    await sql`
                        UPDATE productvariant
                        SET price = ${newPrice}
                        WHERE variant_id = ${variantId}
                    `;
                } else if (productId) {
                    // Cập nhật giá bán sản phẩm
                    await sql`
                        UPDATE product
                        SET price = ${newPrice}
                        WHERE product_id = ${productId}
                    `;
                }
            }
        });
        res.status(200).json({ message: "Cập nhật giá bán thành công!" });
    } catch (err) {
        console.error('Lỗi Update Price:', err);
        res.status(500).json({ message: "Lỗi hệ thống khi cập nhật giá bán!", error: err.message });
    }
});

router.post('/save', async (req, res) => {
    try {
        const { id, name, category, price, weight, specs, variants, low_stock_threshold, base_price_numeric } = req.body;

        if (!name || !category) {
            return res.status(400).json({ success: false, message: "Vui lòng cung cấp tên và danh mục sản phẩm!" });
        }

        const parsedSpecs = specs ? JSON.parse(specs) : {};
        const parsedVariants = variants ? JSON.parse(variants) : [];
        const numericPrice = price ? parseInt(price, 10) : 0;
        const numericWeight = weight ? parseFloat(weight) : null;
        const numericLowStock = low_stock_threshold !== undefined && low_stock_threshold !== '' ? parseInt(low_stock_threshold, 10) : 5;
        const numericBasePrice = base_price_numeric !== undefined && base_price_numeric !== '' ? parseFloat(base_price_numeric) : numericPrice;

        let thumbPath = null;
        if (req.files && Array.isArray(req.files)) {
            const thumbFile = req.files.find(f => f.fieldname === 'thumbnail');
            if (thumbFile) {
                thumbPath = `/uploads/${thumbFile.filename}`;
            }
        }

        await sql.begin(async (sql) => {
            let productId = id;
            if (productId) {
                const updateFields = {
                    name: name, category_id: category, price: numericPrice,
                    base_price_numeric: numericBasePrice, weight: numericWeight, specs: parsedSpecs,
                    low_stock_threshold: numericLowStock
                };
                if (thumbPath) updateFields.img_thumb = thumbPath;
                await sql`UPDATE product SET ${sql(updateFields)} WHERE product_id = ${productId}`;
            } else {
                productId = `PROD-${Date.now()}`;
                await sql`
                    INSERT INTO product (product_id, name, category_id, price, base_price_numeric, weight, specs, img_thumb, low_stock_threshold)
                    VALUES (${productId}, ${name}, ${category}, ${numericPrice}, ${numericBasePrice}, ${numericWeight}, ${parsedSpecs}, ${thumbPath}, ${numericLowStock})
                `;
            }

            for (let i = 0; i < parsedVariants.length; i++) {
                const v = parsedVariants[i];
                let variantId = v.id;
                const vPrice = v.price ? parseInt(v.price, 10) : numericPrice;
                const vStock = v.stock ? parseInt(v.stock, 10) : 0;
                const vPriceBase = v.price_base ? parseInt(v.price_base, 10) : 0;
                const initialTotalValue = vStock * vPrice; // Giá vốn ban đầu

                const isNewVariant = !variantId || (typeof variantId === 'string' && variantId.startsWith('v-'));

                if (isNewVariant) {
                    const newVar = await sql`
                        INSERT INTO productvariant (product_id, color_name, price, price_base)
                        VALUES (${productId}, ${v.variant_name || 'Mặc định'}, ${vPrice}, ${vPriceBase}) RETURNING variant_id
                    `;
                    variantId = newVar[0].variant_id;

                    await sql`
                        INSERT INTO productinventory (variant_id, quantity, total_value, last_updated)
                        VALUES (${variantId}, ${vStock}, ${initialTotalValue}, CURRENT_TIMESTAMP)
                    `;
                } else {
                    await sql`
                        UPDATE productvariant SET color_name = ${v.variant_name}, price = ${vPrice}, price_base = ${vPriceBase}
                        WHERE variant_id = ${variantId} AND product_id = ${productId}
                    `;
                    await sql`
                        INSERT INTO productinventory (variant_id, quantity, total_value, last_updated)
                        VALUES (${variantId}, ${vStock}, ${initialTotalValue}, CURRENT_TIMESTAMP)
                            ON CONFLICT (variant_id) 
                        DO UPDATE SET
                            quantity = ${vStock},
                            total_value = ${initialTotalValue},
                            last_updated = CURRENT_TIMESTAMP
                    `;
                }
            }
        });
        return res.status(200).json({ success: true, message: id ? "Sửa sản phẩm thành công!" : "Lưu sản phẩm mới thành công!" });
    } catch (err) {
        console.error('Lỗi API Save Product:', err);
        return res.status(500).json({ success: false, message: "Lỗi hệ thống khi lưu sản phẩm!", error: err.message });
    }
});

router.post('/search', async (req, res) => {
    const { keyword } = req.body;

    if (!keyword) {
        return res.status(400).json({ message: "Vui lòng nhập từ khóa tìm kiếm" });
    }

    try {
        // Query the column configuration for search layout
        const pageDataList = await sql`SELECT column_in_productlist FROM pagenigation LIMIT 1`;
        const column = pageDataList.length > 0 ? pageDataList[0].column_in_productlist : 4;

        // Bước 1: Tìm tất cả các group_id (base_id) có chứa sản phẩm khớp từ khóa
        const products = await sql`
            WITH MatchedGroups AS (
                SELECT DISTINCT COALESCE(p.base_id, p.product_id) as group_id
                FROM product p
                WHERE p.name ILIKE ${'%' + keyword + '%'}
                OR p.product_id = ${keyword}
                LIMIT 20
                )
            SELECT
                p.*,
                (
                    SELECT ROUND(COALESCE(AVG(r.rating), 4.9)::numeric, 1)
                    FROM reviews r
                    WHERE r.product_id = p.product_id AND r.status = 'approved'
                ) as rating,
                mg.group_id,
                MAX(prom.discount_percent) as discount_percent,
                COALESCE(
                        json_agg(
                                json_build_object(
                                        'variant_id', v.variant_id,
                                        'color_name', v.color_name,
                                        'color_code', v.color_code,
                                        'price', v.price,
                                        'price_str', v.price_str,
                                        'local_gallery', v.local_gallery,
                                        'quantity', i.quantity,
                                        'reserved', i.reserved,
                                        'price_base', v.price_base
                                )
                        ) FILTER (WHERE v.variant_id IS NOT NULL),
                        '[]'
                ) as variants
            FROM product p
                     JOIN MatchedGroups mg ON COALESCE(p.base_id, p.product_id) = mg.group_id
                     LEFT JOIN productvariant v ON p.product_id = v.product_id
                     LEFT JOIN productinventory i ON v.variant_id = i.variant_id
                     LEFT JOIN promotions prom ON p.product_id = prom.product_id AND NOW() BETWEEN prom.start_date AND prom.end_date AND prom.is_deleted IS NOT TRUE
            GROUP BY p.product_id, mg.group_id
            ORDER BY mg.group_id DESC, p.price ASC
        `;

        // Bước 2: Format đường dẫn ảnh (nếu mày có hàm formatProductPaths)
        const formattedProducts = typeof formatProductPaths === 'function'
            ? products.map(formatProductPaths)
            : products;

        // Bước 3: Gom nhóm các Product thành các Card (giống logic all-products)
        const groupedData = [];
        const groupMap = new Map();

        formattedProducts.forEach(prod => {
            const groupId = prod.group_id;
            const { group_id, ...productData } = prod;

            if (!groupMap.has(groupId)) {
                const newCard = {
                    ...productData,
                    group_id: groupId,
                    versions: [productData]
                };
                groupMap.set(groupId, newCard);
                groupedData.push(newCard);
            } else {
                groupMap.get(groupId).versions.push(productData);
            }
        });

        res.status(200).json({
            success: true,
            data: groupedData,
            column: column
        });

    } catch (e) {
        console.error("Lỗi Search nâng cao:", e.message);
        res.status(500).json({ success: false, message: "Lỗi hệ thống khi tìm kiếm" });
    }
});

module.exports = router;