package routes

import (
	"encoding/json"
	"github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/checkout/session"
	"github.com/stripe/stripe-go/v82/price"
	"log"
	"net/http"
	"os"
)

func CreateCheckoutSession(w http.ResponseWriter, r *http.Request) {
	stripeKey := os.Getenv("STRIPE_SECRET_KEY")
	if stripeKey == "" {
		log.Println("❌ STRIPE_SECRET_KEY ist leer oder nicht geladen!")
		writeJSON(w, map[string]string{"error": "Stripe secret key not set"}, http.StatusInternalServerError)
		return
	}

	stripe.Key = stripeKey
	log.Println("✅ STRIPE_SECRET_KEY erfolgreich geladen.")

	// Optional: Zum Debuggen aller verfügbaren Preise im aktuellen Konto
	debugPrices()

	domain := "https://educonnect-bmsd22a.bbzwinf.ch" // Frontend-URL

	params := &stripe.CheckoutSessionParams{
		UIMode:    stripe.String("embedded"),
		ReturnURL: stripe.String(domain + "/dashboard"),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				Price:    stripe.String("price_1RBbSa01266L6uW7c54LKNCg"), // <-- gültige Preis-ID hier
				Quantity: stripe.Int64(1),
			},
		},
		Mode: stripe.String(string(stripe.CheckoutSessionModeSubscription)), // für Abo-Preis
		AutomaticTax: &stripe.CheckoutSessionAutomaticTaxParams{
			Enabled: stripe.Bool(true),
		},
	}

	s, err := session.New(params)
	if err != nil {
		log.Printf("❌ session.New error: %v", err)
		writeJSON(w, map[string]string{"error": "Failed to create Stripe session"}, http.StatusInternalServerError)
		return
	}

	writeJSON(w, map[string]string{
		"clientSecret": s.ClientSecret,
	}, http.StatusOK)
}

func RetrieveCheckoutSession(w http.ResponseWriter, r *http.Request) {
	sessionID := r.URL.Query().Get("session_id")
	s, err := session.Get(sessionID, nil)
	if err != nil {
		log.Printf("❌ session.Get error: %v", err)
		writeJSON(w, map[string]string{"error": "Session not found"}, http.StatusBadRequest)
		return
	}

	writeJSON(w, map[string]string{
		"status":         string(s.Status),
		"customer_email": s.CustomerDetails.Email,
	}, http.StatusOK)
}

func writeJSON(w http.ResponseWriter, v interface{}, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("❌ JSON encoding error: %v", err)
	}
}

func debugPrices() {
	params := &stripe.PriceListParams{}
	i := price.List(params)
	for i.Next() {
		p := i.Price()
		log.Printf("🔍 Preis-ID: %s, Produkt: %v", p.ID, p.Product)
	}
}
