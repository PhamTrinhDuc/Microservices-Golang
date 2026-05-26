import re
import sys
import html
import json
import time
import os
import argparse
from selenium import webdriver
from selenium.webdriver.common.by import By
from selenium.webdriver.chrome.options import Options
from selenium.webdriver.chrome.service import Service
from webdriver_manager.chrome import ChromeDriverManager

sys.stdout.reconfigure(encoding='utf-8')

def init_driver():
    options = Options()
    options.add_argument("--headless")
    options.add_argument("--disable-blink-features=AutomationControlled")
    options.add_argument("--start-maximized")
    options.add_argument("user-agent=Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
    
    driver = webdriver.Chrome(
        service=Service(ChromeDriverManager().install()),
        options=options
    )
    return driver

def clean_product_name(raw_name):
    # Remove category prefixes at the beginning (including accessories)
    prefixes = [
        r'Điện thoại', r'Laptop', r'Máy tính bảng', r'Đồng hồ thông minh', r'Smartwatch', r'Đồng hồ', 
        r'Máy ảnh', r'Máy in', r'Tai nghe Bluetooth True Wireless', r'Tai nghe Bluetooth', r'Tai nghe có dây', 
        r'Tai nghe chụp tai', r'Tai nghe', r'Pin sạc dự phòng', r'Sạc dự phòng', r'Cáp sạc', r'Adapter sạc', 
        r'Sạc xe hơi', r'Chuột không dây', r'Chuột có dây', r'Chuột gaming', r'Chuột', r'Bàn phím không dây', 
        r'Bàn phím có dây', r'Bàn phím gaming', r'Bàn phím', r'Thẻ nhớ', r'Loa Bluetooth', r'Loa vi tính', 
        r'Loa', r'Ốp lưng điện thoại', r'Ốp lưng máy tính bảng', r'Ốp lưng', r'Vỏ bảo vệ'
    ]
    name = raw_name
    for p in prefixes:
        name = re.sub(rf'^{p}\s+', '', name, flags=re.IGNORECASE)
        
    # Strip memory configs (e.g. 12GB/256GB, 256GB, 12GB, 128GB)
    name = re.sub(r'\b\d+\s*(?:GB|TB)(?:\s*/\s*\d+\s*(?:GB|TB))?\b', '', name, flags=re.IGNORECASE)
    # Strip empty parentheses and clean spacing
    name = re.sub(r'\(\s*\)', '', name)
    name = name.strip()
    name = re.sub(r'\s+-\s*$', '', name)
    return name

def slugify(text):
    text = text.lower().strip()
    # Replace Vietnamese tones with unsigned letters
    vietnamese_map = {
        'a': 'áàảãạăắằẳẵặâấầẩẫậ',
        'e': 'éèẻẽẹêếềểễệ',
        'i': 'íìỉĩị',
        'o': 'óòỏõọôốồổỗộơớờởỡợ',
        'u': 'úùủũụưứừửữự',
        'y': 'ýỳỷỹỵ',
        'd': 'đ'
    }
    for k, v in vietnamese_map.items():
        for char in v:
            text = text.replace(char, k)
            
    text = re.sub(r'[^\w\s-]', '', text)
    text = re.sub(r'[\s_]+', '-', text)
    return text

def parse_weight(specs):
    weight_keys = ["Kích thước, khối lượng", "Trọng lượng", "Khối lượng", "Kích thước - Trọng lượng", "Nặng"]
    for key in weight_keys:
        val = specs.get(key)
        if val:
            g_match = re.search(r'(\d+(?:\.\d+)?)\s*(?:g|gram|g\b)', val, re.IGNORECASE)
            if g_match:
                return float(g_match.group(1))
            kg_match = re.search(r'(\d+(?:\.\d+)?)\s*kg', val, re.IGNORECASE)
            if kg_match:
                return float(kg_match.group(1)) * 1000
    return None

def split_value_unit(value_str):
    value_str = value_str.strip()
    units = ["GB", "MB", "mAh", "MP", "GHz", "W", "VND", "inch", "g", "gram", "mm", "px", "Pixels", "Hz", "nits", "fps"]
    for u in units:
        pat = rf'^(\d+(?:\.\d+)?)\s*{u}$'
        m = re.match(pat, value_str, re.IGNORECASE)
        if m:
            return m.group(1), u
    return value_str, ""

def parse_html_grouped_specs(html_content):
    grouped_specs = {}
    matches = list(re.finditer(r'<div class="box-specifi">', html_content))
    for idx, m in enumerate(matches):
        start_pos = m.start()
        end_pos = html_content.find("</ul>", start_pos)
        if end_pos != -1:
            end_pos += len("</ul>")
            chunk = html_content[start_pos:end_pos]
            
            title_match = re.search(r'<h3>(.*?)</h3>', chunk)
            if title_match:
                group_title = re.sub(r'<[^>]+>', '', title_match.group(1)).strip()
                group_title = html.unescape(group_title)
            else:
                group_title = "Thông số kỹ thuật"
                
            if group_title not in grouped_specs:
                grouped_specs[group_title] = {}
            
            li_matches = re.findall(r'<li>(.*?)</li>', chunk, re.DOTALL)
            for li in li_matches:
                asides = re.findall(r'<aside>(.*?)</aside>', li, re.DOTALL)
                if len(asides) >= 2:
                    key = re.sub(r'<[^>]+>', '', asides[0]).strip()
                    key = html.unescape(key).rstrip(":")
                    val = re.sub(r'<[^>]+>', '', asides[1]).strip()
                    val = html.unescape(val)
                    grouped_specs[group_title][key] = val
                    
    return grouped_specs

def parse_tgdd_product(url, driver):
    driver.get(url)
    time.sleep(4)
    page_source = driver.page_source
    
    # 1. Page Metadata
    meta_title = ""
    meta_description = ""
    try:
        meta_title = driver.title.strip()
        meta_desc_elem = driver.find_element(By.CSS_SELECTOR, "meta[name='description']")
        meta_description = meta_desc_elem.get_attribute("content").strip()
    except Exception as e:
        pass
        
    # 2. JSON-LD structured data
    json_ld_match = re.search(r'<script type="application/ld\+json" id="productld">(.*?)</script>', page_source, re.DOTALL)
    product_data = {}
    flat_specs = {}
    
    if json_ld_match:
        try:
            structured_data = json.loads(json_ld_match.group(1).strip())
            product_data["name"] = structured_data.get("name")
            product_data["sku"] = structured_data.get("sku")
            product_data["brand"] = structured_data.get("brand", {}).get("name", ["N/A"])[0]
            product_data["description"] = structured_data.get("description")
            product_data["thumbnail"] = structured_data.get("image", {}).get("contentUrl")
            
            offers = structured_data.get("offers", {})
            product_data["price"] = offers.get("price")
            product_data["currency"] = offers.get("priceCurrency")
            product_data["availability"] = offers.get("availability")
            
            for prop in structured_data.get("additionalProperty", []):
                name = prop.get("name")
                value = prop.get("value")
                if name and value:
                    cleaned_val = re.sub(r'<[^>]+>', '', value).strip()
                    flat_specs[name] = html.unescape(cleaned_val)
        except Exception as json_err:
            pass
            
    if not product_data:
        try:
            product_data["name"] = driver.find_element(By.CSS_SELECTOR, ".product-name h1").text.strip()
        except:
            pass
            
    # 3. HTML Grouped Specifications
    grouped_specs = parse_html_grouped_specs(page_source)
    
    # Merge flat_specs if grouped_specs is empty
    if not grouped_specs and flat_specs:
        grouped_specs["Thông số kỹ thuật"] = flat_specs
        
    # Compile flat specifications for weight lookup
    all_specs_flat = {}
    for group, items in grouped_specs.items():
        for k, v in items.items():
            all_specs_flat[k] = v
            
    # Parse Weight
    weight = parse_weight(all_specs_flat)
    
    # Clean product name and create slug
    raw_name = product_data.get("name", "")
    clean_name = clean_product_name(raw_name)
    product_slug = slugify(clean_name)
    
    # 4. Color & Capacity Options
    colors = []
    capacities = []
    
    # Colors: class="box03 color group desk"
    try:
        color_elements = driver.find_elements(By.CSS_SELECTOR, "div.box03.color.group.desk a")
        for elem in color_elements:
            href = elem.get_attribute("href")
            color_id = elem.get_attribute("data-color")
            code = elem.get_attribute("data-code")
            label = elem.text.strip()
            is_active = "act" in elem.get_attribute("class") or "active" in elem.get_attribute("class")
            
            bg_color = None
            try:
                style = elem.find_element(By.TAG_NAME, "i").get_attribute("style")
                if style:
                    bg_match = re.search(r'background-color:\s*([^";]+)', style)
                    if bg_match:
                        bg_color = bg_match.group(1).strip()
            except:
                pass
                    
            colors.append({
                "label": label,
                "href": href,
                "color_id": color_id,
                "code": code,
                "is_active": is_active,
                "bg_color": bg_color
            })
    except Exception as e:
        pass
        
    # Capacities: class="box03 group desk"
    try:
        boxes = driver.find_elements(By.CSS_SELECTOR, "div.box03.group.desk")
        for box in boxes:
            classes = box.get_attribute("class")
            if "color" not in classes:
                capacity_elements = box.find_elements(By.TAG_NAME, "a")
                for elem in capacity_elements:
                    href = elem.get_attribute("href")
                    label = elem.text.strip()
                    is_active = "act" in elem.get_attribute("class") or "active" in elem.get_attribute("class")
                    
                    capacities.append({
                        "label": label,
                        "href": href,
                        "is_active": is_active
                    })
    except Exception as e:
        pass
        
    # Default Variant Fallback
    if not colors:
        color_label = "Mặc định"
        thumb_url = product_data.get("thumbnail", "")
        if thumb_url:
            filename = thumb_url.split("/")[-1].lower()
            for c in ["trang", "den", "vang", "cam", "tim", "xanh", "hong", "titan", "xam", "do", "bac"]:
                if c in filename:
                    trans = {
                        "trang": "Trắng", "den": "Đen", "vang": "Vàng", "cam": "Cam", 
                        "tim": "Tím", "xanh": "Xanh", "hong": "Hồng", "titan": "Titan",
                        "xam": "Xám", "do": "Đỏ", "bac": "Bạc"
                    }
                    color_label = trans[c]
                    break
        colors.append({
            "label": color_label,
            "href": url,
            "color_id": "0",
            "code": product_data.get("sku", "0"),
            "is_active": True,
            "bg_color": None
        })
        
    if not capacities:
        capacity_label = "Mặc định"
        storage_spec = all_specs_flat.get("Dung lượng lưu trữ") or all_specs_flat.get("Bộ nhớ trong") or all_specs_flat.get("Dung lượng pin")
        if storage_spec:
            capacity_label = storage_spec.strip()
        else:
            cap_match = re.search(r'\b\d+\s*(?:GB|TB)\b', raw_name, re.IGNORECASE)
            if cap_match:
                capacity_label = cap_match.group(0).strip()
                
        capacities.append({
            "label": capacity_label,
            "href": url,
            "is_active": True
        })
        
    # Parse specifications value and units
    specs_with_units = {}
    for group, items in grouped_specs.items():
        specs_with_units[group] = {}
        for k, v in items.items():
            val, unit = split_value_unit(v)
            specs_with_units[group][k] = {
                "raw_value": v,
                "value": val,
                "unit": unit
            }

    # 5. Extract Detailed Gallery Images (ProductImage Table)
    gallery_images = []
    main_thumb = product_data.get("thumbnail")
    if main_thumb:
        if main_thumb.startswith("//"):
            main_thumb = "https:" + main_thumb
        gallery_images.append({
            "url": main_thumb,
            "alt_text": f"Ảnh đại diện {clean_name}",
            "sort_order": 1,
            "is_thumbnail": True
        })
        
    try:
        # Find gallery elements
        slider_imgs = driver.find_elements(By.CSS_SELECTOR, "div.slider-img img.owl-lazy, div#slider-detail img.owl-lazy, .media-slider img.owl-lazy")
        if not slider_imgs:
            slider_imgs = driver.find_elements(By.CSS_SELECTOR, "div.slider-img img, div#slider-detail img, .media-slider img")
            
        seen_urls = {main_thumb} if main_thumb else set()
        order = len(gallery_images) + 1
        for img in slider_imgs:
            src = img.get_attribute("data-src") or img.get_attribute("src")
            if src:
                if src.startswith("//"):
                    src = "https:" + src
                if src not in seen_urls and not src.endswith(".gif") and "Banner" not in src:
                    seen_urls.add(src)
                    gallery_images.append({
                        "url": src,
                        "alt_text": f"Ảnh chi tiết {clean_name}",
                        "sort_order": order,
                        "is_thumbnail": False
                    })
                    order += 1
    except Exception as img_err:
        pass
        
    # 6. Extract Promotions (Promotions Table)
    promotions = []
    try:
        # Check standard TGDD promo boxes
        promo_items = driver.find_elements(By.CSS_SELECTOR, "div.block__promo .pr-item, div.block__promo p, div.policy-promo p, div.box-promo p")
        seen_promos = set()
        for item in promo_items:
            text = item.text.strip()
            if text and len(text) > 8 and text not in seen_promos:
                seen_promos.add(text)
                
                # Check for percentage discount in promo text
                discount_percent = 0
                pct_match = re.search(r'giảm\s*(\d+)%', text, re.IGNORECASE)
                if pct_match:
                    discount_percent = int(pct_match.group(1))
                    
                promotions.append({
                    "name": "Khuyến mãi Thế Giới Di Động",
                    "description": text,
                    "discount_percent": discount_percent,
                    "is_active": True
                })
    except Exception as promo_err:
        pass
            
    result = {
        "source_url": url,
        "product_info": {
            "id": product_data.get("sku"),
            "brand_name": product_data.get("brand"),
            "raw_name": raw_name,
            "clean_name": clean_name,
            "slug": product_slug,
            "meta_title": meta_title,
            "meta_description": meta_description,
            "img_thumb": main_thumb,
            "price": product_data.get("price"),
            "weight": weight,
            "is_active": product_data.get("availability") == "https://schema.org/InStock"
        },
        "colors": colors,
        "capacities": capacities,
        "specifications": specs_with_units,
        "gallery_images": gallery_images,
        "promotions": promotions
    }
    
    return result

def crawl_category(category_url, category_name, limit=5):
    print(f"\n--- Starting Bulk Product Crawl: {category_name} (Limit: {limit}) ---")
    driver = init_driver()
    
    product_links = []
    all_products = []
    
    try:
        print(f"Loading category page: {category_url}...")
        driver.get(category_url)
        time.sleep(4)
        
        items = driver.find_elements(By.CSS_SELECTOR, "ul.listproduct li.item > a")
        print(f"Found {len(items)} product elements on category page.")
        
        for item in items:
            href = item.get_attribute("href")
            if href and href not in product_links:
                product_links.append(href)
                if len(product_links) >= limit:
                    break
                
        print(f"URLs to crawl: {product_links}")
        
        for idx, url in enumerate(product_links):
            print(f"\n[{idx+1}/{len(product_links)}] Crawling: {url}")
            try:
                product_data = parse_tgdd_product(url, driver)
                product_data["category_name"] = category_name
                
                print(f"  -> Extracted: {product_data['product_info']['clean_name']}")
                print(f"  -> Gallery Images count: {len(product_data['gallery_images'])}")
                print(f"  -> Promotions count: {len(product_data['promotions'])}")
                
                all_products.append(product_data)
                time.sleep(2)
            except Exception as e:
                print(f"Error crawling product {url}: {e}")
                
    finally:
        driver.quit()
        
    output_dir = "./data"
    os.makedirs(output_dir, exist_ok=True)
    output_file = os.path.join(output_dir, "products.json")
    
    with open(output_file, "w", encoding="utf-8") as f:
        json.dump(all_products, f, indent=4, ensure_ascii=False)
        
    print(f"\nSuccessfully crawled {len(all_products)} products.")
    print(f"Saved dataset to: {os.path.abspath(output_file)}")

def main():
    parser = argparse.ArgumentParser(description="TGDD Product, Gallery, and Promo Crawler")
    parser.add_argument("--url", help="Crawl a single product page directly")
    parser.add_argument("--category", help="Category listing page URL to crawl in bulk")
    parser.add_argument("--name", default="Sản phẩm", help="Name of the category being crawled in bulk")
    parser.add_argument("--limit", type=int, default=3, help="Max number of products to crawl in bulk")
    
    args = parser.parse_args()
    
    if args.url:
        driver = init_driver()
        try:
            res = parse_tgdd_product(args.url, driver)
            
            output_dir = "./data"
            os.makedirs(output_dir, exist_ok=True)
            output_file = os.path.join(output_dir, "test_product_schema_aligned.json")
            
            with open(output_file, "w", encoding="utf-8") as f:
                json.dump(res, f, indent=4, ensure_ascii=False)
                
            print(f"\nSuccessfully crawled 1 product: {res['product_info']['clean_name']}")
            print(f"Saved data to: {os.path.abspath(output_file)}")
            print(f"Gallery Images count: {len(res['gallery_images'])}")
            print(f"Promotions count: {len(res['promotions'])}")
        finally:
            driver.quit()
    elif args.category:
        crawl_category(args.category, args.name, args.limit)
    else:
        # Default run demo on 1 phone & 1 headphone
        print("No arguments provided. Running default crawl demo...")
        demo_links = [
            "https://www.thegioididong.com/dtdd/iphone-17",
            "https://www.thegioididong.com/tai-nghe/tai-nghe-bluetooth-true-wireless-ava-go-p310"
        ]
        
        driver = init_driver()
        demo_results = []
        try:
            for idx, url in enumerate(demo_links):
                print(f"\n[{idx+1}/{len(demo_links)}] Crawling: {url}")
                try:
                    res = parse_tgdd_product(url, driver)
                    res["category_name"] = "Điện thoại" if "dtdd" in url else "Tai nghe"
                    demo_results.append(res)
                    time.sleep(2)
                except Exception as e:
                    print(f"Error: {e}")
        finally:
            driver.quit()
            
        output_dir = "./data"
        os.makedirs(output_dir, exist_ok=True)
        output_file = os.path.join(output_dir, "products.json")
        with open(output_file, "w", encoding="utf-8") as f:
            json.dump(demo_results, f, indent=4, ensure_ascii=False)
        print(f"\nDemo finished. Saved results to: {os.path.abspath(output_file)}")

if __name__ == "__main__":
    main()
