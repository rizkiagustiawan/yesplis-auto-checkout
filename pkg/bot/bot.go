package bot

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

type Config struct {
	EventURL       string
	TicketCategory string
	Quantity       int
	Headless       bool
	CookieFile     string
}

type Bot struct {
	config Config
}

func NewBot(cfg Config) *Bot {
	if cfg.CookieFile == "" {
		cfg.CookieFile = "cookies.json"
	}
	return &Bot{config: cfg}
}

func (b *Bot) Run(ctx context.Context) error {
	// Enterprise Stealth: Disabling automation flags
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.NoFirstRun,
		chromedp.NoDefaultBrowserCheck,
		chromedp.DisableGPU,
		chromedp.Flag("disable-blink-features", "AutomationControlled"),
		chromedp.Flag("disable-infobars", true),
		chromedp.Flag("excludeSwitches", "enable-automation"),
	)

	if !b.config.Headless {
		opts = append(opts, chromedp.Flag("headless", false))
	} else {
		// New headless mode handles anti-detect better
		opts = append(opts, chromedp.Flag("headless", "new"))
	}

	allocCtx, cancel := chromedp.NewExecAllocator(ctx, opts...)
	defer cancel()

	ctx, cancel = chromedp.NewContext(allocCtx)
	defer cancel()

	// Enterprise Stealth: Inject script to mock human fingerprints
	stealthScript := `
		Object.defineProperty(navigator, 'webdriver', { get: () => undefined });
		window.chrome = { runtime: {} };
		Object.defineProperty(navigator, 'languages', { get: () => ['en-US', 'en'] });
		Object.defineProperty(navigator, 'plugins', { get: () => [1, 2, 3, 4, 5] });
	`

	// 0. Load Cookies
	if _, err := os.Stat(b.config.CookieFile); err == nil {
		log.Println("Loading cookies...")
		err = b.loadCookies(ctx)
		if err != nil {
			log.Printf("Failed to load cookies: %v", err)
		}
	}

	log.Printf("Navigating to %s...", b.config.EventURL)

	err := chromedp.Run(ctx,
		chromedp.ActionFunc(func(ctx context.Context) error {
			_, err := page.AddScriptToEvaluateOnNewDocument(stealthScript).Do(ctx)
			return err
		}),
		chromedp.Navigate(b.config.EventURL),
	)
	if err != nil {
		return fmt.Errorf("failed to navigate: %v", err)
	}

	// 1. Wait for "Beli Tiket" button and click it
	log.Println("Waiting for 'Beli Tiket' button (War Mode)...")
	// Robust case-insensitive selectors
	selectors := []string{
		`//button[contains(translate(text(), 'ABCDEFGHIJKLMNOPQRSTUVWXYZ', 'abcdefghijklmnopqrstuvwxyz'), 'beli tiket')]`,
		`//a[contains(translate(text(), 'ABCDEFGHIJKLMNOPQRSTUVWXYZ', 'abcdefghijklmnopqrstuvwxyz'), 'beli tiket')]`,
		`.btn-buy-ticket`,
		`#btn-buy-ticket`,
		`//button[contains(translate(text(), 'ABCDEFGHIJKLMNOPQRSTUVWXYZ', 'abcdefghijklmnopqrstuvwxyz'), 'buy ticket')]`,
		`//a[contains(translate(text(), 'ABCDEFGHIJKLMNOPQRSTUVWXYZ', 'abcdefghijklmnopqrstuvwxyz'), 'buy ticket')]`,
	}

	for {
		var found bool
		var targetSelector string

		for _, s := range selectors {
			err = chromedp.Run(ctx,
				chromedp.Evaluate(b.jsSelectorCheck(s), &found),
			)
			if err == nil && found {
				targetSelector = s
				break
			}
		}

		if found {
			log.Printf("Button found with selector: %s! Clicking...", targetSelector)
			err = b.trustedClickXY(ctx, targetSelector)
			if err == nil {
				break
			} else {
				log.Printf("Click failed: %v", err)
			}
		}

		log.Println("Button not found. Refreshing in 300ms...")
		time.Sleep(300 * time.Millisecond)
		err = chromedp.Run(ctx, chromedp.Reload())
		if err != nil {
			log.Printf("Reload failed: %v", err)
		}
	}

	// 2. Select Ticket Category and Quantity
	log.Printf("Selecting ticket category: %s...", b.config.TicketCategory)

	lowerCategory := strings.ToLower(b.config.TicketCategory)
	plusButtonXPath := fmt.Sprintf(`//div[(contains(@class, 'ticket') or contains(@class, 'card') or contains(@class, 'item') or contains(@class, 'list')) and contains(translate(., 'ABCDEFGHIJKLMNOPQRSTUVWXYZ', 'abcdefghijklmnopqrstuvwxyz'), '%s')]//button[contains(@class, 'inc') or contains(@class, 'plus') or contains(text(), '+')]`, lowerCategory)

	// We will poll for the plus button instead of WaitVisible to be immune to page load hangs
	log.Println("Waiting for ticket category to appear...")
	for {
		var found bool
		err = chromedp.Run(ctx, chromedp.Evaluate(b.jsSelectorCheck(plusButtonXPath), &found))
		if err == nil && found {
			break
		}
		time.Sleep(150 * time.Millisecond) // Fast polling
	}

	log.Printf("Adding %d tickets...", b.config.Quantity)
	for i := 0; i < b.config.Quantity; i++ {
		err = b.trustedClickXY(ctx, plusButtonXPath)
		if err != nil {
			return fmt.Errorf("failed to click plus button for %s: %v", b.config.TicketCategory, err)
		}
		time.Sleep(100 * time.Millisecond) // Slight jitter for stealth
	}

	// 3. Click Checkout/Bayar
	log.Println("Proceeding to checkout...")
	checkoutSelectors := []string{
		`//button[contains(translate(text(), 'ABCDEFGHIJKLMNOPQRSTUVWXYZ', 'abcdefghijklmnopqrstuvwxyz'), 'checkout')]`,
		`//button[contains(translate(text(), 'ABCDEFGHIJKLMNOPQRSTUVWXYZ', 'abcdefghijklmnopqrstuvwxyz'), 'bayar')]`,
		`.btn-checkout`,
		`.btn-buy`,
		`//button[contains(translate(text(), 'ABCDEFGHIJKLMNOPQRSTUVWXYZ', 'abcdefghijklmnopqrstuvwxyz'), 'lanjut')]`,
		`//a[contains(translate(text(), 'ABCDEFGHIJKLMNOPQRSTUVWXYZ', 'abcdefghijklmnopqrstuvwxyz'), 'checkout')]`,
	}

	err = b.waitForAnyAndTrustedClick(ctx, checkoutSelectors)
	if err != nil {
		return fmt.Errorf("failed to click checkout: %v", err)
	}

	// 4. Save Cookies after successful steps
	log.Println("Saving cookies...")
	err = b.saveCookies(ctx)
	if err != nil {
		log.Printf("Failed to save cookies: %v", err)
	}

	log.Println("Reached checkout page. Please complete payment manually.")
	time.Sleep(30 * time.Minute) // Keep browser open for user to pay

	return nil
}

