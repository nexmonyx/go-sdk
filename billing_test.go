package nexmonyx

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestBillingService_GetBillingInfo tests retrieving billing information
func TestBillingService_GetBillingInfo(t *testing.T) {
	tests := []struct {
		name           string
		organizationID string
		mockStatus     int
		mockBody       interface{}
		wantErr        bool
		checkFunc      func(*testing.T, *BillingInfo)
	}{
		{
			name:           "successful get billing info",
			organizationID: "org-123",
			mockStatus:     http.StatusOK,
			mockBody: StandardResponse{
				Status:  "success",
				Message: "Billing info retrieved",
				Data: &BillingInfo{
					OrganizationID:   1,
					StripeCustomerID: "cus_123abc",
					CurrentBalance:   -50.00,
					Credits:          100.00,
					BillingCycle:     "monthly",
					PaymentMethod: &PaymentMethod{
						ID:     "pm_123",
						Type:   "card",
						Last4:  "4242",
						Brand:  "visa",
					},
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, info *BillingInfo) {
				if info.StripeCustomerID != "cus_123abc" {
					t.Errorf("Expected StripeCustomerID 'cus_123abc', got '%s'", info.StripeCustomerID)
				}
				if info.CurrentBalance != -50.00 {
					t.Errorf("Expected CurrentBalance -50.00, got %f", info.CurrentBalance)
				}
				if info.Credits != 100.00 {
					t.Errorf("Expected Credits 100.00, got %f", info.Credits)
				}
				if info.PaymentMethod == nil {
					t.Error("Expected PaymentMethod to be present")
				} else if info.PaymentMethod.Last4 != "4242" {
					t.Errorf("Expected Last4 '4242', got '%s'", info.PaymentMethod.Last4)
				}
			},
		},
		{
			name:           "billing info not found",
			organizationID: "invalid-org",
			mockStatus:     http.StatusNotFound,
			mockBody: StandardResponse{
				Status:  "error",
				Message: "Organization not found",
			},
			wantErr: true,
		},
		{
			name:           "unauthorized access",
			organizationID: "org-123",
			mockStatus:     http.StatusUnauthorized,
			mockBody: StandardResponse{
				Status:  "error",
				Message: "Unauthorized",
			},
			wantErr: true,
		},
		{
			name:           "server error",
			organizationID: "org-123",
			mockStatus:     http.StatusInternalServerError,
			mockBody: StandardResponse{
				Status:  "error",
				Message: "Internal server error",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				expectedPath := fmt.Sprintf("/v1/organizations/%s/billing", tt.organizationID)
				if r.URL.Path != expectedPath {
					t.Errorf("Expected path '%s', got '%s'", expectedPath, r.URL.Path)
				}
				if r.Method != http.MethodGet {
					t.Errorf("Expected method GET, got %s", r.Method)
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatus)
				json.NewEncoder(w).Encode(tt.mockBody)
			}))
			defer server.Close()

			client, _ := NewClient(&Config{
				BaseURL: server.URL,
				Auth:    AuthConfig{Token: "test-token"},
			})

			result, err := client.Billing.GetBillingInfo(context.Background(), tt.organizationID)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetBillingInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.checkFunc != nil {
				tt.checkFunc(t, result)
			}
		})
	}
}

// TestBillingService_GetSubscription tests retrieving subscription details
func TestBillingService_GetSubscription(t *testing.T) {
	tests := []struct {
		name           string
		organizationID string
		mockStatus     int
		mockBody       interface{}
		wantErr        bool
		checkFunc      func(*testing.T, *Subscription)
	}{
		{
			name:           "successful get subscription",
			organizationID: "org-123",
			mockStatus:     http.StatusOK,
			mockBody: StandardResponse{
				Status:  "success",
				Message: "Subscription retrieved",
				Data: &Subscription{
					ID:                "sub_123",
					OrganizationID:    1,
					PlanID:            "plan_pro",
					PlanName:          "Professional Plan",
					Status:            "active",
					Quantity:          10,
					CancelAtPeriodEnd: false,
					AddOns: []SubscriptionAddOn{
						{
							ID:       "addon_1",
							Name:     "Extra Storage",
							Quantity: 5,
							Price:    10.00,
						},
					},
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, sub *Subscription) {
				if sub.ID != "sub_123" {
					t.Errorf("Expected ID 'sub_123', got '%s'", sub.ID)
				}
				if sub.PlanName != "Professional Plan" {
					t.Errorf("Expected PlanName 'Professional Plan', got '%s'", sub.PlanName)
				}
				if sub.Status != "active" {
					t.Errorf("Expected Status 'active', got '%s'", sub.Status)
				}
				if sub.Quantity != 10 {
					t.Errorf("Expected Quantity 10, got %d", sub.Quantity)
				}
				if len(sub.AddOns) != 1 {
					t.Errorf("Expected 1 AddOn, got %d", len(sub.AddOns))
				}
			},
		},
		{
			name:           "subscription not found",
			organizationID: "invalid-org",
			mockStatus:     http.StatusNotFound,
			mockBody: StandardResponse{
				Status:  "error",
				Message: "Subscription not found",
			},
			wantErr: true,
		},
		{
			name:           "no active subscription",
			organizationID: "org-no-sub",
			mockStatus:     http.StatusOK,
			mockBody: StandardResponse{
				Status:  "success",
				Message: "No active subscription",
				Data: &Subscription{
					ID:     "",
					Status: "none",
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, sub *Subscription) {
				if sub.Status != "none" {
					t.Errorf("Expected Status 'none', got '%s'", sub.Status)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				expectedPath := fmt.Sprintf("/v1/organizations/%s/subscription", tt.organizationID)
				if r.URL.Path != expectedPath {
					t.Errorf("Expected path '%s', got '%s'", expectedPath, r.URL.Path)
				}
				if r.Method != http.MethodGet {
					t.Errorf("Expected method GET, got %s", r.Method)
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatus)
				json.NewEncoder(w).Encode(tt.mockBody)
			}))
			defer server.Close()

			client, _ := NewClient(&Config{
				BaseURL: server.URL,
				Auth:    AuthConfig{Token: "test-token"},
			})

			result, err := client.Billing.GetSubscription(context.Background(), tt.organizationID)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetSubscription() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.checkFunc != nil {
				tt.checkFunc(t, result)
			}
		})
	}
}

