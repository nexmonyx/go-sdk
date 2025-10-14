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
