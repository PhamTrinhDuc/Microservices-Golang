import time
import sys
import os
import argparse
import html2text
from selenium import webdriver
from selenium.webdriver.common.by import By
from selenium.webdriver.chrome.options import Options
from selenium.webdriver.chrome.service import Service
from webdriver_manager.chrome import ChromeDriverManager

sys.stdout.reconfigure(encoding='utf-8')

POLICIES = {
    "giao-hang": "https://www.thegioididong.com/giao-hang",
    "chinh-sach-bao-hanh": "https://www.thegioididong.com/bao-hanh",
    "giao-hang-thanh-toan": "https://www.thegioididong.com/giao-hang",
    "huong-dan-mua-online": "https://www.thegioididong.com/huong-dan-mua-hang",
    "chinh-sach-doi-tra": "https://www.thegioididong.com/chinh-sach-bao-hanh-san-pham",
    "quy-che-hoat-dong": "https://www.thegioididong.com/tos",
    "bao-mat-thong-tin": "https://www.thegioididong.com/chinh-sach-xu-ly-du-lieu-ca-nhan",
    "noi-quy-cua-hang": "https://www.thegioididong.com/noi-quy-cua-hang",
    "chat-luong-phuc-vu": "https://www.thegioididong.com/chat-luong-phuc-vu",
    "khui-hop-apple": "https://www.thegioididong.com/chinh-sach-khui-hop-apple",
    "mua-tra-cham": "https://www.thegioididong.com/tra-gop",
    "huong-dan-dang-bl": "https://www.thegioididong.com/huong-dan-dang-binh-luan",
    "su-dung-mxh": "https://www.thegioididong.com/thoa-thuan-su-dung-trang-mxh",
}

def init_driver():
    options = Options()
    options.add_argument("--headless")
    options.add_argument("--disable-blink-features=AutomationControlled")
    options.add_argument("--start-maximized")
    options.add_argument("user-agent=Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
    
    # Eager loading: only wait for HTML/DOM, not images/CSS
    options.page_load_strategy = 'eager'
    
    driver = webdriver.Chrome(
        service=Service(ChromeDriverManager().install()),
        options=options
    )
    
    # Limit max page load time to prevent hanging
    driver.set_page_load_timeout(25)
    return driver


def scrape_policy_text(driver, url):
    try:
        driver.get(url)
    except Exception as e:
        print(f"  -> Warning: Page load timeout for {url}, attempting to extract data anyway.")
        
    time.sleep(3)
    
    title = driver.title.strip()
    
    # Expanded selectors to cover all pages correctly
    selectors = [
        "article", 
        "div.content-detail", 
        "div.detail-content", 
        "div.policy-content", 
        "div.main-content", 
        "div.box-content", 
        "#help-main",
        "section",
        "div.content"
    ]
    
    content_html = ""
    for sel in selectors:
        try:
            elems = driver.find_elements(By.CSS_SELECTOR, sel)
            valid_elems = [el for el in elems if len(el.text.strip()) > 150]
            if valid_elems:
                # Get the HTML structure instead of flat text to preserve formatting
                content_html = valid_elems[0].get_attribute("outerHTML")
                break
        except Exception as e:
            pass
            
    if not content_html:
        # Fallback to body HTML
        try:
            body = driver.find_element(By.TAG_NAME, "body")
            content_html = body.get_attribute("outerHTML")
        except:
            pass

    if content_html:
        # Convert HTML to structured Markdown using html2text
        converter = html2text.HTML2Text()
        converter.ignore_links = False        # Keep links for reference
        converter.ignore_images = True       # Ignore images for text RAG
        converter.body_width = 0             # Disable automatic line wrapping
        converter.protect_links = True
        converter.unicode_snob = True        # Keep Vietnamese Unicode character set
        
        content_text = converter.handle(content_html).strip()
    else:
        content_text = "Không thể bóc tách nội dung chính."
            
    return title, content_text

def main():
    parser = argparse.ArgumentParser(description="TGDD Policy & Terms Crawler for RAG")
    parser.add_argument("--output-dir", default="./data/policies", help="Directory to save policy Markdown files")
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
