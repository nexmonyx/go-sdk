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
		Path:   fmt.Sprintf("/v1/organizations/%s/billing", organizationID),
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
		Path:   fmt.Sprintf("/v1/organizations/%s/subscription", organizationID),
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
		Path:   fmt.Sprintf("/v1/organizations/%s/invoices", organizationID),
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
		Path:   fmt.Sprintf("/v1/organizations/%s/payment-method", organizationID),
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

// ============================================================================
// Self-Service Subscription Methods (Task #3939)
// ============================================================================

// GetMySubscription retrieves the subscription for the authenticated user's organization
// Authentication: JWT Token required
// Endpoint: GET /v1/subscription
func (s *BillingService) GetMySubscription(ctx context.Context) (*SubscriptionResponse, error) {
	var resp StandardResponse
	resp.Data = &SubscriptionResponse{}

	_, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   "/v1/subscription",
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if sub, ok := resp.Data.(*SubscriptionResponse); ok {
		return sub, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// CreateCheckout creates a Stripe checkout session for subscribing to a plan
// Authentication: JWT Token required
// Endpoint: POST /v1/subscription/checkout
func (s *BillingService) CreateCheckout(ctx context.Context, req *CreateCheckoutRequest) (*CheckoutSessionResponse, error) {
	var resp StandardResponse
	resp.Data = &CheckoutSessionResponse{}

	_, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   "/v1/subscription/checkout",
		Body:   req,
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if checkout, ok := resp.Data.(*CheckoutSessionResponse); ok {
		return checkout, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// UpdateMySubscription updates the subscription for the authenticated user's organization
// Authentication: JWT Token required
// Endpoint: PUT /v1/subscription
func (s *BillingService) UpdateMySubscription(ctx context.Context, req *UpdateSubscriptionRequest) (*SubscriptionResponse, error) {
	var resp StandardResponse
	resp.Data = &SubscriptionResponse{}

	_, err := s.client.Do(ctx, &Request{
		Method: "PUT",
		Path:   "/v1/subscription",
		Body:   req,
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if sub, ok := resp.Data.(*SubscriptionResponse); ok {
		return sub, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// CancelMySubscription cancels the subscription for the authenticated user's organization
// Authentication: JWT Token required
// Endpoint: DELETE /v1/subscription
func (s *BillingService) CancelMySubscription(ctx context.Context, req *CancelSubscriptionRequest) error {
	var resp StandardResponse

	_, err := s.client.Do(ctx, &Request{
		Method: "DELETE",
		Path:   "/v1/subscription",
		Body:   req,
		Result: &resp,
	})
	return err
}

// CreatePortalSession creates a Stripe customer portal session
// Authentication: JWT Token required
// Endpoint: POST /v1/billing/portal
func (s *BillingService) CreatePortalSession(ctx context.Context, returnURL string) (*PortalSessionResponse, error) {
	var resp StandardResponse
	resp.Data = &PortalSessionResponse{}

	_, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   "/v1/billing/portal",
		Body: map[string]string{
			"return_url": returnURL,
		},
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if portal, ok := resp.Data.(*PortalSessionResponse); ok {
		return portal, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// ============================================================================
// Plan Methods (Task #3939)
// ============================================================================

// ListPlans retrieves all available subscription plans
// Authentication: None (public endpoint)
// Endpoint: GET /v1/pricing/plans
func (s *BillingService) ListPlans(ctx context.Context) ([]*Plan, error) {
	var resp StandardResponse
	var plans []*Plan
	resp.Data = &plans

	_, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   "/v1/pricing/plans",
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	return plans, nil
}

// GetPlan retrieves a specific subscription plan by ID
// Authentication: None (public endpoint)
// Endpoint: GET /v1/pricing/plans/:plan_id
func (s *BillingService) GetPlan(ctx context.Context, planID string) (*Plan, error) {
	var resp StandardResponse
	resp.Data = &Plan{}

	_, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   fmt.Sprintf("/v1/pricing/plans/%s", planID),
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if plan, ok := resp.Data.(*Plan); ok {
		return plan, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// GetPlanFeatures retrieves the feature comparison matrix for all plans
// Authentication: None (public endpoint)
// Endpoint: GET /v1/pricing/features
func (s *BillingService) GetPlanFeatures(ctx context.Context) (*FeatureMatrix, error) {
	var resp StandardResponse
	resp.Data = &FeatureMatrix{}

	_, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   "/v1/pricing/features",
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if matrix, ok := resp.Data.(*FeatureMatrix); ok {
		return matrix, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// ============================================================================
// Invoice Methods (Task #3939)
// ============================================================================

// ListMyInvoices retrieves invoices for the authenticated user's organization
// Authentication: JWT Token required
// Endpoint: GET /v1/billing/invoices
func (s *BillingService) ListMyInvoices(ctx context.Context, opts *ListInvoiceOptions) ([]*Invoice, *PaginationMeta, error) {
	var resp PaginatedResponse
	var invoices []*Invoice
	resp.Data = &invoices

	req := &Request{
		Method: "GET",
		Path:   "/v1/billing/invoices",
		Result: &resp,
	}

	if opts != nil {
		query := make(map[string]string)
		if opts.Page > 0 {
			query["page"] = fmt.Sprintf("%d", opts.Page)
		}
		if opts.Limit > 0 {
			query["limit"] = fmt.Sprintf("%d", opts.Limit)
		}
		if opts.Status != "" {
			query["status"] = opts.Status
		}
		req.Query = query
	}

	_, err := s.client.Do(ctx, req)
	if err != nil {
		return nil, nil, err
	}

	return invoices, resp.Meta, nil
}

// GetMyInvoice retrieves a specific invoice by ID
// Authentication: JWT Token required
// Endpoint: GET /v1/billing/invoices/:invoice_id
func (s *BillingService) GetMyInvoice(ctx context.Context, invoiceID string) (*Invoice, error) {
	var resp StandardResponse
	resp.Data = &Invoice{}

	_, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   fmt.Sprintf("/v1/billing/invoices/%s", invoiceID),
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if invoice, ok := resp.Data.(*Invoice); ok {
		return invoice, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// DownloadInvoicePDF downloads an invoice as PDF
// Authentication: JWT Token required
// Endpoint: GET /v1/billing/invoices/:invoice_id/download
func (s *BillingService) DownloadInvoicePDF(ctx context.Context, invoiceID string) ([]byte, error) {
	resp, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   fmt.Sprintf("/v1/billing/invoices/%s/download", invoiceID),
	})
	if err != nil {
		return nil, err
	}

	return resp.Body, nil
}

// GetBillingHistory retrieves billing history for the authenticated user's organization
// Authentication: JWT Token required
// Endpoint: GET /v1/billing/history
func (s *BillingService) GetBillingHistory(ctx context.Context, opts *ListOptions) (*BillingHistoryResponse, error) {
	var resp StandardResponse
	resp.Data = &BillingHistoryResponse{}

	req := &Request{
		Method: "GET",
		Path:   "/v1/billing/history",
		Result: &resp,
	}

	if opts != nil {
		req.Query = opts.ToQuery()
	}

	_, err := s.client.Do(ctx, req)
	if err != nil {
		return nil, err
	}

	if history, ok := resp.Data.(*BillingHistoryResponse); ok {
		return history, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// ============================================================================
// Payment Method Methods (Task #3939)
// ============================================================================

// ListPaymentMethods retrieves all payment methods for the authenticated user's organization
// Authentication: JWT Token required
// Endpoint: GET /v1/billing/payment-methods
func (s *BillingService) ListPaymentMethods(ctx context.Context) ([]*PaymentMethod, error) {
	var resp StandardResponse
	var methods []*PaymentMethod
	resp.Data = &methods

	_, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   "/v1/billing/payment-methods",
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	return methods, nil
}

// AddPaymentMethod adds a new payment method to the authenticated user's organization
// Authentication: JWT Token required
// Endpoint: POST /v1/billing/payment-methods
func (s *BillingService) AddPaymentMethod(ctx context.Context, req *AddPaymentMethodRequest) (*PaymentMethod, error) {
	var resp StandardResponse
	resp.Data = &PaymentMethod{}

	_, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   "/v1/billing/payment-methods",
		Body:   req,
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if method, ok := resp.Data.(*PaymentMethod); ok {
		return method, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// RemovePaymentMethod removes a payment method from the authenticated user's organization
// Authentication: JWT Token required
// Endpoint: DELETE /v1/billing/payment-methods/:payment_method_id
func (s *BillingService) RemovePaymentMethod(ctx context.Context, paymentMethodID string) error {
	var resp StandardResponse

	_, err := s.client.Do(ctx, &Request{
		Method: "DELETE",
		Path:   fmt.Sprintf("/v1/billing/payment-methods/%s", paymentMethodID),
		Result: &resp,
	})
	return err
}

// SetDefaultPaymentMethod sets a payment method as the default for the authenticated user's organization
// Authentication: JWT Token required
// Endpoint: PUT /v1/billing/payment-methods/:payment_method_id/default
func (s *BillingService) SetDefaultPaymentMethod(ctx context.Context, paymentMethodID string) error {
	var resp StandardResponse

	_, err := s.client.Do(ctx, &Request{
		Method: "PUT",
		Path:   fmt.Sprintf("/v1/billing/payment-methods/%s/default", paymentMethodID),
		Result: &resp,
	})
	return err
}

// ============================================================================
// Request/Response Types (Task #3939)
// ============================================================================

// CreateCheckoutRequest represents the request body for creating a checkout session
type CreateCheckoutRequest struct {
	PlanID       string `json:"plan_id"`
	BillingCycle string `json:"billing_cycle"` // monthly, yearly
	SuccessURL   string `json:"success_url"`
	CancelURL    string `json:"cancel_url"`
}

// UpdateSubscriptionRequest represents the request body for updating a subscription
type UpdateSubscriptionRequest struct {
	PlanID       string `json:"plan_id,omitempty"`
	BillingCycle string `json:"billing_cycle,omitempty"`
}

// CancelSubscriptionRequest represents the request body for canceling a subscription
type CancelSubscriptionRequest struct {
	Reason          string `json:"reason,omitempty"`
	CancelAtPeriod  bool   `json:"cancel_at_period_end"`
	FeedbackReason  string `json:"feedback_reason,omitempty"`
}

// AddPaymentMethodRequest represents the request body for adding a payment method
type AddPaymentMethodRequest struct {
	PaymentMethodToken string `json:"payment_method_token"`
	SetDefault         bool   `json:"set_default,omitempty"`
}

// SubscriptionResponse represents the subscription response from self-service endpoints
type SubscriptionResponse struct {
	ID                   string      `json:"id"`
	OrganizationID       uint        `json:"organization_id"`
	PlanID               string      `json:"plan_id"`
	PlanName             string      `json:"plan_name"`
	Status               string      `json:"status"`
	BillingCycle         string      `json:"billing_cycle"`
	CurrentPeriodStart   *CustomTime `json:"current_period_start"`
	CurrentPeriodEnd     *CustomTime `json:"current_period_end"`
	TrialEnd             *CustomTime `json:"trial_end,omitempty"`
	CancelAtPeriodEnd    bool        `json:"cancel_at_period_end"`
	CanceledAt           *CustomTime `json:"canceled_at,omitempty"`
	StripeSubscriptionID string      `json:"stripe_subscription_id,omitempty"`
}

// CheckoutSessionResponse represents the response from creating a checkout session
type CheckoutSessionResponse struct {
	SessionID  string `json:"session_id"`
	SessionURL string `json:"session_url"`
	ExpiresAt  int64  `json:"expires_at"`
}

// PortalSessionResponse represents the response from creating a portal session
type PortalSessionResponse struct {
	URL       string `json:"url"`
	ExpiresAt int64  `json:"expires_at"`
}

// Plan represents a subscription plan
type Plan struct {
	ID           string        `json:"id"`
	Name         string        `json:"name"`
	Description  string        `json:"description"`
	MonthlyPrice int64         `json:"monthly_price"`
	YearlyPrice  int64         `json:"yearly_price"`
	Currency     string        `json:"currency"`
	Features     []PlanFeature `json:"features"`
	Limits       PlanLimits    `json:"limits"`
	IsPublic     bool          `json:"is_public"`
	SortOrder    int           `json:"sort_order"`
}

// PlanFeature represents a feature included in a plan
type PlanFeature struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Included    bool   `json:"included"`
	Limit       string `json:"limit,omitempty"`
}

// PlanLimits represents the resource limits for a plan
type PlanLimits struct {
	MaxServers        int `json:"max_servers"`
	MaxUsers          int `json:"max_users"`
	MaxProbes         int `json:"max_probes"`
	DataRetentionDays int `json:"data_retention_days"`
}

// FeatureMatrix represents the feature comparison matrix for plans
type FeatureMatrix struct {
	Features []FeatureRow `json:"features"`
	Plans    []Plan       `json:"plans"`
}

// FeatureRow represents a row in the feature comparison matrix
type FeatureRow struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Category    string            `json:"category"`
	PlanValues  map[string]string `json:"plan_values"`
}

// ListInvoiceOptions represents options for listing invoices
type ListInvoiceOptions struct {
	Page   int    `json:"page,omitempty"`
	Limit  int    `json:"limit,omitempty"`
	Status string `json:"status,omitempty"`
}

// BillingHistoryResponse represents the billing history response
type BillingHistoryResponse struct {
	Invoices   []*Invoice `json:"invoices"`
	TotalSpent int64      `json:"total_spent"`
	Currency   string     `json:"currency"`
}
