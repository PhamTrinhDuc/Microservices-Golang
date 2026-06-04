#!/usr/bin/env python3
import os
import json
import re
import random
import sys
from pathlib import Path
import psycopg2
from psycopg2 import extras



# Database Connection Settings
DB_HOST = os.getenv("DB_HOST", "localhost")
DB_PORT = os.getenv("DB_PORT", "5433")
DB_USER = os.getenv("DB_USER", "jiyuu_user")
DB_PASSWORD = os.getenv("DB_PASSWORD", "jiyuu_password")
DB_NAME = os.getenv("DB_NAME", "ecommerce_db")

# Setup project directories
BASE_DIR = Path(__file__).resolve().parent.parent
DATA_DIR = BASE_DIR / "data"

def get_db_connection():
    """Establish a connection to the PostgreSQL database."""
    print(f"Connecting to database {DB_NAME} at {DB_HOST}:{DB_PORT}...")
    try:
        conn = psycopg2.connect(
            host=DB_HOST,
            port=DB_PORT,
            user=DB_USER,
            password=DB_PASSWORD,
            dbname=DB_NAME
        )
        return conn
    except Exception as e:
        print(f"Error: Unable to connect to the database. Detail: {e}")
        sys.exit(1)

def slugify(value):
    """
    Generate an ASCII-friendly slug for a string, 
    properly cleaning up accented Vietnamese characters.
    """
    vietnamese_map = {
        'à':'a','á':'a','ả':'a','ã':'a','ạ':'a','ă':'a','ằ':'a','ắ':'a','ẳ':'a','ẵ':'a','ặ':'a','â':'a','ầ':'a','ấ':'a','ẩ':'a','ẫ':'a','ậ':'a',
        'đ':'d',
        'è':'e','é':'e','ẻ':'e','ẽ':'e','ẹ':'e','ê':'e','ề':'e','ế':'e','ể':'e','ễ':'e','ệ':'e',
        'ì':'i','í':'i','ỉ':'i','ĩ':'i','ị':'i',
        'ò':'o','ó':'o','ỏ':'o','õ':'o','ọ':'o','ô':'o','ồ':'o','ố':'o','ổ':'o','ỗ':'o','ộ':'o','ơ':'o','ờ':'o','ớ':'o','ở':'o','ỡ':'o','ợ':'o',
        'ù':'u','ú':'u','ủ':'u','ũ':'u','ụ':'u','ư':'u','ừ':'u','ứ':'u','ử':'u','ữ':'u','ự':'u',
        'ỳ':'y','ý':'y','ỷ':'y','ỹ':'y','ỵ':'y',
        'À':'A','Á':'A','Ả':'A','Ã':'A','Ạ':'A','Ă':'A','Ằ':'A','Ắ':'A','Ẳ':'A','Ẵ':'A','Ặ':'A','Â':'A','Ầ':'A','Ấ':'A','Ẩ':'A','Ẫ':'A','Ậ':'A',
        'Đ':'D',
        'È':'E','É':'E','Ẻ':'E','Ẽ':'E','Ẹ':'E','Ê':'E','Ề':'E','Ế':'E','Ể':'E','Ễ':'E','Ệ':'E',
        'Ì':'I','Í':'I','Ỉ':'I','Ĩ':'I','Ị':'I',
        'Ò':'O','Ó':'O','Ỏ':'O','Õ':'O','Ọ':'O','Ô':'O','Ồ':'O','Ố':'O','Ổ':'O','Ỗ':'O','Ộ':'O','Ơ':'O','Ờ':'O','Ớ':'O','Ở':'O','Ỡ':'O','Ợ':'O',
        'Ù':'U','Ú':'U','Ủ':'U','Ũ':'U','Ụ':'U','Ư':'U','Ừ':'U','Ứ':'U','Ử':'U','Ữ':'U','Ự':'U',
        'Ỳ':'Y','Ý':'Y','Ỷ':'Y','Ỹ':'Y','Ỵ':'Y'
    }
    res = []
    for c in str(value):
        res.append(vietnamese_map.get(c, c))
    value = "".join(res)
    value = value.lower()
    value = value.replace('+', ' plus')
    value = re.sub(r'[^\w\s-]', '', value)
    value = re.sub(r'[-\s]+', '-', value)
    return value.strip('-')

def load_json_file(file_path):
    """Safely load and return JSON file contents."""
    if not file_path.exists():
        print(f"Warning: File not found at {file_path}")
        return None
    try:
        with open(file_path, "r", encoding="utf-8") as f:
            return json.load(f)
    except Exception as e:
        print(f"Error reading {file_path.name}: {e}")
        return None

