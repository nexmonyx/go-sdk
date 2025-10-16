package nexmonyx

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestBillingService_GetBillingInfoComprehensive tests the GetBillingInfo method with various scenarios
func TestBillingService_GetBillingInfoComprehensive(t *testing.T) {
	tests := []struct {
		name           string
		organizationID string
		mockStatus     int
		mockBody       interface{}
		wantErr        bool
		checkFunc      func(*testing.T, *BillingInfo)
	}{
		{
			name:           "success - full billing info",
			organizationID: "org-123",
			mockStatus:     http.StatusOK,
			mockBody: map[string]interface{}{
				"status": "success",
				"data": map[string]interface{}{
					"organization_id":     123,
					"stripe_customer_id":  "cus_123",
					"current_balance":     150.50,
					"credits":             25.00,
					"billing_cycle":       "monthly",
					"next_billing_date":   "2024-02-01T00:00:00Z",
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, info *BillingInfo) {
				assert.Equal(t, uint(123), info.OrganizationID)
				assert.Equal(t, "cus_123", info.StripeCustomerID)
				assert.Equal(t, 150.50, info.CurrentBalance)
				assert.Equal(t, 25.00, info.Credits)
			},
		},
		{
			name:           "success - minimal billing info",
			organizationID: "org-456",
			mockStatus:     http.StatusOK,
			mockBody: map[string]interface{}{
				"status": "success",
				"data": map[string]interface{}{
					"organization_id": 456,
					"current_balance": 0.00,
					"credits":         0.00,
					"billing_cycle":   "monthly",
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, info *BillingInfo) {
				assert.Equal(t, uint(456), info.OrganizationID)
			},
		},
		{
			name:           "not found",
			organizationID: "org-999",
			mockStatus:     http.StatusNotFound,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Organization not found",
			},
			wantErr: true,
		},
		{
			name:           "unauthorized",
			organizationID: "org-123",
			mockStatus:     http.StatusUnauthorized,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Authentication required",
			},
			wantErr: true,
		},
		{
			name:           "forbidden",
			organizationID: "org-123",
			mockStatus:     http.StatusForbidden,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Access denied to billing information",
			},
			wantErr: true,
		},
		{
			name:           "server error",
			organizationID: "org-123",
			mockStatus:     http.StatusInternalServerError,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Internal server error",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "GET", r.Method)
				assert.Contains(t, r.URL.Path, tt.organizationID)
				assert.Contains(t, r.URL.Path, "/billing")

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatus)
				json.NewEncoder(w).Encode(tt.mockBody)
			}))
			defer server.Close()

			client, err := NewClient(&Config{
				BaseURL:    server.URL,
				Auth:       AuthConfig{Token: "test-token"},
				RetryCount: 0,
			})
			require.NoError(t, err)

			ctx := context.Background()
			if tt.wantErr && tt.mockStatus >= 500 {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(context.Background(), 2*time.Second)
				defer cancel()
			}

			result, err := client.Billing.GetBillingInfo(ctx, tt.organizationID)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				if tt.checkFunc != nil {
					tt.checkFunc(t, result)
				}
			}
		})
	}
}

