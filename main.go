package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"yesplis-auto-order-ticket/pkg/bot"
)

func main() {
	eventURL := flag.String("url", "", "Yesplis event URL")
	ticketCategory := flag.String("ticket", "", "Ticket category name")
	quantity := flag.Int("qty", 1, "Quantity of tickets")
	headless := flag.Bool("headless", false, "Run in headless mode")

	flag.Parse()

	if *eventURL == "" {
		log.Fatal("Event URL is required. Use -url flag.")
	}
	if *ticketCategory == "" {
		log.Fatal("Ticket category name is required. Use -ticket flag.")
	}

	cfg := bot.Config{
		EventURL:       *eventURL,
		TicketCategory: *ticketCategory,
		Quantity:       *quantity,
		Headless:       *headless,
	}

	b := bot.NewBot(cfg)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	if err := b.Run(ctx); err != nil {
		log.Fatalf("Bot failed: %v", err)
	}

	log.Println("Bot finished successfully.")
}