def seed_statuses(cursor):
    """Seed order, payment, and shipping status tables."""
    print("Seeding order, payment, and shipping status reference tables...")
    
    # 1. Order Statuses
    order_statuses = [
        ('pending', 'Chờ xử lý', 1),
        ('confirmed', 'Đã xác nhận', 2),
        ('processing', 'Đang xử lý', 3),
        ('shipping', 'Đang giao hàng', 4),
        ('delivered', 'Đã giao hàng', 5),
        ('cancelled', 'Đã hủy', 6)
    ]
    cursor.executemany("""
        INSERT INTO order_status (code, label, sort_order) 
        VALUES (%s, %s, %s)
        ON CONFLICT (code) DO NOTHING;
    """, order_statuses)
    
    # 2. Payment Statuses
    payment_statuses = [
        ('unpaid', 'Chưa thanh toán', 1),
        ('paid', 'Đã thanh toán', 2),
        ('refunded', 'Đã hoàn tiền', 3)
    ]
    cursor.executemany("""
        INSERT INTO payment_status (code, label, sort_order) 
        VALUES (%s, %s, %s)
        ON CONFLICT (code) DO NOTHING;
    """, payment_statuses)
    
    # 3. Shipping Statuses
    shipping_statuses = [
        ('not_shipped', 'Chưa giao hàng', 1),
        ('shipped', 'Đang giao hàng', 2),
        ('delivered', 'Đã giao hàng', 3)
    ]
    cursor.executemany("""
        INSERT INTO shipping_status (code, label, sort_order) 
        VALUES (%s, %s, %s)
        ON CONFLICT (code) DO NOTHING;
    """, shipping_statuses)

def ingest_stores(cursor, stores_data):
    """Ingest stores from stores_data into store table, returning a list of store IDs."""
    if not stores_data:
        print("No store data to ingest.")
        return []
    
    print(f"Ingesting {len(stores_data)} stores...")
    store_ids = []
    
    for store in stores_data:
        # Clean address details
        name = (store.get("name") or "").strip()
        hotline = (store.get("hotline") or "18001060").strip()
        province = (store.get("province") or "Hồ Chí Minh").strip()
        ward = (store.get("ward") or "").strip()
        road = (store.get("road") or "").strip()
        lat = store.get("lat")
        lng = store.get("lng")
        is_active = store.get("is_active", True)
        
        # Approximate district from name or road if missing, or use a default
        district = (store.get("district") or "").strip()
        if not district:
            # Simple heuristic: try to extract 'Quận/Huyện' from road
            match = re.search(r'(Quận\s+\d+|Quận\s+\w+|Huyện\s+\w+)', road, re.IGNORECASE)
            district = match.group(1) if match else "Quận 1"
            
        # Check if store already exists by name and road
        cursor.execute("SELECT id FROM store WHERE name = %s AND road = %s;", (name, road))
        row = cursor.fetchone()
        if row:
            store_ids.append(row[0])
        else:
            cursor.execute("""
                INSERT INTO store (name, hotline, district, province, ward, road, lat, lng, is_active)
                VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s)
                RETURNING id;
            """, (name, hotline, district, province, ward, road, lat, lng, is_active))
            store_ids.append(cursor.fetchone()[0])
            
    print(f"Successfully processed store table. Total stores: {len(store_ids)}")
    return store_ids