// TestBillingService_ListInvoices tests listing invoices with pagination
func TestBillingService_ListInvoices(t *testing.T) {
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
			name:           "successful list invoices",
			organizationID: "org-123",
			opts:           &ListOptions{Page: 1, Limit: 25},
			mockStatus:     http.StatusOK,
			mockBody: PaginatedResponse{
				Status:  "success",
				Message: "Invoices retrieved",
				Data: &[]*Invoice{
					{
						ID:             "inv_001",
						OrganizationID: 1,
						InvoiceNumber:  "INV-2025-001",
						Status:         "paid",
						Amount:         199.99,
						Currency:       "USD",
						LineItems: []InvoiceLineItem{
							{
								Description: "Professional Plan",
								Quantity:    1,
								UnitPrice:   199.99,
								Amount:      199.99,
								Type:        "subscription",
							},
						},
					},
					{
						ID:             "inv_002",
						OrganizationID: 1,
						InvoiceNumber:  "INV-2025-002",
						Status:         "pending",
						Amount:         199.99,
						Currency:       "USD",
					},
				},
				Meta: &PaginationMeta{
					Page:       1,
					TotalPages: 1,
					TotalItems: 2,
					Limit:      25,
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, invoices []*Invoice, meta *PaginationMeta) {
				if len(invoices) != 2 {
					t.Errorf("Expected 2 invoices, got %d", len(invoices))
				}
				if meta.TotalItems != 2 {
					t.Errorf("Expected TotalItems 2, got %d", meta.TotalItems)
				}
				if invoices[0].Status != "paid" {
					t.Errorf("Expected first invoice status 'paid', got '%s'", invoices[0].Status)
				}
				if len(invoices[0].LineItems) != 1 {
					t.Errorf("Expected 1 line item, got %d", len(invoices[0].LineItems))
				}
			},
		},
		{
			name:           "empty invoice list",
			organizationID: "org-new",
			opts:           &ListOptions{Page: 1, Limit: 25},
			mockStatus:     http.StatusOK,
			mockBody: PaginatedResponse{
				Status:  "success",
				Message: "No invoices found",
				Data:    &[]*Invoice{},
				Meta: &PaginationMeta{
					Page:       1,
					TotalPages: 0,
					TotalItems: 0,
					Limit:      25,
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, invoices []*Invoice, meta *PaginationMeta) {
				if len(invoices) != 0 {
					t.Errorf("Expected 0 invoices, got %d", len(invoices))
				}
			},
		},
		{
			name:           "with filters",
			organizationID: "org-123",
			opts:           &ListOptions{Page: 1, Limit: 10, Sort: "created_at", Order: "desc"},
			mockStatus:     http.StatusOK,
			mockBody: PaginatedResponse{
				Status:  "success",
				Message: "Invoices retrieved",
				Data: &[]*Invoice{
					{
						ID:             "inv_003",
						OrganizationID: 1,
						InvoiceNumber:  "INV-2025-003",
						Status:         "paid",
						Amount:         299.99,
						Currency:       "USD",
					},
				},
				Meta: &PaginationMeta{
					Page:       1,
					TotalPages: 1,
					TotalItems: 1,
					Limit:      10,
				},
			},
			wantErr: false,
		},
		{
			name:           "organization not found",
			organizationID: "invalid-org",
			opts:           &ListOptions{Page: 1, Limit: 25},
			mockStatus:     http.StatusNotFound,
			mockBody: PaginatedResponse{
				Status:  "error",
				Message: "Organization not found",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				expectedPath := fmt.Sprintf("/v1/organizations/%s/invoices", tt.organizationID)
				if r.URL.Path != expectedPath {
					t.Errorf("Expected path '%s', got '%s'", expectedPath, r.URL.Path)
				}
				if r.Method != http.MethodGet {
					t.Errorf("Expected method GET, got %s", r.Method)
				}

				// Check query parameters if options provided
				if tt.opts != nil {
					if tt.opts.Sort != "" {
						if r.URL.Query().Get("sort") != tt.opts.Sort {
							t.Errorf("Expected sort '%s', got '%s'", tt.opts.Sort, r.URL.Query().Get("sort"))
						}
					}
					if tt.opts.Order != "" {
						if r.URL.Query().Get("order") != tt.opts.Order {
							t.Errorf("Expected order '%s', got '%s'", tt.opts.Order, r.URL.Query().Get("order"))
						}
					}
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatus)
				json.NewEncoder(w).Encode(tt.mockBody)
			}))
			defer server.Close()

			client, _ := NewClient(&Config{
				BaseURL: server.URL,
				Auth:    AuthConfig{Token: "test-token"},
			})

			invoices, meta, err := client.Billing.ListInvoices(context.Background(), tt.organizationID, tt.opts)

			if (err != nil) != tt.wantErr {
				t.Errorf("ListInvoices() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.checkFunc != nil {
				tt.checkFunc(t, invoices, meta)
			}
		})
	}
}

