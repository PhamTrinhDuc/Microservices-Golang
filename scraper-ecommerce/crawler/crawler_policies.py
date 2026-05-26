import time
import sys
import os
import argparse
from selenium import webdriver
from selenium.webdriver.common.by import By
from selenium.webdriver.chrome.options import Options
from selenium.webdriver.chrome.service import Service
from webdriver_manager.chrome import ChromeDriverManager

sys.stdout.reconfigure(encoding='utf-8')

POLICIES = {
    "chinh-sach-bao-hanh": "https://www.thegioididong.com/bao-hanh",
    "chinh-sach-doi-tra": "https://www.thegioididong.com/chinh-sach-bao-hanh-san-pham",
    "giao-hang-thanh-toan": "https://www.thegioididong.com/giao-hang",
    "huong-dan-mua-online": "https://www.thegioididong.com/huong-dan-mua-hang",
    "quy-che-hoat-dong": "https://www.thegioididong.com/tos",
    "bao-mat-thong-tin": "https://www.thegioididong.com/chinh-sach-xu-ly-du-lieu-ca-nhan",
    "noi-quy-cua-hang": "https://www.thegioididong.com/noi-quy-cua-hang",
    "chat-luong-phuc-vu": "https://www.thegioididong.com/chat-luong-phuc-vu",
    "khui-hop-apple": "https://www.thegioididong.com/chinh-sach-khui-hop-apple",
    "mua-tra-cham": "https://www.thegioididong.com/tra-gop"
}

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

def scrape_policy_text(driver, url):
    driver.get(url)
    time.sleep(4)
    
    title = driver.title.strip()
    
    # Selectors for main content
    selectors = [
        "article", 
        "div.content-detail", 
        "div.detail-content", 
        "div.policy-content", 
        "div.main-content", 
        "div.box-content", 
        "div.content"
    ]
    
    content_text = ""
    for sel in selectors:
        try:
            elems = driver.find_elements(By.CSS_SELECTOR, sel)
            valid_elems = [el for el in elems if len(el.text.strip()) > 150]
            if valid_elems:
                # Use get_attribute("textContent") to get complete text
                raw_text = valid_elems[0].get_attribute("textContent").strip()
                # Clean duplicate lines and format nicely
                lines = [line.strip() for line in raw_text.split('\n')]
                # Filter empty lines and join
                content_text = "\n".join([line for line in lines if line])
                break
        except Exception as e:
            pass
            
    if not content_text:
        # Fallback to body text
        try:
            body = driver.find_element(By.TAG_NAME, "body")
            raw_text = body.get_attribute("textContent").strip()
            lines = [line.strip() for line in raw_text.split('\n')]
            content_text = "\n".join([line for line in lines if line])
        except:
            content_text = "Không thể bóc tách nội dung chính."
            
    return title, content_text

def main():
    parser = argparse.ArgumentParser(description="TGDD Policy & Terms Crawler for RAG")
    parser.add_argument("--output-dir", default="../data/policies", help="Directory to save policy Markdown files")
    parser.add_argument("--limit", type=int, help="Limit the number of policies to crawl (for quick testing)")
    
    args = parser.parse_args()
    
    os.makedirs(args.output_dir, exist_ok=True)
    
    driver = init_driver()
    
    targets = list(POLICIES.items())
    if args.limit:
        targets = targets[:args.limit]
        
    print(f"Starting Policy Crawler. Target policies to scrape: {len(targets)}")
    
    success_count = 0
    try:
        for idx, (name, url) in enumerate(targets):
            print(f"\n[{idx+1}/{len(targets)}] Scraping policy: {name}...")
            print(f"URL: {url}")
            try:
                title, text = scrape_policy_text(driver, url)
                
                # Format to Markdown
                md_content = f"# {title}\n"
                md_content += f"**Source URL**: {url}\n\n"
                md_content += text
                
                output_file = os.path.join(args.output_dir, f"{name}.md")
                with open(output_file, "w", encoding="utf-8") as f:
                    f.write(md_content)
                    
                print(f"  -> Successfully saved to: {output_file} ({len(text)} characters)")
                success_count += 1
                time.sleep(2)
            except Exception as e:
                print(f"  -> Error scraping {name}: {e}")
                
    finally:
        driver.quit()
        
    print(f"\nFinished scraping. Successfully extracted {success_count}/{len(targets)} policy files.")
    print(f"Saved policy Markdown database under: {os.path.abspath(args.output_dir)}")

if __name__ == "__main__":
    main()