def ingest_products_catalog(cursor, products_data, store_ids):
    """Ingest products, brands, categories, specifications, variants, and options."""
    if not products_data:
        print("No product catalog data to ingest.")
        return
    
    print(f"Ingesting {len(products_data)} products and generating variants...")
    
    # 1. Ingest/Look up standard Admin User for Invoices
    cursor.execute("SELECT id FROM users WHERE role = 'admin' LIMIT 1;")
    admin_row = cursor.fetchone()
    if admin_row:
        admin_id = admin_row[0]
    else:
        cursor.execute("""
            INSERT INTO users (full_name, email, password, role, is_verified)
            VALUES ('Admin Manager', 'admin@tgdd-ecommerce.com', 'admin_pwd_hash', 'admin', TRUE)
            RETURNING id;
        """)
        admin_id = cursor.fetchone()[0]
        
    # 2. Mock Suppliers
    suppliers = [
        {"name": "Apple Vietnam LLC", "address": "Quận 1, TP. Hồ Chí Minh", "phone": "18001122"},
        {"name": "OPPO Vietnam Distribution", "address": "Quận 3, TP. Hồ Chí Minh", "phone": "18003344"},
        {"name": "AVA+ Accessories Corp", "address": "Quận Tân Bình, TP. Hồ Chí Minh", "phone": "18005566"},
    ]
    supplier_ids = []
    for sup in suppliers:
        cursor.execute("SELECT id FROM suppliers WHERE name = %s;", (sup['name'],))
        row = cursor.fetchone()
        if row:
            sup_id = row[0]
        else:
            cursor.execute("INSERT INTO suppliers (name, address, phone) VALUES (%s, %s, %s) RETURNING id;", (sup['name'], sup['address'], sup['phone']))
            sup_id = cursor.fetchone()[0]
        supplier_ids.append(sup_id)

    # Cache for categories and brands
    brand_cache = {}
    category_cache = {}
    
    # Track variants created to seed inventory later
    created_variants = []

    # Load existing product slug mappings from database to handle duplicate slugs
    cursor.execute("SELECT id, slug FROM product;")
    existing_slugs = {row[1]: str(row[0]) for row in cursor.fetchall()}

    for item in products_data:
        info = item.get("product_info", {})
        if not info:
            continue
            
        p_id = str(info.get("id"))
        brand_name = (info.get("brand_name") or "Khác").strip()
        category_name = (item.get("category_name") or "Điện thoại").strip()
        
        # Ingest Brand
        if brand_name not in brand_cache:
            brand_slug = slugify(brand_name)
            cursor.execute("""
                INSERT INTO brand (name, slug, is_active)
                VALUES (%s, %s, TRUE)
                ON CONFLICT (slug) DO UPDATE SET name = EXCLUDED.name
                RETURNING id;
            """, (brand_name, brand_slug))
            brand_cache[brand_name] = cursor.fetchone()[0]
        brand_id = brand_cache[brand_name]
        
        # Ingest Category
        if category_name not in category_cache:
            category_slug = slugify(category_name)
            cursor.execute("""
                INSERT INTO category (name, slug)
                VALUES (%s, %s)
                ON CONFLICT (slug) DO UPDATE SET name = EXCLUDED.name
                RETURNING id;
            """, (category_name, category_slug))
            category_cache[category_name] = cursor.fetchone()[0]
        category_id = category_cache[category_name]
        
        # Ingest Product
        clean_name = (info.get("clean_name") or info.get("raw_name") or "").strip()
        base_slug = (info.get("slug") or slugify(clean_name)).strip()
        
        # Resolve duplicate slugs for different product IDs
        slug = base_slug
        counter = 1
        while slug in existing_slugs and existing_slugs[slug] != p_id:
            slug = f"{base_slug}-{counter}"
            counter += 1
        existing_slugs[slug] = p_id
        
        meta_title = (info.get("meta_title") or "").strip()
        meta_description = (info.get("meta_description") or "").strip()
        img_thumb = (info.get("img_thumb") or "").strip()
        weight = float(info.get("weight") or 180.0)
        is_active = info.get("is_active", True)
        
        # Convert specifications to JSON string
        specs_json = json.dumps(item.get("specifications", {}))
        
        cursor.execute("""
            INSERT INTO product (id, category_id, brand_id, name, slug, meta_title, meta_description, img_thumb, weight, specs_jsonb, is_active)
            VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s)
            ON CONFLICT (id) DO UPDATE SET
                category_id = EXCLUDED.category_id,
                brand_id = EXCLUDED.brand_id,
                name = EXCLUDED.name,
                slug = EXCLUDED.slug,
                meta_title = EXCLUDED.meta_title,
                meta_description = EXCLUDED.meta_description,
                img_thumb = EXCLUDED.img_thumb,
                weight = EXCLUDED.weight,
                specs_jsonb = EXCLUDED.specs_jsonb,
                is_active = EXCLUDED.is_active
            RETURNING id;
        """, (p_id, category_id, brand_id, clean_name, slug, meta_title, meta_description, img_thumb, weight, specs_json, is_active))
        
        # Ingest Gallery Images
        cursor.execute("DELETE FROM product_image WHERE product_id = %s AND variant_id IS NULL;", (p_id,))
        gallery_images = item.get("gallery_images", [])
        for img in gallery_images:
            img_url = (img.get("url") or "").strip()
            if not img_url:
                continue
            alt_text = (img.get("alt_text") or clean_name).strip()
            sort_order = int(img.get("sort_order") or 0)
            is_thumb_flag = bool(img.get("is_thumbnail") or False)
            
            cursor.execute("""
                INSERT INTO product_image (product_id, variant_id, url, alt_text, sort_order, is_thumbnail)
                VALUES (%s, NULL, %s, %s, %s, %s);
            """, (p_id, img_url, alt_text, sort_order, is_thumb_flag))
            
        # Ingest Specifications (Product Specs EAV table)
        cursor.execute("DELETE FROM product_spec WHERE product_id = %s;", (p_id,))
        specs = item.get("specifications", {})
        for group_name, group_data in specs.items():
            if not isinstance(group_data, dict):
                continue
            for key, val_dict in group_data.items():
                if not isinstance(val_dict, dict):
                    continue
                raw_val = (val_dict.get("raw_value") or "").strip()
                val = (val_dict.get("value") or raw_val).strip()
                unit = (val_dict.get("unit") or "").strip()
                
                # Truncate strings to match DB schema limits (keeping group, key, and unit limits)
                group_truncated = group_name[:100].strip()
                key_truncated = key[:100].strip()
                unit_truncated = unit[:50].strip()
                
                cursor.execute("""
                    INSERT INTO product_spec (product_id, "group", key, value, unit)
                    VALUES (%s, %s, %s, %s, %s);
                """, (p_id, group_truncated, key_truncated, val, unit_truncated))
                
        # Ingest Option Types and Value Mapping
        # Setup standard "Màu sắc" option
        cursor.execute("SELECT id FROM product_option_type WHERE product_id = %s AND name = %s;", (p_id, "Màu sắc"))
        row_color = cursor.fetchone()
        if row_color:
            color_type_id = row_color[0]
        else:
            cursor.execute("INSERT INTO product_option_type (product_id, name, sort_order) VALUES (%s, 'Màu sắc', 0) RETURNING id;", (p_id,))
            color_type_id = cursor.fetchone()[0]
            
        # Setup standard "Dung lượng" option
        cursor.execute("SELECT id FROM product_option_type WHERE product_id = %s AND name = %s;", (p_id, "Dung lượng"))
        row_cap = cursor.fetchone()
        if row_cap:
            capacity_type_id = row_cap[0]
        else:
            cursor.execute("INSERT INTO product_option_type (product_id, name, sort_order) VALUES (%s, 'Dung lượng', 1) RETURNING id;", (p_id,))
            capacity_type_id = cursor.fetchone()[0]

        # Ingest Option Values
        color_val_ids = {}
        for col in item.get("colors", []):
            c_label = (col.get("label") or "").strip()
            if not c_label:
                continue
            c_code = col.get("bg_color") or ""
            cursor.execute("SELECT id FROM product_option_value WHERE option_type_id = %s AND value = %s;", (color_type_id, c_label))
            row_val = cursor.fetchone()
            if row_val:
                c_val_id = row_val[0]
            else:
                cursor.execute("INSERT INTO product_option_value (option_type_id, value, color_code) VALUES (%s, %s, %s) RETURNING id;", (color_type_id, c_label, c_code))
                c_val_id = cursor.fetchone()[0]
            color_val_ids[c_label] = c_val_id
            
        capacity_val_ids = {}
        for cap in item.get("capacities", []):
            cap_label = (cap.get("label") or "").strip()
            if not cap_label:
                continue
            cursor.execute("SELECT id FROM product_option_value WHERE option_type_id = %s AND value = %s;", (capacity_type_id, cap_label))
            row_val = cursor.fetchone()
            if row_val:
                cap_val_id = row_val[0]
            else:
                cursor.execute("INSERT INTO product_option_value (option_type_id, value) VALUES (%s, %s) RETURNING id;", (capacity_type_id, cap_label))
                cap_val_id = cursor.fetchone()[0]
            capacity_val_ids[cap_label] = cap_val_id

        # Generate variants by cross-matching colors and capacities
        colors_list = list(color_val_ids.keys()) if color_val_ids else ["Mặc định"]
        capacities_list = list(capacity_val_ids.keys()) if capacity_val_ids else ["Mặc định"]
        
        # Ensure at least option placeholder exists
        if not color_val_ids:
            cursor.execute("SELECT id FROM product_option_value WHERE option_type_id = %s AND value = 'Mặc định';", (color_type_id,))
            col_def = cursor.fetchone()
            col_def_id = col_def[0] if col_def else None
            if not col_def_id:
                cursor.execute("INSERT INTO product_option_value (option_type_id, value) VALUES (%s, 'Mặc định') RETURNING id;", (color_type_id,))
                col_def_id = cursor.fetchone()[0]
            color_val_ids["Mặc định"] = col_def_id
            
        if not capacity_val_ids:
            cursor.execute("SELECT id FROM product_option_value WHERE option_type_id = %s AND value = 'Mặc định';", (capacity_type_id,))
            cap_def = cursor.fetchone()
            cap_def_id = cap_def[0] if cap_def else None
            if not cap_def_id:
                cursor.execute("INSERT INTO product_option_value (option_type_id, value) VALUES (%s, 'Mặc định') RETURNING id;", (capacity_type_id,))
                cap_def_id = cursor.fetchone()[0]
            capacity_val_ids["Mặc định"] = cap_def_id

        base_price = float(info.get("price") or 0.0)
        if base_price == 0.0:
            base_price = 150000.0  # default base price for missing

        for cap_label in capacities_list:
            for col_label in colors_list:
                # Generate variant name & SKU
                var_name = clean_name
                if cap_label != "Mặc định":
                    var_name += f" {cap_label}"
                if col_label != "Mặc định":
                    var_name += f" - {col_label}"
                
                sku_cap = slugify(cap_label)
                sku_col = slugify(col_label)
                v_sku = f"{p_id}-{sku_cap}-{sku_col}"
                
                # Mock variant price depending on storage size
                var_price = base_price
                if "512" in cap_label:
                    var_price += 3000000.0
                elif "1tb" in cap_label.lower():
                    var_price += 7000000.0
                elif "128" in cap_label and "256" in clean_name:
                    var_price -= 2000000.0
                
                var_price_base = var_price * 1.15  # 15% discount mock
                
                # Ingest variant
                cursor.execute("""
                    INSERT INTO product_variant (product_id, name, sku, sell_price, compare_price, weight, is_active)
                    VALUES (%s, %s, %s, %s, %s, %s, TRUE)
                    ON CONFLICT (sku) DO UPDATE SET
                        name = EXCLUDED.name,
                        sell_price = EXCLUDED.sell_price,
                        compare_price = EXCLUDED.compare_price,
                        weight = EXCLUDED.weight
                    RETURNING id;
                """, (p_id, var_name, v_sku, var_price, var_price_base, weight))
                var_id = cursor.fetchone()[0]
                created_variants.append((var_id, var_price))
                
                # Link Variant Options
                col_val_id = color_val_ids[col_label]
                cap_val_id = capacity_val_ids[cap_label]
                
                cursor.execute("""
                    INSERT INTO product_variant_option (variant_id, option_value_id)
                    VALUES (%s, %s)
                    ON CONFLICT (variant_id, option_value_id) DO NOTHING;
                """, (var_id, col_val_id))
                
                cursor.execute("""
                    INSERT INTO product_variant_option (variant_id, option_value_id)
                    VALUES (%s, %s)
                    ON CONFLICT (variant_id, option_value_id) DO NOTHING;
                """, (var_id, cap_val_id))

                # Ingest Gallery Thumbnail for variant
                if img_thumb:
                    cursor.execute("SELECT id FROM product_image WHERE product_id = %s AND url = %s;", (p_id, img_thumb))
                    if not cursor.fetchone():
                        cursor.execute("""
                            INSERT INTO product_image (product_id, variant_id, url, alt_text, sort_order, is_thumbnail)
                            VALUES (%s, %s, %s, %s, 0, TRUE);
                        """, (p_id, var_id, img_thumb, var_name))
                        
    # 3. Seed stock levels via Import Invoices across stores
    if store_ids and created_variants:
        print(f"Mocking product inventory and import invoices across {len(store_ids)} stores...")
        for s_id in store_ids:
            # Pick a random distributor
            sup_id = random.choice(supplier_ids)
            
            # Check/Create invoice
            cursor.execute("SELECT id FROM import_invoices WHERE supplier_id = %s AND store_id = %s LIMIT 1;", (sup_id, s_id))
            inv_row = cursor.fetchone()
            if inv_row:
                invoice_id = inv_row[0]
            else:
                cursor.execute("""
                    INSERT INTO import_invoices (supplier_id, store_id, created_by, total_items, note)
                    VALUES (%s, %s, %s, 0, 'Nhập hàng đầu kỳ phục vụ kinh doanh cửa hàng')
                    RETURNING id;
                """, (sup_id, s_id, admin_id))
                invoice_id = cursor.fetchone()[0]
                
            total_qty = 0
            for var_id, price in created_variants:
                qty = random.randint(15, 60)
                
                # Ingest Inventory
                cursor.execute("""
                    INSERT INTO product_inventory (variant_id, store_id, quantity, reserved, last_updated)
                    VALUES (%s, %s, %s, 0, NOW())
                    ON CONFLICT (variant_id, store_id) DO UPDATE SET quantity = EXCLUDED.quantity, last_updated = NOW();
                """, (var_id, s_id, qty))
                
                # Check/Create details
                cursor.execute("SELECT id FROM import_invoice_details WHERE invoice_id = %s AND variant_id = %s;", (invoice_id, var_id))
                if not cursor.fetchone():
                    import_price = float(price) * 0.8  # Wholesale cost
                    cursor.execute("""
                        INSERT INTO import_invoice_details (invoice_id, variant_id, quantity, stock_before, price_import)
                        VALUES (%s, %s, %s, 0, %s);
                    """, (invoice_id, var_id, qty, import_price))
                    total_qty += qty
                else:
                    import_price = float(price) * 0.8

                # Sync latest_cost_price for variant to keep it consistent with the invoice detail
                cursor.execute("""
                    UPDATE product_variant SET latest_cost_price = %s WHERE id = %s;
                """, (import_price, var_id))
            
            # Update total items sum on invoice
            cursor.execute("UPDATE import_invoices SET total_items = total_items + %s WHERE id = %s;", (total_qty, invoice_id))