// TestBillingService_UpdatePaymentMethod tests updating payment method
func TestBillingService_UpdatePaymentMethod(t *testing.T) {
	tests := []struct {
		name           string
		organizationID string
		paymentMethod  *PaymentMethod
		mockStatus     int
		mockBody       interface{}
		wantErr        bool
	}{
		{
			name:           "successful update payment method",
			organizationID: "org-123",
			paymentMethod: &PaymentMethod{
				ID:          "pm_new",
				Type:        "card",
				Last4:       "1234",
				Brand:       "mastercard",
				ExpiryMonth: 12,
				ExpiryYear:  2027,
				IsDefault:   true,
			},
			mockStatus: http.StatusOK,
			mockBody: StandardResponse{
				Status:  "success",
				Message: "Payment method updated",
			},
			wantErr: false,
		},
		{
			name:           "update bank account payment method",
			organizationID: "org-123",
			paymentMethod: &PaymentMethod{
				ID:        "pm_bank",
				Type:      "bank_account",
				Last4:     "6789",
				BankName:  "Chase Bank",
				IsDefault: true,
			},
			mockStatus: http.StatusOK,
			mockBody: StandardResponse{
				Status:  "success",
				Message: "Payment method updated",
			},
			wantErr: false,
		},
		{
			name:           "invalid payment method",
			organizationID: "org-123",
			paymentMethod: &PaymentMethod{
				ID:   "pm_invalid",
				Type: "invalid_type",
			},
			mockStatus: http.StatusBadRequest,
			mockBody: StandardResponse{
				Status:  "error",
				Message: "Invalid payment method type",
			},
			wantErr: true,
		},
		{
			name:           "organization not found",
			organizationID: "invalid-org",
			paymentMethod: &PaymentMethod{
				ID:   "pm_123",
				Type: "card",
			},
			mockStatus: http.StatusNotFound,
			mockBody: StandardResponse{
				Status:  "error",
				Message: "Organization not found",
			},
			wantErr: true,
		},
		{
			name:           "unauthorized access",
			organizationID: "org-123",
			paymentMethod: &PaymentMethod{
				ID:   "pm_123",
				Type: "card",
			},
			mockStatus: http.StatusUnauthorized,
			mockBody: StandardResponse{
				Status:  "error",
				Message: "Unauthorized",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				expectedPath := fmt.Sprintf("/v1/organizations/%s/payment-method", tt.organizationID)
				if r.URL.Path != expectedPath {
					t.Errorf("Expected path '%s', got '%s'", expectedPath, r.URL.Path)
				}
				if r.Method != http.MethodPut {
					t.Errorf("Expected method PUT, got %s", r.Method)
				}

				// Verify request body contains payment method
				var body PaymentMethod
				if err := json.NewDecoder(r.Body).Decode(&body); err == nil {
					if body.Type != tt.paymentMethod.Type {
						t.Errorf("Expected payment method type '%s', got '%s'", tt.paymentMethod.Type, body.Type)
					}
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatus)
				json.NewEncoder(w).Encode(tt.mockBody)
			}))
			defer server.Close()

			client, _ := NewClient(&Config{
				BaseURL: server.URL,
				Auth:    AuthConfig{Token: "test-token"},
			})

			err := client.Billing.UpdatePaymentMethod(context.Background(), tt.organizationID, tt.paymentMethod)

			if (err != nil) != tt.wantErr {
				t.Errorf("UpdatePaymentMethod() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestBillingInfo_JSON tests JSON marshaling/unmarshaling of BillingInfo
func TestBillingInfo_JSON(t *testing.T) {
	t.Run("marshal and unmarshal billing info", func(t *testing.T) {
		original := &BillingInfo{
			OrganizationID:   1,
			StripeCustomerID: "cus_test",
			CurrentBalance:   -25.50,
			Credits:          50.00,
			BillingCycle:     "monthly",
			PaymentMethod: &PaymentMethod{
				ID:          "pm_test",
				Type:        "card",
				Last4:       "4242",
				Brand:       "visa",
				ExpiryMonth: 12,
				ExpiryYear:  2025,
				IsDefault:   true,
			},
			BillingAddress: &BillingAddress{
				Line1:      "123 Main St",
				City:       "San Francisco",
				State:      "CA",
				PostalCode: "94105",
				Country:    "US",
			},
			TaxInfo: &TaxInfo{
				TaxID:       "US123456789",
				TaxType:     "vat",
				CompanyName: "Test Company Inc",
				TaxExempt:   false,
			},
		}

		// Marshal to JSON
		data, err := json.Marshal(original)
		if err != nil {
			t.Fatalf("Failed to marshal BillingInfo: %v", err)
		}

		// Unmarshal back
		var decoded BillingInfo
		if err := json.Unmarshal(data, &decoded); err != nil {
			t.Fatalf("Failed to unmarshal BillingInfo: %v", err)
		}

		// Verify fields
		if decoded.OrganizationID != original.OrganizationID {
			t.Errorf("OrganizationID mismatch: got %d, want %d", decoded.OrganizationID, original.OrganizationID)
		}
		if decoded.StripeCustomerID != original.StripeCustomerID {
			t.Errorf("StripeCustomerID mismatch: got %s, want %s", decoded.StripeCustomerID, original.StripeCustomerID)
		}
		if decoded.CurrentBalance != original.CurrentBalance {
			t.Errorf("CurrentBalance mismatch: got %f, want %f", decoded.CurrentBalance, original.CurrentBalance)
		}
	})
}

// TestSubscription_JSON tests JSON marshaling/unmarshaling of Subscription
func TestSubscription_JSON(t *testing.T) {
	t.Run("marshal and unmarshal subscription", func(t *testing.T) {
		original := &Subscription{
			ID:                "sub_test",
			OrganizationID:    1,
			PlanID:            "plan_pro",
			PlanName:          "Professional",
			Status:            "active",
			Quantity:          5,
			CancelAtPeriodEnd: false,
			AddOns: []SubscriptionAddOn{
				{
					ID:       "addon_1",
					Name:     "Extra Storage",
					Quantity: 10,
					Price:    5.00,
				},
			},
		}

		// Marshal to JSON
		data, err := json.Marshal(original)
		if err != nil {
			t.Fatalf("Failed to marshal Subscription: %v", err)
		}

		// Unmarshal back
		var decoded Subscription
		if err := json.Unmarshal(data, &decoded); err != nil {
			t.Fatalf("Failed to unmarshal Subscription: %v", err)
		}

		// Verify fields
		if decoded.ID != original.ID {
			t.Errorf("ID mismatch: got %s, want %s", decoded.ID, original.ID)
		}
		if decoded.PlanName != original.PlanName {
			t.Errorf("PlanName mismatch: got %s, want %s", decoded.PlanName, original.PlanName)
		}
		if len(decoded.AddOns) != len(original.AddOns) {
			t.Errorf("AddOns length mismatch: got %d, want %d", len(decoded.AddOns), len(original.AddOns))
		}
	})
}

// TestInvoice_JSON tests JSON marshaling/unmarshaling of Invoice
func TestInvoice_JSON(t *testing.T) {
	t.Run("marshal and unmarshal invoice", func(t *testing.T) {
		original := &Invoice{
			ID:             "inv_test",
			OrganizationID: 1,
			InvoiceNumber:  "INV-2025-001",
			Status:         "paid",
			Amount:         199.99,
			Currency:       "USD",
			PDFURL:         "https://example.com/invoice.pdf",
			LineItems: []InvoiceLineItem{
				{
					Description: "Professional Plan",
					Quantity:    1,
					UnitPrice:   199.99,
					Amount:      199.99,
					Type:        "subscription",
				},
				{
					Description: "Extra Users",
					Quantity:    5,
					UnitPrice:   10.00,
					Amount:      50.00,
					Type:        "add_on",
				},
			},
		}

		// Marshal to JSON
		data, err := json.Marshal(original)
		if err != nil {
			t.Fatalf("Failed to marshal Invoice: %v", err)
		}

		// Unmarshal back
		var decoded Invoice
		if err := json.Unmarshal(data, &decoded); err != nil {
			t.Fatalf("Failed to unmarshal Invoice: %v", err)
		}

		// Verify fields
		if decoded.InvoiceNumber != original.InvoiceNumber {
			t.Errorf("InvoiceNumber mismatch: got %s, want %s", decoded.InvoiceNumber, original.InvoiceNumber)
		}
		if decoded.Amount != original.Amount {
			t.Errorf("Amount mismatch: got %f, want %f", decoded.Amount, original.Amount)
		}
		if len(decoded.LineItems) != len(original.LineItems) {
			t.Errorf("LineItems length mismatch: got %d, want %d", len(decoded.LineItems), len(original.LineItems))
		}
	})
}

// ============================================================================
// Self-Service Subscription Method Tests (Task #3939)
// ============================================================================

// TestBillingService_GetMySubscription tests retrieving the current user's subscription
func TestBillingService_GetMySubscription(t *testing.T) {
	tests := []struct {
		name       string
		mockStatus int
		mockBody   interface{}
		wantErr    bool
		checkFunc  func(*testing.T, *SubscriptionResponse)
	}{
		{
			name:       "successful get my subscription",
			mockStatus: http.StatusOK,
			mockBody: StandardResponse{
				Status:  "success",
				Message: "Subscription retrieved",
				Data: &SubscriptionResponse{
					ID:                "sub_123",
					OrganizationID:    1,
					PlanID:            "plan_pro",
					PlanName:          "Professional",
					Status:            "active",
					BillingCycle:      "monthly",
					CancelAtPeriodEnd: false,
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, sub *SubscriptionResponse) {
				if sub.ID != "sub_123" {
					t.Errorf("Expected ID 'sub_123', got '%s'", sub.ID)
				}
				if sub.Status != "active" {
					t.Errorf("Expected Status 'active', got '%s'", sub.Status)
				}
				if sub.BillingCycle != "monthly" {
					t.Errorf("Expected BillingCycle 'monthly', got '%s'", sub.BillingCycle)
				}
			},
		},
		{
			name:       "no active subscription",
			mockStatus: http.StatusOK,
			mockBody: StandardResponse{
				Status:  "success",
				Message: "No active subscription",
				Data: &SubscriptionResponse{
					Status: "none",
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, sub *SubscriptionResponse) {
				if sub.Status != "none" {
					t.Errorf("Expected Status 'none', got '%s'", sub.Status)
				}
			},
		},
		{
			name:       "unauthorized",
			mockStatus: http.StatusUnauthorized,
			mockBody: StandardResponse{
				Status:  "error",
				Message: "Unauthorized",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/v1/subscription" {
					t.Errorf("Expected path '/v1/subscription', got '%s'", r.URL.Path)
				}
				if r.Method != http.MethodGet {
					t.Errorf("Expected method GET, got %s", r.Method)
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatus)
				json.NewEncoder(w).Encode(tt.mockBody)
			}))
			defer server.Close()

			client, _ := NewClient(&Config{
				BaseURL: server.URL,
				Auth:    AuthConfig{Token: "test-token"},
			})

			result, err := client.Billing.GetMySubscription(context.Background())

			if (err != nil) != tt.wantErr {
				t.Errorf("GetMySubscription() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.checkFunc != nil {
				tt.checkFunc(t, result)
			}
		})
	}
}

// TestBillingService_CreateCheckout tests creating a checkout session
func TestBillingService_CreateCheckout(t *testing.T) {
	tests := []struct {
		name       string
		request    *CreateCheckoutRequest
		mockStatus int
		mockBody   interface{}
		wantErr    bool
		checkFunc  func(*testing.T, *CheckoutSessionResponse)
	}{
		{
			name: "successful checkout creation",
			request: &CreateCheckoutRequest{
				PlanID:       "plan_pro",
				BillingCycle: "monthly",
				SuccessURL:   "https://example.com/success",
				CancelURL:    "https://example.com/cancel",
			},
			mockStatus: http.StatusOK,
			mockBody: StandardResponse{
				Status:  "success",
				Message: "Checkout session created",
				Data: &CheckoutSessionResponse{
					SessionID:  "cs_test_123",
					SessionURL: "https://checkout.stripe.com/c/pay/cs_test_123",
					ExpiresAt:  1735689600,
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, resp *CheckoutSessionResponse) {
				if resp.SessionID != "cs_test_123" {
					t.Errorf("Expected SessionID 'cs_test_123', got '%s'", resp.SessionID)
				}
				if resp.SessionURL == "" {
					t.Error("Expected SessionURL to be present")
				}
			},
		},
		{
			name: "invalid plan",
			request: &CreateCheckoutRequest{
				PlanID:       "invalid_plan",
				BillingCycle: "monthly",
				SuccessURL:   "https://example.com/success",
				CancelURL:    "https://example.com/cancel",
			},
			mockStatus: http.StatusBadRequest,
			mockBody: StandardResponse{
				Status:  "error",
				Message: "Invalid plan ID",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/v1/subscription/checkout" {
					t.Errorf("Expected path '/v1/subscription/checkout', got '%s'", r.URL.Path)
				}
				if r.Method != http.MethodPost {
					t.Errorf("Expected method POST, got %s", r.Method)
				}

				// Verify request body
				var body CreateCheckoutRequest
				if err := json.NewDecoder(r.Body).Decode(&body); err == nil {
					if body.PlanID != tt.request.PlanID {
						t.Errorf("Expected PlanID '%s', got '%s'", tt.request.PlanID, body.PlanID)
					}
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatus)
				json.NewEncoder(w).Encode(tt.mockBody)
			}))
			defer server.Close()

			client, _ := NewClient(&Config{
				BaseURL: server.URL,
				Auth:    AuthConfig{Token: "test-token"},
			})

			result, err := client.Billing.CreateCheckout(context.Background(), tt.request)

			if (err != nil) != tt.wantErr {
				t.Errorf("CreateCheckout() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.checkFunc != nil {
				tt.checkFunc(t, result)
			}
		})
	}
}

