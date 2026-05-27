import time
import os
import json
import argparse
import sys
from urllib.parse import urlparse, urljoin
from pathlib import Path
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

def scrape_product_links(category_url, max_clicks=5):
    print(f"Initializing driver to scrape links from: {category_url}...")
    driver = init_driver()
    product_links = []
    
    try:
        driver.get(category_url)
        time.sleep(4)
        
        # Click "Xem thêm" button to load more products dynamically
        for i in range(max_clicks):
            print(f"Clicking 'Xem thêm' (Click {i+1}/{max_clicks})...")
            try:
                # Scroll down to make button visible and trigger lazy loading
                driver.execute_script("window.scrollTo(0, document.body.scrollHeight - 800);")
                time.sleep(1)
                
                # Check for various TGDD "Xem thêm" button selectors
                view_more_selectors = [
                    "div.view-more a",
                    "a.view-more",
                    ".view-more",
                    "a.read-more",
                    "div.read-more a"
                ]
                
                view_more_btn = None
                for selector in view_more_selectors:
                    try:
                        btn = driver.find_element(By.CSS_SELECTOR, selector)
                        if btn.is_displayed():
                            view_more_btn = btn
                            break
                    except Exception:
                        continue
                        
                if view_more_btn:
                    driver.execute_script("arguments[0].click();", view_more_btn)
                    time.sleep(3)
                else:
                    # Scroll all the way to bottom to trigger lazy load if button is absent
                    driver.execute_script("window.scrollTo(0, document.body.scrollHeight);")
                    time.sleep(2)
                    # Check if loaded more or if button disappeared
                    print("No visible 'Xem thêm' button found, checking if page is fully loaded.")
            except Exception as click_err:
                print(f"  Could not click: {click_err}")
                break
                
        # Scroll to load remaining images/elements
        driver.execute_script("window.scrollTo(0, document.body.scrollHeight);")
        time.sleep(2)
        
        # Find all link anchors
        anchors = driver.find_elements(By.CSS_SELECTOR, "ul.listproduct li.item > a, ul.list-product li a, .listproduct li a")
        print(f"Found {len(anchors)} potential product anchors.")
        
        seen_urls = set()
        for anchor in anchors:
            href = anchor.get_attribute("href")
            if not href:
                continue
                
            # Make absolute and clean parameters/query
            url = urljoin("https://www.thegioididong.com", href)
            parsed = urlparse(url)
            clean_url = f"{parsed.scheme}://{parsed.netloc}{parsed.path}"
            
            # Check path parts. TGDD detail pages match format: /category/slug (2 parts)
            # e.g., /dtdd/iphone-17
            path_parts = [p for p in parsed.path.split("/") if p]
            if len(path_parts) == 2:
                # Exclude static info or auxiliary pages
                exclude_keywords = ["tin-tuc", "chinh-sach", "huong-dan", "game-app", "hoi-dap", "sim-so"]
                if not any(k in clean_url for k in exclude_keywords):
                    if clean_url not in seen_urls:
                        seen_urls.add(clean_url)
                        product_links.append(clean_url)
                        
    except Exception as e:
        print(f"Error gathering links: {e}")
DEFAULT_CATEGORIES = {
    "dtdd": "https://www.thegioididong.com/dtdd",
    "laptop": "https://www.thegioididong.com/laptop",
    "tai-nghe": "https://www.thegioididong.com/tai-nghe",
    "tai-nghe-chup-tai": "https://www.thegioididong.com/tai-nghe-chup-tai",
    "ban-phim": "https://www.thegioididong.com/ban-phim",
    "may-tinh-bang": "https://www.thegioididong.com/may-tinh-bang",
    "gia-treo-man-hinh": "https://www.thegioididong.com/gia-treo-man-hinh",
    "may-tinh-de-ban": "https://www.thegioididong.com/may-tinh-de-ban",
    "man-hinh-may-tinh": "https://www.thegioididong.com/man-hinh-may-tinh",
    "may-choi-game-cam-tay": "https://www.thegioididong.com/may-choi-game-cam-tay",
    "dong-ho-thong-minh": "https://www.thegioididong.com/dong-ho-thong-minh",
    "dong-ho-deo-tay": "https://www.thegioididong.com/dong-ho-deo-tay",
    "pin-sac-du-phong": "https://www.thegioididong.com/pin-sac-du-phong",
    "sac-dt-dd": "https://www.thegioididong.com/sac-dt-dd",
    "op-lung-flipcover": "https://www.thegioididong.com/op-lung-flipcover",
    "hub-chuyen-doi": "https://www.thegioididong.com/hub-chuyen-doi",
    "chuot-may-tinh": "https://www.thegioididong.com/chuot-may-tinh",
    "loa-laptop": "https://www.thegioididong.com/loa-laptop",
    "thiet-bi-mang": "https://www.thegioididong.com/thiet-bi-mang",
    "micro-cac-loai": "https://www.thegioididong.com/micro-cac-loai",
}

