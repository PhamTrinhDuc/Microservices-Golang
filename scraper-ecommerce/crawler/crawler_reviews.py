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

def get_product_slug(url):
    # E.g. "https://www.thegioididong.com/dtdd/iphone-17" -> "iphone-17"
    url = url.rstrip("/")
    slug = url.split("/")[-1]
    # Remove query params if any
    slug = slug.split("?")[0]
    return slug

def parse_comment_elements(comment_elements):
    reviews = []
    for li in comment_elements:
        try:
            # 1. Review ID
            review_id = li.get_attribute("id") or ""
            
            # 2. Reviewer Name
            reviewer_name = "Người dùng ẩn danh"
            for selector in [".cmt-top-name", ".cmt-intro p.cmt-top-name", "b"]:
                try:
                    name_elem = li.find_element(By.CSS_SELECTOR, selector)
                    name_txt = name_elem.get_attribute("textContent").strip()
                    if name_txt:
                        reviewer_name = name_txt
                        break
                except Exception:
                    continue
            
            # 3. Rating Stars
            star_elements = li.find_elements(By.CSS_SELECTOR, ".cmt-top-star i.iconcmt-starbuy, .cmt-top-star i.iconcmt-star")
            if not star_elements:
                star_elements = li.find_elements(By.CSS_SELECTOR, ".cmt-intro i[class*='star']")
            rating = len(star_elements) if star_elements else 5
            
            # 4. Comment Text
            comment = ""
            for selector in [".cmt-txt", ".cmt-content p"]:
                try:
                    comment_elem = li.find_element(By.CSS_SELECTOR, selector)
                    comment_txt = comment_elem.get_attribute("textContent").strip()
                    if comment_txt:
                        comment = comment_txt
                        break
                except Exception:
                    continue
                    
            # 5. Usage duration or relative time
            usage_time = ""
            try:
                time_elem = li.find_element(By.CSS_SELECTOR, ".cmtd, .cmt-time, time")
                usage_time = time_elem.get_attribute("textContent").strip()
            except Exception:
                pass
                
            # 6. User-uploaded images
            images = []
            try:
                img_elements = li.find_elements(By.CSS_SELECTOR, "img.commentImg, div.cmt-content-img img, .comment-img img, .gallery-rating img")
                for img in img_elements:
                    src = img.get_attribute("src") or img.get_attribute("data-src")
                    if src:
                        if src.startswith("//"):
                            src = "https:" + src
                        if src not in images and not src.endswith(".gif") and "avatar" not in src:
                            images.append(src)
            except Exception:
                pass
                
            reviews.append({
                "review_id": review_id,
                "reviewer_name": reviewer_name,
                "rating": rating,
                "comment": comment,
                "usage_time": usage_time,
                "images": images,
                "status": "approved"
            })
        except Exception:
            continue
    return reviews

def scrape_reviews_for_product(driver, product_url, max_pages=3, max_reviews=50):
    clean_product_url = product_url.rstrip("/").split("?")[0]
    base_review_url = f"{clean_product_url}/danh-gia"
    
    print(f"Scraping reviews for product URL: {product_url}...")
    print(f"Base Review Page: {base_review_url}")
    
    reviews = []
    
    for page in range(1, max_pages + 1):
        if len(reviews) >= max_reviews:
            break
            
        page_url = f"{base_review_url}?page={page}"
        print(f"  Fetching page {page}: {page_url}...")
        
        try:
            driver.get(page_url)
            time.sleep(3)
            
            # Check if redirect happened to something without '/danh-gia'
            current_url = driver.current_url
            if "/danh-gia" not in current_url:
                print(f"    Page redirected to {current_url}. No dedicated review page found.")
                break
                
            comment_elements = driver.find_elements(By.CSS_SELECTOR, "ul.comment-list li.par")
            print(f"    Found {len(comment_elements)} review elements on page {page}.")
            
            if not comment_elements:
                break
                
            page_reviews = parse_comment_elements(comment_elements)
            reviews.extend(page_reviews)
            
        except Exception as page_err:
            print(f"    Error parsing page {page}: {page_err}")
            break
            
    print(f"Scraped {len(reviews)} reviews from subpages successfully.")
    return reviews