// TestBillingService_UpdateMySubscription tests updating a subscription
func TestBillingService_UpdateMySubscription(t *testing.T) {
	tests := []struct {
		name       string
		request    *UpdateSubscriptionRequest
		mockStatus int
		mockBody   interface{}
		wantErr    bool
		checkFunc  func(*testing.T, *SubscriptionResponse)
	}{
		{
			name: "successful upgrade",
			request: &UpdateSubscriptionRequest{
				PlanID:       "plan_enterprise",
				BillingCycle: "yearly",
			},
			mockStatus: http.StatusOK,
			mockBody: StandardResponse{
				Status:  "success",
				Message: "Subscription updated",
				Data: &SubscriptionResponse{
					ID:           "sub_123",
					PlanID:       "plan_enterprise",
					PlanName:     "Enterprise",
					Status:       "active",
					BillingCycle: "yearly",
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, sub *SubscriptionResponse) {
				if sub.PlanID != "plan_enterprise" {
					t.Errorf("Expected PlanID 'plan_enterprise', got '%s'", sub.PlanID)
				}
				if sub.BillingCycle != "yearly" {
					t.Errorf("Expected BillingCycle 'yearly', got '%s'", sub.BillingCycle)
				}
			},
		},
		{
			name: "no active subscription to update",
			request: &UpdateSubscriptionRequest{
				PlanID: "plan_pro",
			},
			mockStatus: http.StatusNotFound,
			mockBody: StandardResponse{
				Status:  "error",
				Message: "No active subscription found",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/v1/subscription" {
					t.Errorf("Expected path '/v1/subscription', got '%s'", r.URL.Path)
				}
				if r.Method != http.MethodPut {
					t.Errorf("Expected method PUT, got %s", r.Method)
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatus)
				json.NewEncoder(w).Encode(tt.mockBody)
			}))
			defer server.Close()

			client, _ := NewClient(&Config{
				BaseURL: server.URL,
				Auth:    AuthConfig{Token: "test-token"},
			})

			result, err := client.Billing.UpdateMySubscription(context.Background(), tt.request)

			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateMySubscription() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.checkFunc != nil {
				tt.checkFunc(t, result)
			}
		})
	}
}