// TestBillingService_GetSubscriptionComprehensive tests the GetSubscription method
func TestBillingService_GetSubscriptionComprehensive(t *testing.T) {
	tests := []struct {
		name           string
		organizationID string
		mockStatus     int
		mockBody       interface{}
		wantErr        bool
		checkFunc      func(*testing.T, *Subscription)
	}{
		{
			name:           "success - active subscription",
			organizationID: "org-123",
			mockStatus:     http.StatusOK,
			mockBody: map[string]interface{}{
				"status": "success",
				"data": map[string]interface{}{
					"id":                   "sub_123",
					"organization_id":      123,
					"plan_id":              "plan_premium",
					"plan_name":            "Premium Plan",
					"status":               "active",
					"current_period_start": "2024-01-01T00:00:00Z",
					"current_period_end":   "2024-02-01T00:00:00Z",
					"quantity":             1,
					"cancel_at_period_end": false,
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, sub *Subscription) {
				assert.Equal(t, "sub_123", sub.ID)
				assert.Equal(t, uint(123), sub.OrganizationID)
				assert.Equal(t, "Premium Plan", sub.PlanName)
				assert.Equal(t, "active", sub.Status)
			},
		},
		{
			name:           "success - trial subscription",
			organizationID: "org-456",
			mockStatus:     http.StatusOK,
			mockBody: map[string]interface{}{
				"status": "success",
				"data": map[string]interface{}{
					"id":              "sub_456",
					"organization_id": 456,
					"plan_id":         "plan_trial",
					"plan_name":       "Trial Plan",
					"status":          "trialing",
					"trial_start":     "2024-01-01T00:00:00Z",
					"trial_end":       "2024-01-15T00:00:00Z",
					"quantity":        1,
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, sub *Subscription) {
				assert.Equal(t, "trialing", sub.Status)
			},
		},
		{
			name:           "success - canceled subscription",
			organizationID: "org-789",
			mockStatus:     http.StatusOK,
			mockBody: map[string]interface{}{
				"status": "success",
				"data": map[string]interface{}{
					"id":                   "sub_789",
					"organization_id":      789,
					"plan_id":              "plan_basic",
					"plan_name":            "Basic Plan",
					"status":               "active",
					"cancel_at_period_end": true,
					"canceled_at":          "2024-01-15T00:00:00Z",
					"quantity":             1,
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, sub *Subscription) {
				assert.True(t, sub.CancelAtPeriodEnd)
			},
		},
		{
			name:           "not found - no subscription",
			organizationID: "org-999",
			mockStatus:     http.StatusNotFound,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "No active subscription",
			},
			wantErr: true,
		},
		{
			name:           "unauthorized",
			organizationID: "org-123",
			mockStatus:     http.StatusUnauthorized,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Authentication required",
			},
			wantErr: true,
		},
		{
			name:           "forbidden",
			organizationID: "org-123",
			mockStatus:     http.StatusForbidden,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Access denied",
			},
			wantErr: true,
		},
		{
			name:           "server error",
			organizationID: "org-123",
			mockStatus:     http.StatusInternalServerError,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Internal server error",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "GET", r.Method)
				assert.Contains(t, r.URL.Path, tt.organizationID)
				assert.Contains(t, r.URL.Path, "/subscription")

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatus)
				json.NewEncoder(w).Encode(tt.mockBody)
			}))
			defer server.Close()

			client, err := NewClient(&Config{
				BaseURL:    server.URL,
				Auth:       AuthConfig{Token: "test-token"},
				RetryCount: 0,
			})
			require.NoError(t, err)

			ctx := context.Background()
			if tt.wantErr && tt.mockStatus >= 500 {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(context.Background(), 2*time.Second)
				defer cancel()
			}

			result, err := client.Billing.GetSubscription(ctx, tt.organizationID)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				if tt.checkFunc != nil {
					tt.checkFunc(t, result)
				}
			}
		})
	}
}

// TestBillingService_ListInvoicesComprehensive tests the ListInvoices method
func TestBillingService_ListInvoicesComprehensive(t *testing.T) {
	tests := []struct {
		name           string
		organizationID string
		opts           *ListOptions
		mockStatus     int
		mockBody       interface{}
		wantErr        bool
		checkFunc      func(*testing.T, []*Invoice, *PaginationMeta)
	}{
		{
			name:           "success - with invoices",
			organizationID: "org-123",
			opts: &ListOptions{
				Page:  1,
				Limit: 10,
			},
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"status": "success",
				"data": []map[string]interface{}{
					{
						"id":              "inv_1",
						"organization_id": 123,
						"invoice_number":  "INV-001",
						"status":          "paid",
						"amount":          99.99,
						"currency":        "USD",
						"due_date":        "2024-01-31T00:00:00Z",
						"paid_at":         "2024-01-25T00:00:00Z",
					},
					{
						"id":              "inv_2",
						"organization_id": 123,
						"invoice_number":  "INV-002",
						"status":          "open",
						"amount":          149.99,
						"currency":        "USD",
						"due_date":        "2024-02-28T00:00:00Z",
					},
				},
				"meta": map[string]interface{}{
					"page":        1,
					"limit":       10,
					"total_items": 2,
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, invoices []*Invoice, meta *PaginationMeta) {
				assert.Len(t, invoices, 2)
				assert.Equal(t, "INV-001", invoices[0].InvoiceNumber)
				assert.Equal(t, "paid", invoices[0].Status)
				assert.NotNil(t, meta)
			},
		},
		{
			name:           "success - nil options",
			organizationID: "org-123",
			opts:           nil,
			mockStatus:     http.StatusOK,
			mockBody: map[string]interface{}{
				"status": "success",
				"data":   []map[string]interface{}{},
				"meta": map[string]interface{}{
					"total_items": 0,
				},
			},
			wantErr: false,
		},
		{
			name:           "success - empty result",
			organizationID: "org-456",
			opts:           &ListOptions{Page: 1},
			mockStatus:     http.StatusOK,
			mockBody: map[string]interface{}{
				"status": "success",
				"data":   []map[string]interface{}{},
				"meta": map[string]interface{}{
					"total_items": 0,
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, invoices []*Invoice, meta *PaginationMeta) {
				assert.Len(t, invoices, 0)
			},
		},
		{
			name:           "not found",
			organizationID: "org-999",
			opts:           &ListOptions{Page: 1},
			mockStatus:     http.StatusNotFound,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Organization not found",
			},
			wantErr: true,
		},
		{
			name:           "unauthorized",
			organizationID: "org-123",
			opts:           &ListOptions{Page: 1},
			mockStatus:     http.StatusUnauthorized,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Authentication required",
			},
			wantErr: true,
		},
		{
			name:           "forbidden",
			organizationID: "org-123",
			opts:           &ListOptions{Page: 1},
			mockStatus:     http.StatusForbidden,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Access denied",
			},
			wantErr: true,
		},
		{
			name:           "server error",
			organizationID: "org-123",
			opts:           &ListOptions{Page: 1},
			mockStatus:     http.StatusInternalServerError,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Internal server error",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "GET", r.Method)
				assert.Contains(t, r.URL.Path, tt.organizationID)
				assert.Contains(t, r.URL.Path, "/invoices")

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatus)
				json.NewEncoder(w).Encode(tt.mockBody)
			}))
			defer server.Close()

			client, err := NewClient(&Config{
				BaseURL:    server.URL,
				Auth:       AuthConfig{Token: "test-token"},
				RetryCount: 0,
			})
			require.NoError(t, err)

			ctx := context.Background()
			if tt.wantErr && tt.mockStatus >= 500 {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(context.Background(), 2*time.Second)
				defer cancel()
			}

			invoices, meta, err := client.Billing.ListInvoices(ctx, tt.organizationID, tt.opts)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, invoices)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, invoices)
				if tt.checkFunc != nil {
					tt.checkFunc(t, invoices, meta)
				}
			}
		})
	}
}