def ingest_reviews_and_users(cursor, reviews_data, store_ids):
    """
    Ingest user reviews, mapping reviewer names to new user and address entries, 
    and generating historical completed orders to fulfill database constraints.
    """
    if not reviews_data:
        print("No review data to ingest.")
        return
        
    # Normalize reviews_data to a list of dicts
    if isinstance(reviews_data, dict):
        reviews_data_list = [reviews_data]
    elif isinstance(reviews_data, list):
        reviews_data_list = reviews_data
    else:
        print("Invalid review data format.")
        return
        
    # Retrieve status mappings
    cursor.execute("SELECT code, id FROM order_status;")
    order_status_map = {row[0]: row[1] for row in cursor.fetchall()}
    cursor.execute("SELECT code, id FROM payment_status;")
    payment_status_map = {row[0]: row[1] for row in cursor.fetchall()}
    cursor.execute("SELECT code, id FROM shipping_status;")
    shipping_status_map = {row[0]: row[1] for row in cursor.fetchall()}
    
    # Load all product slugs and IDs from database, sorted by slug length descending
    cursor.execute("SELECT id, slug FROM product;")
    db_products = cursor.fetchall()
    db_products_sorted = sorted(db_products, key=lambda x: len(x[1]), reverse=True)
    
    total_reviews_ingested = 0
    
    for item in reviews_data_list:
        reviews_list = item.get("reviews", [])
        prod_slug = item.get("product_slug", "")
        if not reviews_list:
            continue
            
        # Find the matching product using prefix matching
        p_id = None
        for p_id_val, p_slug_val in db_products_sorted:
            if prod_slug == p_slug_val or prod_slug.startswith(p_slug_val + "-"):
                p_id = p_id_val
                break
                
        if not p_id:
            print(f"Skipping reviews: Product with slug '{prod_slug}' not found in database.")
            continue
        
        # Grab a variant for orders
        cursor.execute("SELECT id, sell_price FROM product_variant WHERE product_id = %s LIMIT 1;", (p_id,))
        v_row = cursor.fetchone()
        if not v_row:
            print(f"Skipping reviews: No product variant found for product ID {p_id}.")
            continue
        v_id, price = v_row
        
        print(f"Ingesting {len(reviews_list)} reviews for product '{prod_slug}'...")
        
        for rev in reviews_list:
            reviewer_name = (rev.get("reviewer_name") or "Khách hàng ẩn danh").strip().title()
            reviewer_name = reviewer_name[:255]
            rating = int(rev.get("rating") or 5)
            comment = (rev.get("comment") or "").strip()
            
            # 1. Ingest User
            email_part = slugify(reviewer_name)
            if not email_part:
                email_part = f"user-{random.randint(1000, 9999)}"
            email = (email_part + "@example.com")[:255]
            
            cursor.execute("SELECT id FROM users WHERE email = %s;", (email,))
            u_row = cursor.fetchone()
            if u_row:
                u_id = u_row[0]
            else:
                cursor.execute("""
                    INSERT INTO users (full_name, email, password, role, is_verified)
                    VALUES (%s, %s, 'customer_password_hash', 'customer', TRUE)
                    RETURNING id;
                """, (reviewer_name, email))
                u_id = cursor.fetchone()[0]
                
                # Ingest address for the user
                cursor.execute("""
                    INSERT INTO address (user_id, full_name, phone, district, province, ward, detail_address, is_default)
                    VALUES (%s, %s, '0988776655', 'Quận 1', 'Hồ Chí Minh', 'Phường Bến Nghé', 'Số 99 Đường Đồng Khởi', TRUE);
                """, (u_id, reviewer_name))
                
            # 2. Ingest Completed Order (to satisfy FK constraints on reviews table)
            # Select random store
            s_id = random.choice(store_ids) if store_ids else 1
            
            # Check if order already exists for this customer + variant
            cursor.execute("""
                SELECT o.id FROM orders o
                JOIN order_details od ON o.id = od.order_id
                WHERE o.user_id = %s AND od.variant_id = %s;
            """, (u_id, v_id))
            ord_row = cursor.fetchone()
            
            if ord_row:
                order_id = ord_row[0]
            else:
                order_code = f"ORD-{u_id}-{v_id}-{random.randint(10000, 99999)}"[:100]
                cursor.execute("""
                    INSERT INTO orders (
                        order_code, user_id, store_id, order_status_id, payment_status_id, shipping_status_id, 
                        total_amount, payment_method, receiver_name, receiver_address, receiver_phone
                    )
                    VALUES (%s, %s, %s, %s, %s, %s, %s, 'cod', %s, 'Số 99 Đường Đồng Khởi, Phường Bến Nghé, Quận 1, Hồ Chí Minh', '0988776655')
                    RETURNING id;
                """, (
                    order_code, u_id, s_id, 
                    order_status_map.get('delivered', 1), 
                    payment_status_map.get('paid', 1), 
                    shipping_status_map.get('delivered', 1),
                    price, reviewer_name
                ))
                order_id = cursor.fetchone()[0]
                
                # Ingest order detail
                cursor.execute("""
                    INSERT INTO order_details (order_id, variant_id, quantity, unit_price, total_cost)
                    VALUES (%s, %s, 1, %s, %s);
                """, (order_id, v_id, price, price))
                
            # 3. Ingest Review
            cursor.execute("SELECT id FROM reviews WHERE user_id = %s AND product_id = %s AND order_id = %s;", (u_id, p_id, order_id))
            if not cursor.fetchone():
                cursor.execute("""
                    INSERT INTO reviews (user_id, product_id, order_id, rating, comment, images, status)
                    VALUES (%s, %s, %s, %s, %s, '[]'::jsonb, 'approved');
                """, (u_id, p_id, order_id, rating, comment))
                total_reviews_ingested += 1
                
        print(f"Reviews processed successfully for '{prod_slug}'.")
    print(f"Total reviews ingested: {total_reviews_ingested}")