// TestBillingService_CancelMySubscription tests canceling a subscription
func TestBillingService_CancelMySubscription(t *testing.T) {
	tests := []struct {
		name       string
		request    *CancelSubscriptionRequest
		mockStatus int
		mockBody   interface{}
		wantErr    bool
	}{
		{
			name: "successful cancel at period end",
			request: &CancelSubscriptionRequest{
				CancelAtPeriod: true,
				Reason:         "Too expensive",
			},
			mockStatus: http.StatusOK,
			mockBody: StandardResponse{
				Status:  "success",
				Message: "Subscription will be canceled at period end",
			},
			wantErr: false,
		},
		{
			name: "immediate cancellation",
			request: &CancelSubscriptionRequest{
				CancelAtPeriod: false,
				Reason:         "No longer needed",
			},
			mockStatus: http.StatusOK,
			mockBody: StandardResponse{
				Status:  "success",
				Message: "Subscription canceled",
			},
			wantErr: false,
		},
		{
			name: "no subscription to cancel",
			request: &CancelSubscriptionRequest{
				CancelAtPeriod: true,
			},
			mockStatus: http.StatusNotFound,
			mockBody: StandardResponse{
				Status:  "error",
				Message: "No active subscription found",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/v1/subscription" {
					t.Errorf("Expected path '/v1/subscription', got '%s'", r.URL.Path)
				}
				if r.Method != http.MethodDelete {
					t.Errorf("Expected method DELETE, got %s", r.Method)
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatus)
				json.NewEncoder(w).Encode(tt.mockBody)
			}))
			defer server.Close()

			client, _ := NewClient(&Config{
				BaseURL: server.URL,
				Auth:    AuthConfig{Token: "test-token"},
			})

			err := client.Billing.CancelMySubscription(context.Background(), tt.request)

			if (err != nil) != tt.wantErr {
				t.Errorf("CancelMySubscription() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestBillingService_CreatePortalSession tests creating a Stripe portal session
func TestBillingService_CreatePortalSession(t *testing.T) {
	tests := []struct {
		name       string
		returnURL  string
		mockStatus int
		mockBody   interface{}
		wantErr    bool
		checkFunc  func(*testing.T, *PortalSessionResponse)
	}{
		{
			name:       "successful portal session creation",
			returnURL:  "https://example.com/settings",
			mockStatus: http.StatusOK,
			mockBody: StandardResponse{
				Status:  "success",
				Message: "Portal session created",
				Data: &PortalSessionResponse{
					URL:       "https://billing.stripe.com/session/bps_test_123",
					ExpiresAt: 1735689600,
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, resp *PortalSessionResponse) {
				if resp.URL == "" {
					t.Error("Expected URL to be present")
				}
			},
		},
		{
			name:       "no customer for portal",
			returnURL:  "https://example.com/settings",
			mockStatus: http.StatusBadRequest,
			mockBody: StandardResponse{
				Status:  "error",
				Message: "No Stripe customer found",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/v1/billing/portal" {
					t.Errorf("Expected path '/v1/billing/portal', got '%s'", r.URL.Path)
				}
				if r.Method != http.MethodPost {
					t.Errorf("Expected method POST, got %s", r.Method)
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatus)
				json.NewEncoder(w).Encode(tt.mockBody)
			}))
			defer server.Close()

			client, _ := NewClient(&Config{
				BaseURL: server.URL,
				Auth:    AuthConfig{Token: "test-token"},
			})

			result, err := client.Billing.CreatePortalSession(context.Background(), tt.returnURL)

			if (err != nil) != tt.wantErr {
				t.Errorf("CreatePortalSession() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.checkFunc != nil {
				tt.checkFunc(t, result)
			}
		})
	}
}

// ============================================================================
// Plan Method Tests (Task #3939)
// ============================================================================

// TestBillingService_ListPlans tests listing available plans
func TestBillingService_ListPlans(t *testing.T) {
	tests := []struct {
		name       string
		mockStatus int
		mockBody   interface{}
		wantErr    bool
		checkFunc  func(*testing.T, []*Plan)
	}{
		{
			name:       "successful list plans",
			mockStatus: http.StatusOK,
			mockBody: StandardResponse{
				Status:  "success",
				Message: "Plans retrieved",
				Data: &[]*Plan{
					{
						ID:           "plan_starter",
						Name:         "Starter",
						Description:  "For small teams",
						MonthlyPrice: 2900,
						YearlyPrice:  29000,
						Currency:     "USD",
						IsPublic:     true,
						Limits: PlanLimits{
							MaxServers:        5,
							MaxUsers:          3,
							MaxProbes:         10,
							DataRetentionDays: 30,
						},
					},
					{
						ID:           "plan_pro",
						Name:         "Professional",
						Description:  "For growing teams",
						MonthlyPrice: 9900,
						YearlyPrice:  99000,
						Currency:     "USD",
						IsPublic:     true,
						Limits: PlanLimits{
							MaxServers:        50,
							MaxUsers:          10,
							MaxProbes:         100,
							DataRetentionDays: 90,
						},
					},
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, plans []*Plan) {
				if len(plans) != 2 {
					t.Errorf("Expected 2 plans, got %d", len(plans))
				}
				if plans[0].ID != "plan_starter" {
					t.Errorf("Expected first plan ID 'plan_starter', got '%s'", plans[0].ID)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/v1/pricing/plans" {
					t.Errorf("Expected path '/v1/pricing/plans', got '%s'", r.URL.Path)
				}
				if r.Method != http.MethodGet {
					t.Errorf("Expected method GET, got %s", r.Method)
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatus)
				json.NewEncoder(w).Encode(tt.mockBody)
			}))
			defer server.Close()

			client, _ := NewClient(&Config{
				BaseURL: server.URL,
				Auth:    AuthConfig{Token: "test-token"},
			})

			result, err := client.Billing.ListPlans(context.Background())

			if (err != nil) != tt.wantErr {
				t.Errorf("ListPlans() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.checkFunc != nil {
				tt.checkFunc(t, result)
			}
		})
	}
}

// TestBillingService_GetPlan tests retrieving a specific plan
func TestBillingService_GetPlan(t *testing.T) {
	tests := []struct {
		name       string
		planID     string
		mockStatus int
		mockBody   interface{}
		wantErr    bool
		checkFunc  func(*testing.T, *Plan)
	}{
		{
			name:       "successful get plan",
			planID:     "plan_pro",
			mockStatus: http.StatusOK,
			mockBody: StandardResponse{
				Status:  "success",
				Message: "Plan retrieved",
				Data: &Plan{
					ID:           "plan_pro",
					Name:         "Professional",
					Description:  "For growing teams",
					MonthlyPrice: 9900,
					YearlyPrice:  99000,
					Currency:     "USD",
					IsPublic:     true,
					Features: []PlanFeature{
						{Name: "API Access", Included: true},
						{Name: "Custom Alerts", Included: true},
					},
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, plan *Plan) {
				if plan.ID != "plan_pro" {
					t.Errorf("Expected ID 'plan_pro', got '%s'", plan.ID)
				}
				if len(plan.Features) != 2 {
					t.Errorf("Expected 2 features, got %d", len(plan.Features))
				}
			},
		},
		{
			name:       "plan not found",
			planID:     "invalid_plan",
			mockStatus: http.StatusNotFound,
			mockBody: StandardResponse{
				Status:  "error",
				Message: "Plan not found",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				expectedPath := fmt.Sprintf("/v1/pricing/plans/%s", tt.planID)
				if r.URL.Path != expectedPath {
					t.Errorf("Expected path '%s', got '%s'", expectedPath, r.URL.Path)
				}
				if r.Method != http.MethodGet {
					t.Errorf("Expected method GET, got %s", r.Method)
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatus)
				json.NewEncoder(w).Encode(tt.mockBody)
			}))
			defer server.Close()

			client, _ := NewClient(&Config{
				BaseURL: server.URL,
				Auth:    AuthConfig{Token: "test-token"},
			})

			result, err := client.Billing.GetPlan(context.Background(), tt.planID)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetPlan() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.checkFunc != nil {
				tt.checkFunc(t, result)
			}
		})
	}
}

