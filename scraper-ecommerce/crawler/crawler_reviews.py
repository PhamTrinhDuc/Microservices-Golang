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

def scrape_reviews_for_product(driver, product_url, max_pages=3, max_reviews=50):
    # Reviews URL is constructed by appending /danh-gia to the product detail page URL
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
        driver.get(page_url)
        time.sleep(3)
        
        try:
                
            comment_list = driver.find_elements(By.CSS_SELECTOR, "ul.comment-list li.par")
            print(f"    Found {len(comment_list)} reviews on this page.")
            
            if not comment_list:
                # No more reviews or pagination reached the end
                break
                
            for li in comment_list:
                if len(reviews) >= max_reviews:
                    break
                    
                try:
                    # Review ID
                    review_id = li.get_attribute("id")
                    
                    # Reviewer Name
                    name_elem = li.find_element(By.CSS_SELECTOR, ".cmt-top-name")
                    reviewer_name = name_elem.get_attribute("textContent").strip()
                    
                    # Rating Stars
                    star_elements = li.find_elements(By.CSS_SELECTOR, ".cmt-top-star i.iconcmt-starbuy, .cmt-top-star i.iconcmt-star")
                    # If no star element is found under that selector, fallback to counting all star icons in cmt-intro
                    if not star_elements:
                        star_elements = li.find_elements(By.CSS_SELECTOR, ".cmt-intro i[class*='star']")
                    rating = len(star_elements)
                    if rating == 0:
                        rating = 5 # default fallback
                        
                    # Comment Text
                    try:
                        comment_elem = li.find_element(By.CSS_SELECTOR, ".cmt-txt")
                        comment = comment_elem.get_attribute("textContent").strip()
                    except:
                        comment = ""
                        
                    # Usage duration or relative time
                    usage_time = ""
                    try:
                        time_elem = li.find_element(By.CSS_SELECTOR, ".cmtd, .cmt-time, time")
                        usage_time = time_elem.get_attribute("textContent").strip()
                    except:
                        pass
                        
                    # User-uploaded images
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
                    except:
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
                except Exception as item_err:
                    pass
        except Exception as page_err:
            print(f"    Error parsing page {page}: {page_err}")
            break
            
    print(f"Scraped {len(reviews)} reviews successfully.")
    return reviews

def main():
    parser = argparse.ArgumentParser(description="TGDD Customer Reviews Scraper")
    parser.add_argument("--url", default="https://www.thegioididong.com/dtdd/iphone-17", help="Product detail page URL or reviews page URL")
    parser.add_argument("--pages", type=int, default=3, help="Max number of pages to scrape")
    parser.add_argument("--limit", type=int, default=30, help="Max total reviews to scrape")
    parser.add_argument("--output", default="../data/reviews.json", help="Path to save output JSON file")
    
    args = parser.parse_args()
    
    driver = init_driver()
    try:
        reviews_data = scrape_reviews_for_product(driver, args.url, args.pages, args.limit)
        
        # Save output
        output_dir = os.path.dirname(args.output)
        if output_dir:
            os.makedirs(output_dir, exist_ok=True)
        
        # Get product name
        product_name = None
        for selector in ["div.box-product h3", "ul.breadcrumb-rating a", "div.product-name h1", "section.detail h1"]:
            try:
                product_name = driver.find_element(By.CSS_SELECTOR, selector).get_attribute("textContent").strip()
                if product_name:
                    break
            except Exception:
                continue
        
        if not product_name:
            product_name = "Unknown Product"
        
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