def seed_marketing_vouchers(cursor):
    """Seed sample discount vouchers into the vouchers table."""
    print("Seeding voucher coupon templates...")
    vouchers = [
        ("KM10", "Khuyến mãi 10% tổng hóa đơn", "percentage", 10),
        ("GIAM50K", "Giảm trực tiếp 50,000đ", "fixed_amount", 50000),
    ]
    for code, name, dtype, dval in vouchers:
        cursor.execute("SELECT id FROM vouchers WHERE code = %s;", (code,))
        if not cursor.fetchone():
            cursor.execute("""
                INSERT INTO vouchers (code, name, discount_type, discount_value, start_date, end_date, discount_target, min_order_value, is_deleted)
                VALUES (%s, %s, %s, %s, NOW() - INTERVAL '1 day', NOW() + INTERVAL '30 days', 'order', 100000, FALSE);
            """, (code, name, dtype, dval))

def ingest_banners(cursor):
    """Ingest homepage and category banners into the database."""
    print("Ingesting homepage and category-specific banners...")
    
    # Get category ID map
    cursor.execute("SELECT id, name FROM category;")
    cat_map = {name.lower().strip(): cid for cid, name in cursor.fetchall()}
    
    banners = [
        (
            'LIMIT TIME OFFER', 
            'Sản phẩm công nghệ giảm giá tới 50%', 
            'Cơ hội sở hữu laptop, điện thoại cao cấp chính hãng với giá rẻ nhất thị trường.',
            'https://images.unsplash.com/photo-1496181130204-755241544e35?auto=format&fit=crop&w=1200&q=80', 
            'Điện tử & Công nghệ', 
            '/browse', 
            1, 
            True,
            None
        ),
        (
            'NEW FASHION ERA', 
            'Bộ sưu tập thời trang mùa hè cực hot', 
            'Phong cách tối giản thời thượng. Hoàn tiền 100% nếu không hài lòng.',
            'https://images.unsplash.com/photo-1483985988355-763728e1935b?auto=format&fit=crop&w=1200&q=80', 
            'Fashion & LifeStyle', 
            '/browse', 
            2, 
            True,
            None
        ),
        (
            'THẾ HỆ AI PHONE MỚI',
            'Sở hữu iPhone 16 Pro & Galaxy S24 Ultra',
            'Ưu đãi trả góp 0%, thu cũ đổi mới trợ giá lên tới 2 triệu đồng.',
            'https://images.unsplash.com/photo-1511707171634-5f897ff02aa9?auto=format&fit=crop&w=1200&q=80',
            'Điện thoại',
            '/browse?category=' + str(cat_map.get('điện thoại', '')),
            1,
            True,
            cat_map.get('điện thoại')
        ),
        (
            'LAPTOP GAMING & VĂN PHÒNG',
            'Bứt phá hiệu năng, nhận ngàn quà tặng',
            'Giảm thêm 10% cho học sinh sinh viên. Tặng balo cao cấp & chuột không dây.',
            'https://images.unsplash.com/photo-1603302576837-37561b2e2302?auto=format&fit=crop&w=1200&q=80',
            'Laptop',
            '/browse?category=' + str(cat_map.get('laptop', '')),
            2,
            True,
            cat_map.get('laptop')
        ),
        (
            'THẾ GIỚI PHỤ KIỆN CHÍNH HÃNG',
            'Tai nghe & Loa Bluetooth giảm tới 40%',
            'Âm thanh cực đỉnh từ Sony, JBL, Marshall. Bảo hành chính hãng 12 tháng.',
            'https://images.unsplash.com/photo-1484704849700-f032a568e944?auto=format&fit=crop&w=1200&q=80',
            'Tai nghe',
            '/browse?category=' + str(cat_map.get('tai nghe', '')),
            3,
            True,
            cat_map.get('tai nghe')
        )
    ]
    
    for title, subtitle, description, image_url, tag, link_url, sort_order, is_active, category_id in banners:
        cursor.execute("SELECT id FROM banners WHERE title = %s;", (title,))
        if not cursor.fetchone():
            cursor.execute("""
                INSERT INTO banners (title, subtitle, description, image_url, tag, link_url, sort_order, is_active, category_id, created_at, updated_at)
                VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, NOW(), NOW());
            """, (title, subtitle, description, image_url, tag, link_url, sort_order, is_active, category_id))
    print("Successfully processed banners table.")