// TestBillingService_GetPlanFeatures tests retrieving the feature matrix
func TestBillingService_GetPlanFeatures(t *testing.T) {
	tests := []struct {
		name       string
		mockStatus int
		mockBody   interface{}
		wantErr    bool
		checkFunc  func(*testing.T, *FeatureMatrix)
	}{
		{
			name:       "successful get features",
			mockStatus: http.StatusOK,
			mockBody: StandardResponse{
				Status:  "success",
				Message: "Features retrieved",
				Data: &FeatureMatrix{
					Features: []FeatureRow{
						{
							Name:        "API Access",
							Description: "Access to REST API",
							Category:    "Core",
							PlanValues: map[string]string{
								"starter": "Limited",
								"pro":     "Full",
							},
						},
					},
					Plans: []Plan{
						{ID: "plan_starter", Name: "Starter"},
						{ID: "plan_pro", Name: "Professional"},
					},
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, matrix *FeatureMatrix) {
				if len(matrix.Features) != 1 {
					t.Errorf("Expected 1 feature row, got %d", len(matrix.Features))
				}
				if len(matrix.Plans) != 2 {
					t.Errorf("Expected 2 plans, got %d", len(matrix.Plans))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/v1/pricing/features" {
					t.Errorf("Expected path '/v1/pricing/features', got '%s'", r.URL.Path)
				}
				if r.Method != http.MethodGet {
					t.Errorf("Expected method GET, got %s", r.Method)
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatus)
				json.NewEncoder(w).Encode(tt.mockBody)
			}))
			defer server.Close()

			client, _ := NewClient(&Config{
				BaseURL: server.URL,
				Auth:    AuthConfig{Token: "test-token"},
			})

			result, err := client.Billing.GetPlanFeatures(context.Background())

			if (err != nil) != tt.wantErr {
				t.Errorf("GetPlanFeatures() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.checkFunc != nil {
				tt.checkFunc(t, result)
			}
		})
	}
}

// ============================================================================
// Invoice Method Tests (Task #3939)
// ============================================================================

// TestBillingService_ListMyInvoices tests listing invoices for authenticated user
func TestBillingService_ListMyInvoices(t *testing.T) {
	tests := []struct {
		name       string
		opts       *ListInvoiceOptions
		mockStatus int
		mockBody   interface{}
		wantErr    bool
		checkFunc  func(*testing.T, []*Invoice, *PaginationMeta)
	}{
		{
			name:       "successful list my invoices",
			opts:       &ListInvoiceOptions{Page: 1, Limit: 10},
			mockStatus: http.StatusOK,
			mockBody: PaginatedResponse{
				Status:  "success",
				Message: "Invoices retrieved",
				Data: &[]*Invoice{
					{
						ID:            "inv_001",
						InvoiceNumber: "INV-2025-001",
						Status:        "paid",
						Amount:        199.99,
						Currency:      "USD",
					},
				},
				Meta: &PaginationMeta{
					Page:       1,
					TotalPages: 1,
					TotalItems: 1,
					Limit:      10,
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, invoices []*Invoice, meta *PaginationMeta) {
				if len(invoices) != 1 {
					t.Errorf("Expected 1 invoice, got %d", len(invoices))
				}
				if meta.TotalItems != 1 {
					t.Errorf("Expected TotalItems 1, got %d", meta.TotalItems)
				}
			},
		},
		{
			name:       "filter by status",
			opts:       &ListInvoiceOptions{Page: 1, Limit: 10, Status: "paid"},
			mockStatus: http.StatusOK,
			mockBody: PaginatedResponse{
				Status: "success",
				Data:   &[]*Invoice{},
				Meta: &PaginationMeta{
					Page:       1,
					TotalPages: 0,
					TotalItems: 0,
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/v1/billing/invoices" {
					t.Errorf("Expected path '/v1/billing/invoices', got '%s'", r.URL.Path)
				}
				if r.Method != http.MethodGet {
					t.Errorf("Expected method GET, got %s", r.Method)
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatus)
				json.NewEncoder(w).Encode(tt.mockBody)
			}))
			defer server.Close()

			client, _ := NewClient(&Config{
				BaseURL: server.URL,
				Auth:    AuthConfig{Token: "test-token"},
			})

			invoices, meta, err := client.Billing.ListMyInvoices(context.Background(), tt.opts)

			if (err != nil) != tt.wantErr {
				t.Errorf("ListMyInvoices() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.checkFunc != nil {
				tt.checkFunc(t, invoices, meta)
			}
		})
	}
}

// TestBillingService_GetMyInvoice tests retrieving a specific invoice
func TestBillingService_GetMyInvoice(t *testing.T) {
	tests := []struct {
		name       string
		invoiceID  string
		mockStatus int
		mockBody   interface{}
		wantErr    bool
		checkFunc  func(*testing.T, *Invoice)
	}{
		{
			name:       "successful get invoice",
			invoiceID:  "inv_001",
			mockStatus: http.StatusOK,
			mockBody: StandardResponse{
				Status:  "success",
				Message: "Invoice retrieved",
				Data: &Invoice{
					ID:            "inv_001",
					InvoiceNumber: "INV-2025-001",
					Status:        "paid",
					Amount:        199.99,
					Currency:      "USD",
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, inv *Invoice) {
				if inv.ID != "inv_001" {
					t.Errorf("Expected ID 'inv_001', got '%s'", inv.ID)
				}
			},
		},
		{
			name:       "invoice not found",
			invoiceID:  "invalid_inv",
			mockStatus: http.StatusNotFound,
			mockBody: StandardResponse{
				Status:  "error",
				Message: "Invoice not found",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				expectedPath := fmt.Sprintf("/v1/billing/invoices/%s", tt.invoiceID)
				if r.URL.Path != expectedPath {
					t.Errorf("Expected path '%s', got '%s'", expectedPath, r.URL.Path)
				}
				if r.Method != http.MethodGet {
					t.Errorf("Expected method GET, got %s", r.Method)
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatus)
				json.NewEncoder(w).Encode(tt.mockBody)
			}))
			defer server.Close()

			client, _ := NewClient(&Config{
				BaseURL: server.URL,
				Auth:    AuthConfig{Token: "test-token"},
			})

			result, err := client.Billing.GetMyInvoice(context.Background(), tt.invoiceID)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetMyInvoice() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.checkFunc != nil {
				tt.checkFunc(t, result)
			}
		})
	}
}