def scrape_product_name_and_reviews(driver, url, max_pages=3, max_reviews=50):
    # 1. Visit the main product page first to extract the accurate product name & initial reviews
    print(f"Fetching main product page: {url}")
    product_name = "Unknown Product"
    main_page_reviews = []
    
    try:
        driver.get(url)
        time.sleep(3)
        
        # Extract product name
        for selector in ["div.product-name h1", "section.detail h1", "div.box-product h3", "ul.breadcrumb-rating a"]:
            try:
                elem = driver.find_element(By.CSS_SELECTOR, selector)
                name_txt = elem.get_attribute("textContent").strip()
                if name_txt:
                    product_name = name_txt
                    break
            except Exception:
                continue
                
        if product_name == "Unknown Product" or not product_name:
            raw_title = driver.title.strip()
            cleaned_title = re.sub(r'^(?:\d+\s+review,\s+đánh\s+giá\s+|Mua\s+|Đánh\s+giá\s+|review,\s+đánh\s+giá\s+)', '', raw_title, flags=re.IGNORECASE)
            parts = re.split(r'\s+(?:nhận|giá tốt|tặng|trả góp|từ người|cấu hình|có tốt|uy tín|chính hãng|-)\b', cleaned_title, flags=re.IGNORECASE)
            product_name = parts[0].strip() if parts else "Unknown Product"
            
        # Parse reviews directly visible on the main product detail page
        main_comment_elements = driver.find_elements(By.CSS_SELECTOR, "ul.comment-list li.par")
        if main_comment_elements:
            print(f"  Found {len(main_comment_elements)} reviews directly on main page.")
            main_page_reviews = parse_comment_elements(main_comment_elements)
            
    except Exception as e:
        print(f"  Error loading main product page: {e}")
        
    if not product_name or product_name.lower() in ["lỗi", "error", "untitled", "không tìm thấy trang", "404"]:
        product_name = "Unknown Product"
        
    print(f"  -> Extracted Product Name: {product_name}")
    
    # 2. Scrape reviews from /danh-gia page
    subpage_reviews = []
    try:
        subpage_reviews = scrape_reviews_for_product(driver, url, max_pages, max_reviews)
    except Exception as sub_err:
        print(f"  Error loading subpages /danh-gia: {sub_err}")
        
    # 3. Merge and de-duplicate reviews
    all_reviews = main_page_reviews + subpage_reviews
    seen_keys = set()
    unique_reviews = []
    
    for r in all_reviews:
        # Define a unique key for deduplication
        r_id = r.get("review_id", "")
        if r_id:
            key = f"id_{r_id}"
        else:
            # Fallback key using reviewer name, rating and a hash-like representation of comment text
            comment_text = r.get("comment", "")
            key = f"text_{r.get('reviewer_name')}_{len(comment_text)}_{r.get('rating')}"
            
        if key not in seen_keys:
            seen_keys.add(key)
            unique_reviews.append(r)
            
    # Truncate to max_reviews if exceeded
    unique_reviews = unique_reviews[:max_reviews]
    print(f"  -> Total unique reviews merged: {len(unique_reviews)}")
    
    return product_name, unique_reviews

def main():
    parser = argparse.ArgumentParser(description="TGDD Customer Reviews Scraper")
    parser.add_argument("--url", default="https://www.thegioididong.com/dtdd/iphone-17", help="Product detail page URL or reviews page URL")
    parser.add_argument("--pages", type=int, default=3, help="Max number of pages to scrape per product")
    parser.add_argument("--limit", type=int, default=30, help="Max total reviews to scrape per product")
    parser.add_argument("--file", help="Path to a JSON file containing list of product URLs or product details objects")
    parser.add_argument("--limit-products", type=int, help="Max number of products to crawl reviews for when using --file")
    parser.add_argument("--output", default="./data/reviews.json", help="Path to save output JSON file")
    
    args = parser.parse_args()
    
    # Setup output dir
    output_dir = os.path.dirname(args.output)
    if output_dir:
        os.makedirs(output_dir, exist_ok=True)
        
    driver = init_driver()
    try:
        if args.file:
            print(f"Reading product links from: {args.file}...")
            try:
                with open(args.file, "r", encoding="utf-8") as f:
                    raw_data = json.load(f)
            except Exception as e:
                print(f"Error reading file {args.file}: {e}")
                sys.exit(1)
                
            # Extract URLs
            product_urls = []
            if isinstance(raw_data, list):
                for item in raw_data:
                    if isinstance(item, str):
                        product_urls.append(item)
                    elif isinstance(item, dict):
                        # check in products.json structure
                        url = item.get("source_url") or item.get("product_info", {}).get("source_url")
                        if url:
                            product_urls.append(url)
            
            # De-duplicate links
            product_urls = list(dict.fromkeys(product_urls))
            
            if args.limit_products:
                product_urls = product_urls[:args.limit_products]
                
            print(f"Found {len(product_urls)} unique product URLs to scrape reviews for.")
            batch_results = []
            for idx, url in enumerate(product_urls):
                print(f"\n[{idx+1}/{len(product_urls)}] Scraping reviews for: {url}")
                try:
                    product_name, reviews_data = scrape_product_name_and_reviews(driver, url, args.pages, args.limit)
                    
                    batch_results.append({
                        "product_name": product_name,
                        "product_url": url,
                        "product_slug": get_product_slug(url),
                        "reviews": reviews_data
                    })
                    time.sleep(2)
                except Exception as e:
                    print(f"  Error scraping reviews for {url}: {e}")
                    
            with open(args.output, "w", encoding="utf-8") as f:
                json.dump(batch_results, f, indent=4, ensure_ascii=False)
                
            total_reviews = sum(len(x["reviews"]) for x in batch_results)
            print(f"\nBatch finished. Successfully saved {total_reviews} reviews across {len(batch_results)} products to: {os.path.abspath(args.output)}")
            
        else:
            # Single product reviews crawling
            product_name, reviews_data = scrape_product_name_and_reviews(driver, args.url, args.pages, args.limit)
            
            result = {
                "product_name": product_name,
                "product_url": args.url,
                "product_slug": get_product_slug(args.url),
                "reviews": reviews_data
            }
            
            with open(args.output, "w", encoding="utf-8") as f:
                json.dump(result, f, indent=4, ensure_ascii=False)
                
            print(f"\nSuccessfully saved {len(reviews_data)} reviews to: {os.path.abspath(args.output)}")
            
    finally:
        driver.quit()

if __name__ == "__main__":
    main()
    
# python crawler/crawler_reviews.py --file "./data/product_links.json" --limit-product 5