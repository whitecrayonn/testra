package billing

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	apihttp "github.com/testra/testra/apps/api/internal/shared/http"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) GetSubscription(w http.ResponseWriter, r *http.Request) {
	orgID, err := uuid.Parse(r.URL.Query().Get("organization_id"))
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "organization_id required")
		return
	}
	sub, err := h.service.GetSubscription(r.Context(), orgID)
	if err != nil {
		apihttp.MapError(w, err)
		return
	}
	apihttp.JSON(w, http.StatusOK, toSubscriptionResponse(sub))
}

func (h *Handler) UpdateSubscription(w http.ResponseWriter, r *http.Request) {
	var req struct {
		OrganizationID    string `json:"organization_id"`
		Plan              string `json:"plan"`
		Seats             int    `json:"seats"`
		CancelAtPeriodEnd *bool  `json:"cancel_at_period_end"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", err.Error())
		return
	}
	orgID, err := uuid.Parse(req.OrganizationID)
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "invalid organization_id")
		return
	}
	sub, err := h.service.UpdateSubscription(r.Context(), UpdateSubscriptionInput{
		OrganizationID:    orgID,
		Plan:              req.Plan,
		Seats:             req.Seats,
		CancelAtPeriodEnd: req.CancelAtPeriodEnd,
	})
	if err != nil {
		apihttp.MapError(w, err)
		return
	}
	apihttp.JSON(w, http.StatusOK, toSubscriptionResponse(sub))
}

func (h *Handler) ListInvoices(w http.ResponseWriter, r *http.Request) {
	orgID, err := uuid.Parse(r.URL.Query().Get("organization_id"))
	if err != nil {
		apihttp.ErrorJSON(w, http.StatusBadRequest, "INVALID_INPUT", "organization_id required")
		return
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	invoices, err := h.service.ListInvoices(r.Context(), orgID, limit)
	if err != nil {
		apihttp.MapError(w, err)
		return
	}
	apihttp.JSON(w, http.StatusOK, toInvoiceResponses(invoices))
}

// response helpers

type subscriptionResponse struct {
	ID                     string  `json:"id"`
	OrganizationID         string  `json:"organization_id"`
	ProviderSubscriptionID string  `json:"provider_subscription_id,omitempty"`
	Plan                   string  `json:"plan"`
	Status                 string  `json:"status"`
	Seats                  int     `json:"seats"`
	CurrentPeriodStart     *string `json:"current_period_start,omitempty"`
	CurrentPeriodEnd       *string `json:"current_period_end,omitempty"`
	CancelAtPeriodEnd      bool    `json:"cancel_at_period_end"`
	CreatedAt              string  `json:"created_at"`
	UpdatedAt              string  `json:"updated_at"`
}

func toSubscriptionResponse(s *Subscription) subscriptionResponse {
	resp := subscriptionResponse{
		ID:                     s.ID.String(),
		OrganizationID:         s.OrganizationID.String(),
		ProviderSubscriptionID: s.ProviderSubscriptionID,
		Plan:                   s.Plan,
		Status:                 s.Status,
		Seats:                  s.Seats,
		CancelAtPeriodEnd:      s.CancelAtPeriodEnd,
		CreatedAt:              s.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:              s.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
	if s.CurrentPeriodStart != nil {
		t := s.CurrentPeriodStart.Format("2006-01-02T15:04:05Z")
		resp.CurrentPeriodStart = &t
	}
	if s.CurrentPeriodEnd != nil {
		t := s.CurrentPeriodEnd.Format("2006-01-02T15:04:05Z")
		resp.CurrentPeriodEnd = &t
	}
	return resp
}

type invoiceResponse struct {
	ID                string  `json:"id"`
	OrganizationID    string  `json:"organization_id"`
	ProviderInvoiceID string  `json:"provider_invoice_id,omitempty"`
	AmountCents       int     `json:"amount_cents"`
	Currency          string  `json:"currency"`
	Status            string  `json:"status"`
	PeriodStart       *string `json:"period_start,omitempty"`
	PeriodEnd         *string `json:"period_end,omitempty"`
	CreatedAt         string  `json:"created_at"`
}

func toInvoiceResponse(inv *Invoice) invoiceResponse {
	resp := invoiceResponse{
		ID:                inv.ID.String(),
		OrganizationID:    inv.OrganizationID.String(),
		ProviderInvoiceID: inv.ProviderInvoiceID,
		AmountCents:       inv.AmountCents,
		Currency:          inv.Currency,
		Status:            inv.Status,
		CreatedAt:         inv.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
	if inv.PeriodStart != nil {
		t := inv.PeriodStart.Format("2006-01-02T15:04:05Z")
		resp.PeriodStart = &t
	}
	if inv.PeriodEnd != nil {
		t := inv.PeriodEnd.Format("2006-01-02T15:04:05Z")
		resp.PeriodEnd = &t
	}
	return resp
}

func toInvoiceResponses(invoices []Invoice) []invoiceResponse {
	resp := make([]invoiceResponse, len(invoices))
	for i := range invoices {
		resp[i] = toInvoiceResponse(&invoices[i])
	}
	return resp
}