// TestBillingService_DownloadInvoicePDF tests downloading invoice PDF
func TestBillingService_DownloadInvoicePDF(t *testing.T) {
	tests := []struct {
		name       string
		invoiceID  string
		mockStatus int
		mockBody   []byte
		wantErr    bool
		checkFunc  func(*testing.T, []byte)
	}{
		{
			name:       "successful download",
			invoiceID:  "inv_001",
			mockStatus: http.StatusOK,
			mockBody:   []byte("%PDF-1.4 fake pdf content"),
			wantErr:    false,
			checkFunc: func(t *testing.T, data []byte) {
				if len(data) == 0 {
					t.Error("Expected PDF data, got empty")
				}
			},
		},
		{
			name:       "invoice not found",
			invoiceID:  "invalid_inv",
			mockStatus: http.StatusNotFound,
			mockBody:   []byte(`{"status":"error","message":"Invoice not found"}`),
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				expectedPath := fmt.Sprintf("/v1/billing/invoices/%s/download", tt.invoiceID)
				if r.URL.Path != expectedPath {
					t.Errorf("Expected path '%s', got '%s'", expectedPath, r.URL.Path)
				}
				if r.Method != http.MethodGet {
					t.Errorf("Expected method GET, got %s", r.Method)
				}

				if tt.mockStatus == http.StatusOK {
					w.Header().Set("Content-Type", "application/pdf")
				} else {
					w.Header().Set("Content-Type", "application/json")
				}
				w.WriteHeader(tt.mockStatus)
				w.Write(tt.mockBody)
			}))
			defer server.Close()

			client, _ := NewClient(&Config{
				BaseURL: server.URL,
				Auth:    AuthConfig{Token: "test-token"},
			})

			result, err := client.Billing.DownloadInvoicePDF(context.Background(), tt.invoiceID)

			if (err != nil) != tt.wantErr {
				t.Errorf("DownloadInvoicePDF() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.checkFunc != nil {
				tt.checkFunc(t, result)
			}
		})
	}
}

// TestBillingService_GetBillingHistory tests retrieving billing history
func TestBillingService_GetBillingHistory(t *testing.T) {
	tests := []struct {
		name       string
		opts       *ListOptions
		mockStatus int
		mockBody   interface{}
		wantErr    bool
		checkFunc  func(*testing.T, *BillingHistoryResponse)
	}{
		{
			name:       "successful get history",
			opts:       &ListOptions{Page: 1, Limit: 10},
			mockStatus: http.StatusOK,
			mockBody: StandardResponse{
				Status:  "success",
				Message: "History retrieved",
				Data: &BillingHistoryResponse{
					Invoices: []*Invoice{
						{ID: "inv_001", Amount: 199.99, Status: "paid"},
						{ID: "inv_002", Amount: 199.99, Status: "paid"},
					},
					TotalSpent: 39998,
					Currency:   "USD",
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, resp *BillingHistoryResponse) {
				if len(resp.Invoices) != 2 {
					t.Errorf("Expected 2 invoices, got %d", len(resp.Invoices))
				}
				if resp.TotalSpent != 39998 {
					t.Errorf("Expected TotalSpent 39998, got %d", resp.TotalSpent)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/v1/billing/history" {
					t.Errorf("Expected path '/v1/billing/history', got '%s'", r.URL.Path)
				}
				if r.Method != http.MethodGet {
					t.Errorf("Expected method GET, got %s", r.Method)
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatus)
				json.NewEncoder(w).Encode(tt.mockBody)
			}))
			defer server.Close()

			client, _ := NewClient(&Config{
				BaseURL: server.URL,
				Auth:    AuthConfig{Token: "test-token"},
			})

			result, err := client.Billing.GetBillingHistory(context.Background(), tt.opts)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetBillingHistory() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.checkFunc != nil {
				tt.checkFunc(t, result)
			}
		})
	}
}

// ============================================================================
// Payment Method Tests (Task #3939)
// ============================================================================

// TestBillingService_ListPaymentMethods tests listing payment methods
func TestBillingService_ListPaymentMethods(t *testing.T) {
	tests := []struct {
		name       string
		mockStatus int
		mockBody   interface{}
		wantErr    bool
		checkFunc  func(*testing.T, []*PaymentMethod)
	}{
		{
			name:       "successful list payment methods",
			mockStatus: http.StatusOK,
			mockBody: StandardResponse{
				Status:  "success",
				Message: "Payment methods retrieved",
				Data: &[]*PaymentMethod{
					{
						ID:          "pm_123",
						Type:        "card",
						Last4:       "4242",
						Brand:       "visa",
						ExpiryMonth: 12,
						ExpiryYear:  2027,
						IsDefault:   true,
					},
					{
						ID:        "pm_456",
						Type:      "bank_account",
						Last4:     "6789",
						BankName:  "Chase",
						IsDefault: false,
					},
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, methods []*PaymentMethod) {
				if len(methods) != 2 {
					t.Errorf("Expected 2 payment methods, got %d", len(methods))
				}
				if methods[0].IsDefault != true {
					t.Error("Expected first method to be default")
				}
			},
		},
		{
			name:       "no payment methods",
			mockStatus: http.StatusOK,
			mockBody: StandardResponse{
				Status:  "success",
				Message: "No payment methods",
				Data:    &[]*PaymentMethod{},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, methods []*PaymentMethod) {
				if len(methods) != 0 {
					t.Errorf("Expected 0 payment methods, got %d", len(methods))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/v1/billing/payment-methods" {
					t.Errorf("Expected path '/v1/billing/payment-methods', got '%s'", r.URL.Path)
				}
				if r.Method != http.MethodGet {
					t.Errorf("Expected method GET, got %s", r.Method)
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatus)
				json.NewEncoder(w).Encode(tt.mockBody)
			}))
			defer server.Close()

			client, _ := NewClient(&Config{
				BaseURL: server.URL,
				Auth:    AuthConfig{Token: "test-token"},
			})

			result, err := client.Billing.ListPaymentMethods(context.Background())

			if (err != nil) != tt.wantErr {
				t.Errorf("ListPaymentMethods() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.checkFunc != nil {
				tt.checkFunc(t, result)
			}
		})
	}
}

// TestBillingService_AddPaymentMethod tests adding a new payment method
func TestBillingService_AddPaymentMethod(t *testing.T) {
	tests := []struct {
		name       string
		request    *AddPaymentMethodRequest
		mockStatus int
		mockBody   interface{}
		wantErr    bool
		checkFunc  func(*testing.T, *PaymentMethod)
	}{
		{
			name: "successful add payment method",
			request: &AddPaymentMethodRequest{
				PaymentMethodToken: "pm_token_123",
				SetDefault:         true,
			},
			mockStatus: http.StatusOK,
			mockBody: StandardResponse{
				Status:  "success",
				Message: "Payment method added",
				Data: &PaymentMethod{
					ID:          "pm_new",
					Type:        "card",
					Last4:       "1234",
					Brand:       "mastercard",
					ExpiryMonth: 10,
					ExpiryYear:  2028,
					IsDefault:   true,
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, pm *PaymentMethod) {
				if pm.ID != "pm_new" {
					t.Errorf("Expected ID 'pm_new', got '%s'", pm.ID)
				}
				if pm.IsDefault != true {
					t.Error("Expected payment method to be default")
				}
			},
		},
		{
			name: "invalid token",
			request: &AddPaymentMethodRequest{
				PaymentMethodToken: "invalid_token",
			},
			mockStatus: http.StatusBadRequest,
			mockBody: StandardResponse{
				Status:  "error",
				Message: "Invalid payment method token",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/v1/billing/payment-methods" {
					t.Errorf("Expected path '/v1/billing/payment-methods', got '%s'", r.URL.Path)
				}
				if r.Method != http.MethodPost {
					t.Errorf("Expected method POST, got %s", r.Method)
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatus)
				json.NewEncoder(w).Encode(tt.mockBody)
			}))
			defer server.Close()

			client, _ := NewClient(&Config{
				BaseURL: server.URL,
				Auth:    AuthConfig{Token: "test-token"},
			})

			result, err := client.Billing.AddPaymentMethod(context.Background(), tt.request)

			if (err != nil) != tt.wantErr {
				t.Errorf("AddPaymentMethod() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.checkFunc != nil {
				tt.checkFunc(t, result)
			}
		})
	}
}

