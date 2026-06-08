package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"
)

const (
	BaseURL = "https://api-v4.yesplis.com/api/v3"
)

type Client struct {
	httpClient *http.Client
	token      string
}

type EventDetail struct {
	Name      string `json:"name"`
	Slug      string `json:"slug"`
	StartDate string `json:"start_date"`
	StartTime string `json:"start_time"`
	Status    string `json:"event_status"`
	MinPrice  int    `json:"min_price"`
	Tax       int    `json:"tax"`
	IsSeat    int    `json:"is_seat_ticket"`
}

type TicketType struct {
	ID         int    `json:"id"`
	TicketID   int    `json:"ticket_id"`
	TicketName string `json:"ticket_name"`
	Price      int    `json:"ticket_price"`
	Quantity   int    `json:"quantity"`
}

type CheckoutDetail struct {
	ExpiredAt string       `json:"expired_at"`
	Tickets   []TicketType `json:"tickets"`
}

type PaymentMethod struct {
	ID         string  `json:"id"`
	Method     string  `json:"method"`
	BankAlias  string  `json:"bank_alias"`
	Code       string  `json:"code"`
	FeePercent float64 `json:"fee_percent"`
	InetFee    int     `json:"inet_fee"`
}

type PaymentMethods struct {
	QRCode          []PaymentMethod `json:"qr_code"`
	EWallet         []PaymentMethod `json:"ewallet"`
	VirtualAccount  []PaymentMethod `json:"virtual_account"`
	Cards           []PaymentMethod `json:"cards"`
	OTC             []PaymentMethod `json:"otc"`
	PayLater        []PaymentMethod `json:"paylater"`
}

type BuyTicketRequest struct {
	EventSlug       string `json:"event_slug"`
	IsSeat          bool   `json:"is_seat"`
	PaymentMethodID string `json:"payment_method_id"`
	Tickets         []struct {
		ID           int    `json:"id"`
		TicketID     int    `json:"ticket_id"`
		TicketName   string `json:"ticket_name"`
		TicketPrice  int    `json:"ticket_price"`
		Quantity     int    `json:"quantity"`
		Participants []struct {
			SameWithUser bool `json:"same_with_user"`
		} `json:"participants"`
	} `json:"tickets"`
}

type BuyTicketResponse struct {
	Status  int    `json:"status"`
	Message string `json:"msg"`
	Data    struct {
		OrderID      string `json:"order_id"`
		TotalPrice   int    `json:"total_price"`
		AdminFee     int    `json:"admin_fee"`
		Tax          int    `json:"tax"`
		FinalPrice   int    `json:"final_price"`
		PaymentURL   string `json:"payment_url"`
		ExpiredAt    string `json:"expired_at"`
	} `json:"data"`
}

func NewClient(token string) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		token: token,
	}
}

