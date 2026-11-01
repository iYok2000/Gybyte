Act as an expert Golang developer. I need to build a web scraper and automated site audit tool in Go to evaluate if a given website URL meets the technical and structural requirements for Google AdSense approval.

Please write a complete, runnable Go program that takes a starting URL (e.g., "https://example.com") and performs the following checks. Use appropriate standard libraries (like `net/http`) and popular HTML parsing libraries like `github.com/PuerkitoBio/goquery`.

**Core Requirements for the Go Script:**

**1. Essential Pages Detection (Navigation Check):**
*   **Logic:** Fetch the homepage HTML. Scan for `<a>` tags.
*   **Validation:** Check if the site contains links to essential policy pages. Look for `href` attributes or link text matching keywords (case-insensitive): "privacy", "contact", "about", "terms".
*   **Output:** Print whether each essential page is "Found" or "Missing".

**2. Privacy Policy Compliance (Regex/Text Scraping):**
*   **Logic:** If a Privacy Policy link is found in step 1, fetch that specific page.
*   **Validation:** Scan the raw text or HTML of the Privacy Policy page for AdSense mandatory footprints:
    *   Find the string or link: "google.com/settings/ads" OR "aboutads.info".
    *   Find the keyword: "cookie" (case-insensitive).
    *   Find the keyword: "DART" or "third-party" (case-insensitive).
*   **Output:** Print "Compliant" if the required AdSense opt-out links and cookie declarations are present, otherwise print "Non-Compliant: Missing AdSense specific clauses".

**3. Basic Technical Health (HTTP Status & SSL):**
*   **Logic:** Perform a GET request to the target URL.
*   **Validation:** 
    *   Check if the scheme is `https://`.
    *   Check if the initial HTTP status code is 200 OK.
*   **Output:** Print the HTTP status code and whether the site is running on HTTPS.

**4. Content Volume Indicator (Word Count Estimation):**
*   **Logic:** Extract all text from `<p>` and `<article>` tags on the homepage (or a sample article page if provided).
*   **Validation:** Count the total number of words.
*   **Output:** Print the estimated word count. If < 300 words, flag it as "Warning: Low content volume for this page".

**5. Indexability Check:**
*   **Logic:** Fetch `[URL]/robots.txt`.
*   **Validation:** Check if the file contains `Disallow: /` which blocks all crawlers. (Simple string match is fine for this MVP).
*   **Output:** Print "Indexable" or "Blocked by robots.txt".

Please structure the Go code cleanly with separate functions for each check, handle potential HTTP or parsing errors gracefully, and provide clear CLI output summarizing the audit results.