package billing

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

// NewPaymentProvider returns a Stripe-backed provider when a secret key is
// configured; otherwise it returns nil and the service uses local billing data.
func NewPaymentProvider(stripeSecretKey string) PaymentProvider {
	if stripeSecretKey == "" {
		return nil
	}
	return &stripeProvider{
		secretKey: stripeSecretKey,
		client:    &http.Client{Timeout: 10 * time.Second},
	}
}

type stripeProvider struct {
	secretKey string
	client    *http.Client
}

func (p *stripeProvider) GetSubscription(ctx context.Context, providerID string) (*Subscription, error) {
	if providerID == "" {
		return nil, fmt.Errorf("provider subscription id required")
	}
	respBody, err := p.request(ctx, http.MethodGet, "https://api.stripe.com/v1/subscriptions/"+providerID, nil)
	if err != nil {
		return nil, err
	}
	return parseStripeSubscription(uuid.Nil, respBody)
}

func (p *stripeProvider) UpdateSubscription(ctx context.Context, sub *Subscription) (*Subscription, error) {
	if sub.ProviderSubscriptionID == "" {
		return nil, fmt.Errorf("provider subscription id required")
	}
	form := url.Values{}
	form.Set("cancel_at_period_end", strconv.FormatBool(sub.CancelAtPeriodEnd))
	if sub.Plan != "" {
		form.Set("metadata[plan]", sub.Plan)
	}
	if sub.Seats > 0 {
		form.Set("metadata[seats]", strconv.Itoa(sub.Seats))
	}
	respBody, err := p.request(ctx, http.MethodPost, "https://api.stripe.com/v1/subscriptions/"+sub.ProviderSubscriptionID, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	return parseStripeSubscription(sub.OrganizationID, respBody)
}

func (p *stripeProvider) ListInvoices(ctx context.Context, providerSubscriptionID string) ([]Invoice, error) {
	u := "https://api.stripe.com/v1/invoices?limit=20"
	if providerSubscriptionID != "" {
		u += "&subscription=" + url.QueryEscape(providerSubscriptionID)
	}
	respBody, err := p.request(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	return parseStripeInvoices(respBody)
}

func (p *stripeProvider) request(ctx context.Context, method, urlStr string, body io.Reader) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, method, urlStr, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+p.secretKey)
	if method == http.MethodPost {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("stripe returned %d: %s", resp.StatusCode, string(respBody))
	}
	return respBody, nil
}

func parseStripeSubscription(orgID uuid.UUID, body []byte) (*Subscription, error) {
	var raw struct {
		ID                   string `json:"id"`
		Status               string `json:"status"`
		CurrentPeriodStart   int64  `json:"current_period_start"`
		CurrentPeriodEnd     int64  `json:"current_period_end"`
		CancelAtPeriodEnd    bool   `json:"cancel_at_period_end"`
		Items                struct {
			Data []struct {
				Quantity int `json:"quantity"`
				Price    struct {
					Nickname string `json:"nickname"`
				} `json:"price"`
			} `json:"data"`
		} `json:"items"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, err
	}
	if raw.ID == "" {
		return nil, fmt.Errorf("invalid stripe subscription response")
	}

	var seats int
	if len(raw.Items.Data) > 0 {
		seats = raw.Items.Data[0].Quantity
	}
	start := time.Unix(raw.CurrentPeriodStart, 0).UTC()
	end := time.Unix(raw.CurrentPeriodEnd, 0).UTC()

	s := &Subscription{
		ProviderSubscriptionID: raw.ID,
		Status:                 raw.Status,
		Seats:                  seats,
		CurrentPeriodStart:     &start,
		CurrentPeriodEnd:       &end,
		CancelAtPeriodEnd:      raw.CancelAtPeriodEnd,
	}
	if orgID != uuid.Nil {
		s.OrganizationID = orgID
	}
	return s, nil
}

func parseStripeInvoices(body []byte) ([]Invoice, error) {
	var raw struct {
		Data []struct {
			ID        string `json:"id"`
			AmountDue int    `json:"amount_due"`
			Currency  string `json:"currency"`
			Status    string `json:"status"`
			PeriodStart int64 `json:"period_start"`
			PeriodEnd   int64 `json:"period_end"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, err
	}

	invoices := make([]Invoice, len(raw.Data))
	for i, item := range raw.Data {
		start := time.Unix(item.PeriodStart, 0).UTC()
		end := time.Unix(item.PeriodEnd, 0).UTC()
		invoices[i] = Invoice{
			ProviderInvoiceID: item.ID,
			AmountCents:       item.AmountDue,
			Currency:          strings.ToUpper(item.Currency),
			Status:            item.Status,
			PeriodStart:       &start,
			PeriodEnd:         &end,
		}
	}
	return invoices, nil
}
