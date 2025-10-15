package nexmonyx

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Error path tests to improve coverage from 87.5% to 100%
// Note: These tests actually won't reach the type assertion code since JSON unmarshaling
// fails first. The type assertion line (return nil, fmt.Errorf("unexpected response type"))
// appears to be unreachable in practice since JSON unmarshaling validates types first.

// Additional edge case tests

func TestBillingService_ListInvoices_NilOptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Equal(t, "/v1/organizations/org-123/invoices", r.URL.Path)

		// Verify no query parameters when opts is nil
		assert.Empty(t, r.URL.Query().Get("page"))
		assert.Empty(t, r.URL.Query().Get("limit"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"data":   []*Invoice{},
			"meta": PaginationMeta{
				Page:       1,
				Limit:      25,
				TotalItems: 0,
				TotalPages: 0,
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	invoices, meta, err := client.Billing.ListInvoices(context.Background(), "org-123", nil)
	assert.NoError(t, err)
	assert.NotNil(t, invoices)
	assert.NotNil(t, meta)
	assert.Len(t, invoices, 0)
}

func TestBillingService_UpdatePaymentMethod_EmptyResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		assert.Equal(t, "/v1/organizations/org-123/payment-method", r.URL.Path)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "success",
			"message": "Payment method updated",
		})
	}))
	defer server.Close()

	client, _ := NewClient(&Config{BaseURL: server.URL})
	err := client.Billing.UpdatePaymentMethod(context.Background(), "org-123", &PaymentMethod{
		ID:   "pm_123",
		Type: "card",
		Last4: "4242",
	})
	assert.NoError(t, err)
}
