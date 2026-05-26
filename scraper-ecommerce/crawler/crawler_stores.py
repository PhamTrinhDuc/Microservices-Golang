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

def clean_province_name(name):
    # Remove "Tỉnh ", "Thành phố ", "TP. " prefix
    name = re.sub(r'^(Tỉnh|Thành phố|TP\.)\s+', '', name, flags=re.IGNORECASE)
    return name.strip()

def parse_address_details(raw_address, default_province=""):
    # E.g. "Số 175, quốc lộ 63, khu phố 3, Xã An Biên, Tỉnh An Giang, Việt Nam"
    address = raw_address.strip()
    
    # Strip trailing "Việt Nam"
    address = re.sub(r',\s*Việt Nam$', '', address, flags=re.IGNORECASE).strip()
    
    parts = [p.strip() for p in address.split(',')]
    province = default_province
    ward = ""
    road = ""
    
    if len(parts) >= 1:
        # Check if last part is province
        last_part = parts[-1]
        if "tỉnh" in last_part.lower() or "thành phố" in last_part.lower() or "tp" in last_part.lower() or last_part in [
            "Hồ Chí Minh", "Hà Nội", "Đà Nẵng", "Cần Thơ", "Hai Phòng"
        ]:
            province = clean_province_name(last_part)
            parts.pop()
            
    if len(parts) >= 1:
        # Check if second to last part is ward/district
        last_part = parts[-1]
        if any(keyword in last_part.lower() for keyword in ["xã", "phường", "thị trấn", "quận", "huyện", "tp", "thành phố", "tt", "h", "p"]):
            ward = last_part
            parts.pop()
            
    # The remaining parts are the road / detail address
    road = ", ".join(parts).strip()
    
    # Fallbacks
    if not province:
        province = default_province
    if not ward and road:
        # If ward is empty, try to see if road has ward info
        pass
        
    return {
        "province": clean_province_name(province),
        "ward": ward,
        "road": road
    }

def scrape_provinces(driver):
    url = "https://www.thegioididong.com/he-thong-sieu-thi-the-gioi-di-dong"
    print(f"Fetching provinces index from: {url}...")
    driver.get(url)
    time.sleep(4)
    
    provinces = []
    try:
        # Find province-box__store block
        container = driver.find_element(By.ID, "province-box__store")
        links = container.find_elements(By.TAG_NAME, "a")
        for link in links:
            href = link.get_attribute("href")
            data_value = link.get_attribute("data-value")
            name = link.find_element(By.CLASS_NAME, "address-cr").get_attribute("textContent").strip()
            
            if href and name:
                provinces.append({
                    "name": name,
                    "data_value": data_value,
                    "url": href
                })
    except Exception as e:
        print(f"Error fetching provinces list: {e}")
        
    print(f"Found {len(provinces)} provinces.")
    return provinces

def scrape_stores_in_province(driver, province_url, province_name, limit_stores=None):
    print(f"Scraping stores in province: {province_name} ({province_url})...")
    driver.get(province_url)
    time.sleep(3)
    
    stores = []
    try:
        # Stores are list items in ul or containers
        # From HTML analysis, it is in li element with data-id attribute
        li_elements = driver.find_elements(By.CSS_SELECTOR, "li[data-id]")
        print(f"  Found {len(li_elements)} store elements.")
        
        count = 0
        for li in li_elements:
            if limit_stores and count >= limit_stores:
                break
                
            try:
                # First anchor tag contains store brand + full address text
                store_a = li.find_element(By.CSS_SELECTOR, "a:not(.gd)")
                raw_name_addr = store_a.get_attribute("textContent").strip()
                
                # Google maps query link
                map_a = li.find_element(By.CSS_SELECTOR, "a.gd")
                map_href = map_a.get_attribute("href")
                
                # Extract lat/lng
                lat, lng = None, None
                if map_href:
                    coords_match = re.search(r'query=([-+]?\d*\.\d+|\d+),([-+]?\d*\.\d+|\d+)', map_href)
                    if coords_match:
                        lat = float(coords_match.group(1))
                        lng = float(coords_match.group(2))
                
                # Parse name and address details
                # raw_name_addr looks like: "Điện máy Xanh Số 175 Quốc Lộ 63 (Thứ Ba), Số 175, quốc lộ 63, khu phố 3, Xã An Biên, Tỉnh An Giang, Việt Nam"
                # Split by brand name if we want to separate brand name and address details
                brand = "Thế Giới Di Động"
                address_part = raw_name_addr
                
                for b_prefix in ["Điện máy Xanh", "Thế giới di động", "TopZone", "An Khang"]:
                    if raw_name_addr.lower().startswith(b_prefix.lower()):
                        brand = b_prefix
                        # Remove prefix and space/comma
                        address_part = raw_name_addr[len(b_prefix):].strip(" ,")
                        break
                
                addr_parsed = parse_address_details(address_part, province_name)
                
                # Construct name
                store_name = f"{brand} {addr_parsed['road'] or addr_parsed['ward']}".strip()
                
                stores.append({
                    "name": store_name,
                    "hotline": "18001060", # default hotline
                    "province": addr_parsed["province"],
                    "ward": addr_parsed["ward"],
                    "road": addr_parsed["road"],
                    "lat": lat,
                    "lng": lng,
                    "is_active": True
                })
                count += 1
            except Exception as item_err:
                pass
    except Exception as e:
        print(f"  Error scraping stores: {e}")
        
    print(f"  Successfully extracted {len(stores)} stores.")
    return stores

def main():
    parser = argparse.ArgumentParser(description="TGDD Retail Stores Crawler")
    parser.add_argument("--limit-provinces", type=int, help="Limit number of provinces to crawl (for quick testing)")
    parser.add_argument("--limit-stores", type=int, help="Limit number of stores per province")
    parser.add_argument("--output", default="data/stores.json", help="Path to save output JSON file")
    
    args = parser.parse_args()
    
    driver = init_driver()
    all_stores = []
    
    try:
        provinces = scrape_provinces(driver)
        
        target_provinces = provinces
        if args.limit_provinces:
            target_provinces = provinces[:args.limit_provinces]
            print(f"Limiting to first {len(target_provinces)} provinces.")
            
        for idx, prov in enumerate(target_provinces):
            print(f"\n[{idx+1}/{len(target_provinces)}] Scraping {prov['name']}...")
            prov_stores = scrape_stores_in_province(driver, prov["url"], prov["name"], args.limit_stores)
            all_stores.extend(prov_stores)
            time.sleep(2)
            
    finally:
        driver.quit()
        
    # Save results
    output_dir = os.path.dirname(args.output)
    if output_dir:
        os.makedirs(output_dir, exist_ok=True)
        
    with open(args.output, "w", encoding="utf-8") as f:
        json.dump(all_stores, f, indent=4, ensure_ascii=False)
        
    print(f"\nFinished crawling. Total stores collected: {len(all_stores)}")
    print(f"Saved store dataset to: {os.path.abspath(args.output)}")

if __name__ == "__main__":
    main()