def seed_promotions(cursor):
    """Seed sample promotions for some products to drive the Flash Sale section."""
    print("Seeding flash sale promotions...")
    
    # Get some product IDs
    cursor.execute("SELECT id, name FROM product LIMIT 5;")
    products = cursor.fetchall()
    
    if not products:
        print("No products found in the database. Cannot seed promotions.")
        return
        
    print(f"Found {len(products)} products to seed promotions for.")
    
    # We will create active promotions that end 3-7 days in the future, and start 1 day in the past.
    for i, (p_id, p_name) in enumerate(products):
        if i % 2 == 0:
            discount_type = "percentage"
            discount_value = 10.0 + (i * 5) # 10%, 15%, 20%...
            name = f"Flash Sale {p_name[:20]} - Giảm {int(discount_value)}%"
        else:
            discount_type = "fixed"
            discount_value = 50000.0 * (i + 1) # 50k, 100k...
            name = f"Flash Sale {p_name[:20]} - Giảm {int(discount_value):,}đ"
            
        description = f"Chương trình Flash Sale đặc biệt cho sản phẩm {p_name}."
        
        # Check if a promotion already exists for this product
        cursor.execute("SELECT id FROM promotions WHERE product_id = %s AND is_deleted = false;", (p_id,))
        if not cursor.fetchone():
            cursor.execute("""
                INSERT INTO promotions (product_id, name, description, discount_type, discount_value, start_date, end_date, is_active, is_deleted)
                VALUES (%s, %s, %s, %s, %s, NOW() - INTERVAL '1 day', NOW() + INTERVAL '5 days', TRUE, FALSE);
            """, (p_id, name, description, discount_type, discount_value))
            
    print("Successfully processed promotions table.")