// TestBillingService_UpdatePaymentMethodComprehensive tests the UpdatePaymentMethod method
func TestBillingService_UpdatePaymentMethodComprehensive(t *testing.T) {
	tests := []struct {
		name           string
		organizationID string
		paymentMethod  *PaymentMethod
		mockStatus     int
		mockBody       interface{}
		wantErr        bool
	}{
		{
			name:           "success - update card",
			organizationID: "org-123",
			paymentMethod: &PaymentMethod{
				Type:        "card",
				Last4:       "4242",
				Brand:       "visa",
				ExpiryMonth: 12,
				ExpiryYear:  2025,
				IsDefault:   true,
			},
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"status":  "success",
				"message": "Payment method updated",
			},
			wantErr: false,
		},
		{
			name:           "success - update bank account",
			organizationID: "org-456",
			paymentMethod: &PaymentMethod{
				Type:      "bank_account",
				Last4:     "6789",
				BankName:  "Chase",
				IsDefault: true,
			},
			mockStatus: http.StatusOK,
			mockBody: map[string]interface{}{
				"status":  "success",
				"message": "Payment method updated",
			},
			wantErr: false,
		},
		{
			name:           "validation error - invalid card",
			organizationID: "org-123",
			paymentMethod: &PaymentMethod{
				Type:  "card",
				Last4: "0000",
			},
			mockStatus: http.StatusBadRequest,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Invalid payment method",
			},
			wantErr: true,
		},
		{
			name:           "not found",
			organizationID: "org-999",
			paymentMethod: &PaymentMethod{
				Type:  "card",
				Last4: "4242",
			},
			mockStatus: http.StatusNotFound,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Organization not found",
			},
			wantErr: true,
		},
		{
			name:           "unauthorized",
			organizationID: "org-123",
			paymentMethod: &PaymentMethod{
				Type:  "card",
				Last4: "4242",
			},
			mockStatus: http.StatusUnauthorized,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Authentication required",
			},
			wantErr: true,
		},
		{
			name:           "forbidden",
			organizationID: "org-123",
			paymentMethod: &PaymentMethod{
				Type:  "card",
				Last4: "4242",
			},
			mockStatus: http.StatusForbidden,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Insufficient permissions",
			},
			wantErr: true,
		},
		{
			name:           "server error",
			organizationID: "org-123",
			paymentMethod: &PaymentMethod{
				Type:  "card",
				Last4: "4242",
			},
			mockStatus: http.StatusInternalServerError,
			mockBody: map[string]interface{}{
				"status":  "error",
				"message": "Payment processing error",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "PUT", r.Method)
				assert.Contains(t, r.URL.Path, tt.organizationID)
				assert.Contains(t, r.URL.Path, "/payment-method")

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatus)
				json.NewEncoder(w).Encode(tt.mockBody)
			}))
			defer server.Close()

			client, err := NewClient(&Config{
				BaseURL:    server.URL,
				Auth:       AuthConfig{Token: "test-token"},
				RetryCount: 0,
			})
			require.NoError(t, err)

			ctx := context.Background()
			if tt.wantErr && tt.mockStatus >= 500 {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(context.Background(), 2*time.Second)
				defer cancel()
			}

			err = client.Billing.UpdatePaymentMethod(ctx, tt.organizationID, tt.paymentMethod)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