func (c *Client) GetEventDetail(slug string) (*EventDetail, error) {
	url := fmt.Sprintf("%s/public/events/detail/%s", BaseURL, slug)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	c.setHeaders(req, slug, "tickets")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result struct {
		Status int          `json:"status"`
		Msg    string       `json:"msg"`
		Data   *EventDetail `json:"data"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	if result.Status != 1 {
		return nil, fmt.Errorf("API error: %s", result.Msg)
	}

	return result.Data, nil
}

func (c *Client) GetCheckoutDetail(slug string) (*CheckoutDetail, error) {
	url := fmt.Sprintf("%s/transaction/checkout/detail", BaseURL)

	payload := map[string]string{"event_slug": slug}
	jsonData, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	c.setHeaders(req, slug, "checkout")
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result struct {
		Status int            `json:"status"`
		Msg    string         `json:"msg"`
		Data   *CheckoutDetail `json:"data"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	if result.Status != 1 {
		return nil, fmt.Errorf("API error: %s", result.Msg)
	}

	return result.Data, nil
}

func (c *Client) GetPaymentMethods(slug string) (*PaymentMethods, error) {
	url := fmt.Sprintf("%s/public/payment-methods/%s", BaseURL, slug)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	c.setHeaders(req, slug, "checkout")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result struct {
		Status int             `json:"status"`
		Msg    string          `json:"msg"`
		Data   *PaymentMethods `json:"data"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	if result.Status != 1 {
		return nil, fmt.Errorf("API error: %s", result.Msg)
	}

	return result.Data, nil
}

func (c *Client) BuyTicket(slug string, ticket TicketType, paymentMethodID string, isSeat bool) (*BuyTicketResponse, error) {
	url := fmt.Sprintf("%s/transaction/buy-ticket", BaseURL)

	payload := BuyTicketRequest{
		EventSlug:       slug,
		IsSeat:          isSeat,
		PaymentMethodID: paymentMethodID,
		Tickets: []struct {
			ID           int    `json:"id"`
			TicketID     int    `json:"ticket_id"`
			TicketName   string `json:"ticket_name"`
			TicketPrice  int    `json:"ticket_price"`
			Quantity     int    `json:"quantity"`
			Participants []struct {
				SameWithUser bool `json:"same_with_user"`
			} `json:"participants"`
		}{
			{
				ID:          ticket.ID,
				TicketID:    ticket.TicketID,
				TicketName:  ticket.TicketName,
				TicketPrice: ticket.Price,
				Quantity:    1,
				Participants: []struct {
					SameWithUser bool `json:"same_with_user"`
				}{
					{SameWithUser: true},
				},
			},
		},
	}

	jsonData, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	c.setHeaders(req, slug, "checkout")
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result BuyTicketResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	if result.Status != 1 {
		return nil, fmt.Errorf("API error: %s", result.Message)
	}

	return &result, nil
}

func (c *Client) setHeaders(req *http.Request, slug string, page string) {
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:151.0) Gecko/20100101 Firefox/151.0")
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Origin", "https://www.yesplis.com")
	req.Header.Set("Referer", "https://www.yesplis.com/")
	req.Header.Set("Cookie", fmt.Sprintf("access_token=%s", c.token))
	req.Header.Set("yp-page-code", fmt.Sprintf("https://www.yesplis.com/event/%s/%s", slug, page))
}

type WarBot struct {
	client         *Client
	eventSlug      string
	ticketCategory string
	quantity       int
	timeout        time.Duration
	paymentMethod  string
}

func NewWarBot(token, eventSlug, ticketCategory, paymentMethod string, quantity int, timeout time.Duration) *WarBot {
	return &WarBot{
		client:         NewClient(token),
		eventSlug:      eventSlug,
		ticketCategory: ticketCategory,
		quantity:       quantity,
		timeout:        timeout,
		paymentMethod:  paymentMethod,
	}
}

func (w *WarBot) Run(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, w.timeout)
	defer cancel()

	// Step 1: Get event details
	log.Printf("Getting event details for: %s", w.eventSlug)
	event, err := w.client.GetEventDetail(w.eventSlug)
	if err != nil {
		return fmt.Errorf("get event detail: %w", err)
	}
	log.Printf("Event: %s", event.Name)
	log.Printf("Status: %s", event.Status)
	log.Printf("Tax: %d%%", event.Tax)

	// Step 2: Get payment methods
	log.Printf("Getting payment methods...")
	methods, err := w.client.GetPaymentMethods(w.eventSlug)
	if err != nil {
		return fmt.Errorf("get payment methods: %w", err)
	}

	// Find payment method ID
	var paymentMethodID string
	allMethods := append(methods.QRCode, methods.EWallet...)
	allMethods = append(allMethods, methods.VirtualAccount...)
	allMethods = append(allMethods, methods.Cards...)
	allMethods = append(allMethods, methods.OTC...)
	allMethods = append(allMethods, methods.PayLater...)

	for _, m := range allMethods {
		if strings.EqualFold(m.Code, w.paymentMethod) || strings.EqualFold(m.BankAlias, w.paymentMethod) {
			paymentMethodID = m.ID
			log.Printf("Payment method: %s (Fee: %.1f%% + Rp %d)", m.BankAlias, m.FeePercent, m.InetFee)
			break
		}
	}

	if paymentMethodID == "" {
		return fmt.Errorf("payment method not found: %s", w.paymentMethod)
	}

	// Step 3: War mode - poll for tickets
	log.Printf("Waiting for tickets (War Mode)...")
	log.Printf("Target: %s x%d", w.ticketCategory, w.quantity)

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			// Try to buy ticket directly
			log.Printf("Attempting to buy ticket...")

			// First get checkout detail to see available tickets
			checkout, err := w.client.GetCheckoutDetail(w.eventSlug)
			if err != nil {
				log.Printf("Checkout detail error: %v", err)
				continue
			}

			// Find matching ticket
			var targetTicket *TicketType
			for _, t := range checkout.Tickets {
				if strings.Contains(strings.ToLower(t.TicketName), strings.ToLower(w.ticketCategory)) {
					targetTicket = &t
					break
				}
			}

			if targetTicket == nil {
				log.Printf("No matching ticket found, retrying...")
				continue
			}

			log.Printf("Found ticket: %s (Price: Rp %d)", targetTicket.TicketName, targetTicket.Price)

			// Try to buy
			result, err := w.client.BuyTicket(w.eventSlug, *targetTicket, paymentMethodID, event.IsSeat == 1)
			if err != nil {
				if strings.Contains(err.Error(), "try again") || strings.Contains(err.Error(), "sold out") || strings.Contains(err.Error(), "not available") {
					log.Printf("Ticket not available, retrying...")
					continue
				}
				return fmt.Errorf("buy ticket: %w", err)
			}

			// Success!
			log.Printf("SUCCESS! Order ID: %s", result.Data.OrderID)
			log.Printf("Total: Rp %d", result.Data.FinalPrice)
			if result.Data.PaymentURL != "" {
				log.Printf("Payment URL: %s", result.Data.PaymentURL)
			}
			return nil
		}
	}
}

func main() {
	token := flag.String("token", "", "Access token (JWT)")
	eventSlug := flag.String("event", "", "Event slug (e.g., harbour-fest-2026)")
	ticketCategory := flag.String("ticket", "", "Ticket category name")
	paymentMethod := flag.String("payment", "QRIS", "Payment method (QRIS, OVO, BCA, etc.)")
	quantity := flag.Int("qty", 1, "Quantity of tickets")
	timeout := flag.Duration("timeout", 2*time.Minute, "Global timeout")

	flag.Parse()

	if *token == "" {
		log.Fatal("Token is required. Use -token flag.")
	}
	if *eventSlug == "" {
		log.Fatal("Event slug is required. Use -event flag.")
	}
	if *ticketCategory == "" {
		log.Fatal("Ticket category is required. Use -ticket flag.")
	}

	bot := NewWarBot(*token, *eventSlug, *ticketCategory, *paymentMethod, *quantity, *timeout)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	if err := bot.Run(ctx); err != nil {
		log.Fatalf("War failed: %v", err)
	}

	log.Println("War ticket success!")
}