def main():
    print("====================================================")
    print("           TGDD CORE BUSINESS ETL PIPELINE          ")
    print("====================================================")
    
    # 1. Read JSON Datasets
    stores_file = DATA_DIR / "stores.json"
    reviews_file = DATA_DIR / "reviews.json"
    products_file = DATA_DIR / "products.json"
        
    print(f"Datasets Path:\n- Stores: {stores_file}\n- Products: {products_file}\n- Reviews: {reviews_file}")
    
    stores_data = load_json_file(stores_file)
    products_data = load_json_file(products_file)
    reviews_data = load_json_file(reviews_file)
    
    # 2. Connect to Database and run ETL steps
    conn = get_db_connection()
    try:
        with conn.cursor() as cursor:
            # Step 0: Apply dynamic schema patch (ensure product_spec.value is TEXT)
            print("Applying database schema patch: altering product_spec.value type to TEXT...")
            cursor.execute("ALTER TABLE product_spec ALTER COLUMN value TYPE TEXT;")
            
            # Step 1: Reference Statuses
            seed_statuses(cursor)
            
            # Step 2: Ingest Stores
            store_ids = ingest_stores(cursor, stores_data)
            
            # Step 3: Ingest Products and related specifications, options & variants
            ingest_products_catalog(cursor, products_data, store_ids)
            
            # Step 4: Seed Vouchers
            seed_marketing_vouchers(cursor)
            
            # Step 4b: Seed Banners
            ingest_banners(cursor)
            
            # Step 4c: Seed Promotions (Flash Sale)
            seed_promotions(cursor)
            
            # Step 5: Ingest Users, Mock Orders and Reviews
            ingest_reviews_and_users(cursor, reviews_data, store_ids)
            
            # Commit the transactions
            conn.commit()
            print("\nETL Transaction COMMITTED successfully.")
            
    except Exception as e:
        import traceback
        traceback.print_exc()
        conn.rollback()
        print(f"\nETL Ingestion FAILED. Transaction ROLLED BACK. Detail: {e}")
        sys.exit(1)
    finally:
        conn.close()
        print("Database connection closed.")
        print("====================================================")

if __name__ == "__main__":
    main() 