def scrape_product_links(category_urls, max_clicks=5):
    """Scrape product links from a list of category URLs."""
    driver = init_driver()
    all_links = []
    seen_urls = set()
    
    try:
        for idx, category_url in enumerate(category_urls):
            print(f"\n[{idx+1}/{len(category_urls)}] Processing category listing: {category_url}")
            driver.get(category_url)
            time.sleep(4)
            
            # Click "Xem thêm" button to load more products dynamically
            for i in range(max_clicks):
                print(f"  Clicking 'Xem thêm' (Click {i+1}/{max_clicks})...")
                try:
                    # Scroll down to make button visible and trigger lazy loading
                    driver.execute_script("window.scrollTo(0, document.body.scrollHeight - 800);")
                    time.sleep(1)
                    
                    # Check for various TGDD "Xem thêm" button selectors
                    view_more_selectors = [
                        "div.view-more a",
                        "a.view-more",
                        ".view-more",
                        "a.read-more",
                        "div.read-more a"
                    ]
                    
                    view_more_btn = None
                    for selector in view_more_selectors:
                        try:
                            btn = driver.find_element(By.CSS_SELECTOR, selector)
                            if btn.is_displayed():
                                view_more_btn = btn
                                break
                        except Exception:
                            continue
                            
                    if view_more_btn:
                        driver.execute_script("arguments[0].click();", view_more_btn)
                        time.sleep(3)
                    else:
                        # Scroll all the way to bottom to trigger lazy load if button is absent
                        driver.execute_script("window.scrollTo(0, document.body.scrollHeight);")
                        time.sleep(2)
                        print("  No visible 'Xem thêm' button found, checking if page is fully loaded.")
                except Exception as click_err:
                    print(f"  Could not click: {click_err}")
                    break
                    
            # Scroll to load remaining images/elements
            driver.execute_script("window.scrollTo(0, document.body.scrollHeight);")
            time.sleep(2)
            
            # Find all link anchors
            anchors = driver.find_elements(By.CSS_SELECTOR, "ul.listproduct li.item > a, ul.list-product li a, .listproduct li a")
            print(f"  Found {len(anchors)} potential product anchors.")
            
            cat_links_count = 0
            for anchor in anchors:
                href = anchor.get_attribute("href")
                if not href:
                    continue
                    
                # Make absolute and clean parameters/query
                url = urljoin("https://www.thegioididong.com", href)
                parsed = urlparse(url)
                clean_url = f"{parsed.scheme}://{parsed.netloc}{parsed.path}"
                
                # Check path parts. TGDD detail pages match format: /category/slug (2 parts)
                path_parts = [p for p in parsed.path.split("/") if p]
                if len(path_parts) == 2:
                    # Exclude static info or auxiliary pages
                    exclude_keywords = ["tin-tuc", "chinh-sach", "huong-dan", "game-app", "hoi-dap", "sim-so"]
                    if not any(k in clean_url for k in exclude_keywords):
                        if clean_url not in seen_urls:
                            seen_urls.add(clean_url)
                            all_links.append(clean_url)
                            cat_links_count += 1
            print(f"  Extracted {cat_links_count} unique product URLs from this category.")
                            
    except Exception as e:
        print(f"Error gathering links: {e}")
    finally:
        driver.quit()
        
    print(f"\nTotal scraped unique product URLs across all categories: {len(all_links)}")
    return all_links

def main():
    parser = argparse.ArgumentParser(description="TGDD Product URL List Scraper")
    parser.add_argument("--categories", default="dtdd", help="Comma-separated category keys (e.g. dtdd,laptop,tai-nghe) or direct listing URLs")
    parser.add_argument("--clicks", type=int, default=5, help="Number of times to click 'Xem thêm' per category")
    parser.add_argument("--all", action="store_true", help="Scrape all pre-defined categories in TGDD")
    parser.add_argument("--output", default="../data/product_links.json", help="Path to save links output JSON file")
    
    args = parser.parse_args()
    
    # Process target categories
    urls_to_crawl = []
    if args.all:
        urls_to_crawl = list(DEFAULT_CATEGORIES.values())
        print(f"Configured to crawl ALL {len(urls_to_crawl)} default categories.")
    else:
        # Split input
        parts = [p.strip() for p in args.categories.split(",") if p.strip()]
        for part in parts:
            if part in DEFAULT_CATEGORIES:
                urls_to_crawl.append(DEFAULT_CATEGORIES[part])
            elif part.startswith("http://") or part.startswith("https://"):
                urls_to_crawl.append(part)
            else:
                print(f"Warning: Unknown category key '{part}'. Skipping.")
                
    if not urls_to_crawl:
        print("Error: No valid categories/URLs to crawl.")
        sys.exit(1)
        
    links = scrape_product_links(urls_to_crawl, args.clicks)
    
    # Save output
    output_path = Path(args.output) if os.path.isabs(args.output) else Path(__file__).resolve().parent / args.output
    output_path.parent.mkdir(parents=True, exist_ok=True)
    
    with open(output_path, "w", encoding="utf-8") as f:
        json.dump(links, f, indent=4, ensure_ascii=False)
        
    print(f"Saved {len(links)} links to: {output_path.resolve()}")

if __name__ == "__main__":
    main()


# # Ví dụ 1: Cào link cho 3 nhóm chính: điện thoại, laptop và tai nghe (mỗi nhóm nhấn "Xem thêm" 2 lần)
# python crawler/crawler_links.py --categories "dtdd,laptop,tai-nghe" --clicks 2 --output "./data/product_links.json"

# # Ví dụ 2: Cào link cho TẤT CẢ các danh mục sản phẩm của TGDD
# python crawler/crawler_links.py --all --clicks 3 --output "./data/product_links.json"
