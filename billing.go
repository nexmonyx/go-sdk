package nexmonyx

import (
	"context"
	"fmt"
)

// GetBillingInfo retrieves billing information for an organization
func (s *BillingService) GetBillingInfo(ctx context.Context, organizationID string) (*BillingInfo, error) {
	var resp StandardResponse
	resp.Data = &BillingInfo{}

	_, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   fmt.Sprintf("/api/v1/organizations/%s/billing", organizationID),
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if billing, ok := resp.Data.(*BillingInfo); ok {
		return billing, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// GetSubscription retrieves subscription details
func (s *BillingService) GetSubscription(ctx context.Context, organizationID string) (*Subscription, error) {
	var resp StandardResponse
	resp.Data = &Subscription{}

	_, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   fmt.Sprintf("/api/v1/organizations/%s/subscription", organizationID),
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if sub, ok := resp.Data.(*Subscription); ok {
		return sub, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// ListInvoices retrieves invoices for an organization
func (s *BillingService) ListInvoices(ctx context.Context, organizationID string, opts *ListOptions) ([]*Invoice, *PaginationMeta, error) {
	var resp PaginatedResponse
	var invoices []*Invoice
	resp.Data = &invoices

	req := &Request{
		Method: "GET",
		Path:   fmt.Sprintf("/api/v1/organizations/%s/invoices", organizationID),
		Result: &resp,
	}

	if opts != nil {
		req.Query = opts.ToQuery()
	}

	_, err := s.client.Do(ctx, req)
	if err != nil {
		return nil, nil, err
	}

	return invoices, resp.Meta, nil
}

// UpdatePaymentMethod updates the payment method for an organization
func (s *BillingService) UpdatePaymentMethod(ctx context.Context, organizationID string, paymentMethod *PaymentMethod) error {
	var resp StandardResponse

	_, err := s.client.Do(ctx, &Request{
		Method: "PUT",
		Path:   fmt.Sprintf("/api/v1/organizations/%s/payment-method", organizationID),
		Body:   paymentMethod,
		Result: &resp,
	})
	return err
}

// BillingInfo represents billing information
type BillingInfo struct {
	OrganizationID   uint                   `json:"organization_id"`
	StripeCustomerID string                 `json:"stripe_customer_id"`
	CurrentBalance   float64                `json:"current_balance"`
	Credits          float64                `json:"credits"`
	PaymentMethod    *PaymentMethod         `json:"payment_method,omitempty"`
	BillingAddress   *BillingAddress        `json:"billing_address,omitempty"`
	TaxInfo          *TaxInfo               `json:"tax_info,omitempty"`
	NextBillingDate  *CustomTime            `json:"next_billing_date,omitempty"`
	BillingCycle     string                 `json:"billing_cycle"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
}

// Subscription represents a subscription
type Subscription struct {
	ID                 string                 `json:"id"`
	OrganizationID     uint                   `json:"organization_id"`
	PlanID             string                 `json:"plan_id"`
	PlanName           string                 `json:"plan_name"`
	Status             string                 `json:"status"`
	CurrentPeriodStart *CustomTime            `json:"current_period_start"`
	CurrentPeriodEnd   *CustomTime            `json:"current_period_end"`
	TrialStart         *CustomTime            `json:"trial_start,omitempty"`
	TrialEnd           *CustomTime            `json:"trial_end,omitempty"`
	CancelAtPeriodEnd  bool                   `json:"cancel_at_period_end"`
	CanceledAt         *CustomTime            `json:"canceled_at,omitempty"`
	Quantity           int                    `json:"quantity"`
	AddOns             []SubscriptionAddOn    `json:"add_ons,omitempty"`
	Metadata           map[string]interface{} `json:"metadata,omitempty"`
}

// Invoice represents an invoice
type Invoice struct {
	ID             string                 `json:"id"`
	OrganizationID uint                   `json:"organization_id"`
	InvoiceNumber  string                 `json:"invoice_number"`
	Status         string                 `json:"status"`
	Amount         float64                `json:"amount"`
	Currency       string                 `json:"currency"`
	DueDate        *CustomTime            `json:"due_date"`
	PaidAt         *CustomTime            `json:"paid_at,omitempty"`
	PeriodStart    *CustomTime            `json:"period_start"`
	PeriodEnd      *CustomTime            `json:"period_end"`
	LineItems      []InvoiceLineItem      `json:"line_items"`
	PDFURL         string                 `json:"pdf_url,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// PaymentMethod represents a payment method
type PaymentMethod struct {
	ID          string      `json:"id"`
	Type        string      `json:"type"` // card, bank_account
	Last4       string      `json:"last4"`
	Brand       string      `json:"brand,omitempty"` // For cards
	ExpiryMonth int         `json:"expiry_month,omitempty"`
	ExpiryYear  int         `json:"expiry_year,omitempty"`
	BankName    string      `json:"bank_name,omitempty"` // For bank accounts
	IsDefault   bool        `json:"is_default"`
	CreatedAt   *CustomTime `json:"created_at"`
}

// BillingAddress represents a billing address
type BillingAddress struct {
	Line1      string `json:"line1"`
	Line2      string `json:"line2,omitempty"`
	City       string `json:"city"`
	State      string `json:"state"`
	PostalCode string `json:"postal_code"`
	Country    string `json:"country"`
}

// TaxInfo represents tax information
type TaxInfo struct {
	TaxID           string `json:"tax_id"`
	TaxType         string `json:"tax_type"` // vat, gst, etc.
	CompanyName     string `json:"company_name"`
	TaxExempt       bool   `json:"tax_exempt"`
	TaxExemptReason string `json:"tax_exempt_reason,omitempty"`
}

// SubscriptionAddOn represents an add-on to a subscription
type SubscriptionAddOn struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Quantity int     `json:"quantity"`
	Price    float64 `json:"price"`
}

// InvoiceLineItem represents a line item on an invoice
type InvoiceLineItem struct {
	Description string  `json:"description"`
	Quantity    int     `json:"quantity"`
	UnitPrice   float64 `json:"unit_price"`
	Amount      float64 `json:"amount"`
	Type        string  `json:"type"` // subscription, usage, add_on
}