// TestBillingService_RemovePaymentMethod tests removing a payment method
func TestBillingService_RemovePaymentMethod(t *testing.T) {
	tests := []struct {
		name            string
		paymentMethodID string
		mockStatus      int
		mockBody        interface{}
		wantErr         bool
	}{
		{
			name:            "successful remove",
			paymentMethodID: "pm_123",
			mockStatus:      http.StatusOK,
			mockBody: StandardResponse{
				Status:  "success",
				Message: "Payment method removed",
			},
			wantErr: false,
		},
		{
			name:            "payment method not found",
			paymentMethodID: "pm_invalid",
			mockStatus:      http.StatusNotFound,
			mockBody: StandardResponse{
				Status:  "error",
				Message: "Payment method not found",
			},
			wantErr: true,
		},
		{
			name:            "cannot remove default",
			paymentMethodID: "pm_default",
			mockStatus:      http.StatusBadRequest,
			mockBody: StandardResponse{
				Status:  "error",
				Message: "Cannot remove default payment method",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				expectedPath := fmt.Sprintf("/v1/billing/payment-methods/%s", tt.paymentMethodID)
				if r.URL.Path != expectedPath {
					t.Errorf("Expected path '%s', got '%s'", expectedPath, r.URL.Path)
				}
				if r.Method != http.MethodDelete {
					t.Errorf("Expected method DELETE, got %s", r.Method)
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatus)
				json.NewEncoder(w).Encode(tt.mockBody)
			}))
			defer server.Close()

			client, _ := NewClient(&Config{
				BaseURL: server.URL,
				Auth:    AuthConfig{Token: "test-token"},
			})

			err := client.Billing.RemovePaymentMethod(context.Background(), tt.paymentMethodID)

			if (err != nil) != tt.wantErr {
				t.Errorf("RemovePaymentMethod() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestBillingService_SetDefaultPaymentMethod tests setting a default payment method
func TestBillingService_SetDefaultPaymentMethod(t *testing.T) {
	tests := []struct {
		name            string
		paymentMethodID string
		mockStatus      int
		mockBody        interface{}
		wantErr         bool
	}{
		{
			name:            "successful set default",
			paymentMethodID: "pm_123",
			mockStatus:      http.StatusOK,
			mockBody: StandardResponse{
				Status:  "success",
				Message: "Default payment method updated",
			},
			wantErr: false,
		},
		{
			name:            "payment method not found",
			paymentMethodID: "pm_invalid",
			mockStatus:      http.StatusNotFound,
			mockBody: StandardResponse{
				Status:  "error",
				Message: "Payment method not found",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				expectedPath := fmt.Sprintf("/v1/billing/payment-methods/%s/default", tt.paymentMethodID)
				if r.URL.Path != expectedPath {
					t.Errorf("Expected path '%s', got '%s'", expectedPath, r.URL.Path)
				}
				if r.Method != http.MethodPut {
					t.Errorf("Expected method PUT, got %s", r.Method)
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.mockStatus)
				json.NewEncoder(w).Encode(tt.mockBody)
			}))
			defer server.Close()

			client, _ := NewClient(&Config{
				BaseURL: server.URL,
				Auth:    AuthConfig{Token: "test-token"},
			})

			err := client.Billing.SetDefaultPaymentMethod(context.Background(), tt.paymentMethodID)

			if (err != nil) != tt.wantErr {
				t.Errorf("SetDefaultPaymentMethod() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// ============================================================================
// New Type JSON Tests (Task #3939)
// ============================================================================

// TestCreateCheckoutRequest_JSON tests JSON marshaling of CreateCheckoutRequest
func TestCreateCheckoutRequest_JSON(t *testing.T) {
	original := &CreateCheckoutRequest{
		PlanID:       "plan_pro",
		BillingCycle: "yearly",
		SuccessURL:   "https://example.com/success",
		CancelURL:    "https://example.com/cancel",
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var decoded CreateCheckoutRequest
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if decoded.PlanID != original.PlanID {
		t.Errorf("PlanID mismatch: got %s, want %s", decoded.PlanID, original.PlanID)
	}
	if decoded.BillingCycle != original.BillingCycle {
		t.Errorf("BillingCycle mismatch: got %s, want %s", decoded.BillingCycle, original.BillingCycle)
	}
}

// TestPlan_JSON tests JSON marshaling of Plan
func TestPlan_JSON(t *testing.T) {
	original := &Plan{
		ID:           "plan_pro",
		Name:         "Professional",
		Description:  "For growing teams",
		MonthlyPrice: 9900,
		YearlyPrice:  99000,
		Currency:     "USD",
		Features: []PlanFeature{
			{Name: "API Access", Included: true},
		},
		Limits: PlanLimits{
			MaxServers:        50,
			MaxUsers:          10,
			DataRetentionDays: 90,
		},
		IsPublic:  true,
		SortOrder: 2,
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var decoded Plan
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if decoded.ID != original.ID {
		t.Errorf("ID mismatch: got %s, want %s", decoded.ID, original.ID)
	}
	if decoded.MonthlyPrice != original.MonthlyPrice {
		t.Errorf("MonthlyPrice mismatch: got %d, want %d", decoded.MonthlyPrice, original.MonthlyPrice)
	}
	if decoded.Limits.MaxServers != original.Limits.MaxServers {
		t.Errorf("MaxServers mismatch: got %d, want %d", decoded.Limits.MaxServers, original.Limits.MaxServers)
	}
}

// TestSubscriptionResponse_JSON tests JSON marshaling of SubscriptionResponse
func TestSubscriptionResponse_JSON(t *testing.T) {
	original := &SubscriptionResponse{
		ID:                   "sub_123",
		OrganizationID:       1,
		PlanID:               "plan_pro",
		PlanName:             "Professional",
		Status:               "active",
		BillingCycle:         "monthly",
		CancelAtPeriodEnd:    false,
		StripeSubscriptionID: "sub_stripe_123",
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var decoded SubscriptionResponse
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if decoded.ID != original.ID {
		t.Errorf("ID mismatch: got %s, want %s", decoded.ID, original.ID)
	}
	if decoded.StripeSubscriptionID != original.StripeSubscriptionID {
		t.Errorf("StripeSubscriptionID mismatch: got %s, want %s", decoded.StripeSubscriptionID, original.StripeSubscriptionID)
	}
}