func (b *Bot) jsSelectorCheck(selector string) string {
	return fmt.Sprintf(`(function() {
		const s = "%s";
		try {
			let node = null;
			if (s.startsWith("//") || s.startsWith("(")) {
				node = document.evaluate(s, document, null, XPathResult.FIRST_ORDERED_NODE_TYPE, null).singleNodeValue;
			} else {
				node = document.querySelector(s);
			}
			if (!node) return false;
			return !!(node.offsetWidth || node.offsetHeight || node.getClientRects().length);
		} catch(e) {
			return false;
		}
	})()`, selector)
}

func (b *Bot) trustedClickXY(ctx context.Context, selector string) error {
	var res map[string]interface{}
	err := chromedp.Run(ctx,
		chromedp.Evaluate(fmt.Sprintf(`(function() {
			const s = "%s";
			let el = null;
			if (s.startsWith("//") || s.startsWith("(")) {
				el = document.evaluate(s, document, null, XPathResult.FIRST_ORDERED_NODE_TYPE, null).singleNodeValue;
			} else {
				el = document.querySelector(s);
			}
			if (el) {
				const rect = el.getBoundingClientRect();
				return {
					x: rect.left + (rect.width / 2),
					y: rect.top + (rect.height / 2),
					found: true
				};
			}
			return {found: false};
		})()`, selector), &res),
	)
	if err != nil {
		return err
	}
	if found, ok := res["found"].(bool); ok && found {
		x := res["x"].(float64)
		y := res["y"].(float64)
		// Dispatch physical mouse click (isTrusted: true)
		return chromedp.Run(ctx, chromedp.MouseClickXY(x, y))
	}
	return fmt.Errorf("element not found")
}

func (b *Bot) waitForAnyAndTrustedClick(ctx context.Context, selectors []string) error {
	for i := 0; i < 150; i++ { // Wait for up to 30 seconds (150 * 200ms)
		for _, s := range selectors {
			var found bool
			err := chromedp.Run(ctx,
				chromedp.Evaluate(b.jsSelectorCheck(s), &found),
			)
			if err == nil && found {
				errClick := b.trustedClickXY(ctx, s)
				if errClick == nil {
					return nil
				}
			}
		}
		time.Sleep(200 * time.Millisecond)
	}
	return fmt.Errorf("none of the checkout selectors became visible")
}

func (b *Bot) saveCookies(ctx context.Context) error {
	var cookies []*network.Cookie
	err := chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
		var err error
		cookies, err = network.GetCookies().Do(ctx)
		return err
	}))
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(cookies, "", "  ")
	if err != nil {
		return err
	}

	return ioutil.WriteFile(b.config.CookieFile, data, 0644)
}

func (b *Bot) loadCookies(ctx context.Context) error {
	data, err := ioutil.ReadFile(b.config.CookieFile)
	if err != nil {
		return err
	}

	var cookies []*network.Cookie
	err = json.Unmarshal(data, &cookies)
	if err != nil {
		return err
	}

	return chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
		for _, c := range cookies {
			err := network.SetCookie(c.Name, c.Value).
				WithDomain(c.Domain).
				WithPath(c.Path).
				WithHTTPOnly(c.HTTPOnly).
				WithSecure(c.Secure).
				Do(ctx)
			if err != nil {
				return err
			}
		}
		return nil
	}))
}
