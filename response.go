package nexmonyx

import (
	"fmt"
	"time"
)

// StandardResponse represents a standard API response from the Nexmonyx API
type StandardResponse struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Details string      `json:"details,omitempty"`
}

// PaginatedResponse represents a paginated API response
type PaginatedResponse struct {
	Status  string          `json:"status"`
	Message string          `json:"message"`
	Data    interface{}     `json:"data"`
	Meta    *PaginationMeta `json:"meta"`
}

// PaginationMeta contains pagination metadata
type PaginationMeta struct {
	Page         int    `json:"page"`
	Limit        int    `json:"limit"`
	TotalItems   int    `json:"total_items"`
	TotalPages   int    `json:"total_pages"`
	HasMore      bool   `json:"has_more"`
	NextPage     *int   `json:"next_page,omitempty"`
	PrevPage     *int   `json:"prev_page,omitempty"`
	FirstPage    int    `json:"first_page"`
	LastPage     int    `json:"last_page"`
	From         int    `json:"from"`
	To           int    `json:"to"`
	PerPage      int    `json:"per_page"`
	CurrentPage  int    `json:"current_page"`
	LastPageURL  string `json:"last_page_url,omitempty"`
	NextPageURL  string `json:"next_page_url,omitempty"`
	PrevPageURL  string `json:"prev_page_url,omitempty"`
	FirstPageURL string `json:"first_page_url,omitempty"`
}

// ListOptions specifies options for listing resources
type ListOptions struct {
	Page         int               `url:"page,omitempty"`
	Limit        int               `url:"limit,omitempty"`
	PerPage      int               `url:"per_page,omitempty"`
	Sort         string            `url:"sort,omitempty"`
	Order        string            `url:"order,omitempty"`
	Search       string            `url:"search,omitempty"`
	Query        string            `url:"q,omitempty"`
	Filters      map[string]string `url:"-"`
	Fields       []string          `url:"fields,omitempty,comma"`
	Expand       []string          `url:"expand,omitempty,comma"`
	Include      []string          `url:"include,omitempty,comma"`
	StartDate    string            `url:"start_date,omitempty"`
	EndDate      string            `url:"end_date,omitempty"`
	TimeRange    string            `url:"time_range,omitempty"`
	GroupBy      string            `url:"group_by,omitempty"`
	Aggregation  string            `url:"aggregation,omitempty"`
}

// ToQuery converts ListOptions to query parameters
func (lo *ListOptions) ToQuery() map[string]string {
	params := make(map[string]string)

	if lo.Page > 0 {
		params["page"] = fmt.Sprintf("%d", lo.Page)
	}
	if lo.Limit > 0 {
		params["limit"] = fmt.Sprintf("%d", lo.Limit)
	}
	if lo.PerPage > 0 {
		params["per_page"] = fmt.Sprintf("%d", lo.PerPage)
	}
	if lo.Sort != "" {
		params["sort"] = lo.Sort
	}
	if lo.Order != "" {
		params["order"] = lo.Order
	}
	if lo.Search != "" {
		params["search"] = lo.Search
	}
	if lo.Query != "" {
		params["q"] = lo.Query
	}
	if lo.StartDate != "" {
		params["start_date"] = lo.StartDate
	}
	if lo.EndDate != "" {
		params["end_date"] = lo.EndDate
	}
	if lo.TimeRange != "" {
		params["time_range"] = lo.TimeRange
	}
	if lo.GroupBy != "" {
		params["group_by"] = lo.GroupBy
	}
	if lo.Aggregation != "" {
		params["aggregation"] = lo.Aggregation
	}

	// Add custom filters
	for k, v := range lo.Filters {
		params[k] = v
	}

	return params
}

// QueryTimeRange represents a time range for queries
type QueryTimeRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

// ToStrings converts QueryTimeRange to start/end date strings
func (qtr *QueryTimeRange) ToStrings() (start, end string) {
	return qtr.Start.Format(time.RFC3339), qtr.End.Format(time.RFC3339)
}

// Last24Hours returns a QueryTimeRange for the last 24 hours
func Last24Hours() *QueryTimeRange {
	now := time.Now()
	return &QueryTimeRange{
		Start: now.Add(-24 * time.Hour),
		End:   now,
	}
}

// Last7Days returns a QueryTimeRange for the last 7 days
func Last7Days() *QueryTimeRange {
	now := time.Now()
	return &QueryTimeRange{
		Start: now.Add(-7 * 24 * time.Hour),
		End:   now,
	}
}

// Last30Days returns a QueryTimeRange for the last 30 days
func Last30Days() *QueryTimeRange {
	now := time.Now()
	return &QueryTimeRange{
		Start: now.Add(-30 * 24 * time.Hour),
		End:   now,
	}
}

// ThisMonth returns a QueryTimeRange for the current month
func ThisMonth() *QueryTimeRange {
	now := time.Now()
	start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	return &QueryTimeRange{
		Start: start,
		End:   now,
	}
}

// LastMonth returns a QueryTimeRange for the previous month
func LastMonth() *QueryTimeRange {
	now := time.Now()
	start := time.Date(now.Year(), now.Month()-1, 1, 0, 0, 0, 0, now.Location())
	end := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()).Add(-time.Second)
	return &QueryTimeRange{
		Start: start,
		End:   end,
	}
}

// TimeRange represents a time range for API requests
type TimeRange struct {
	Start string `url:"start,omitempty" json:"start,omitempty"`
	End   string `url:"end,omitempty" json:"end,omitempty"`
}

// BatchResponse represents a response from a batch operation
type BatchResponse struct {
	Status     string               `json:"status"`
	Message    string               `json:"message"`
	Successful []BatchItemResponse  `json:"successful,omitempty"`
	Failed     []BatchItemResponse  `json:"failed,omitempty"`
	Total      int                  `json:"total"`
	Success    int                  `json:"success"`
	Failures   int                  `json:"failures"`
}

// BatchItemResponse represents an individual item response in a batch
type BatchItemResponse struct {
	ID      string      `json:"id"`
	Status  string      `json:"status"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Message string      `json:"message,omitempty"`
}

// StatusResponse represents a status check response
type StatusResponse struct {
	Status      string                 `json:"status"`
	Healthy     bool                   `json:"healthy"`
	Version     string                 `json:"version,omitempty"`
	Uptime      int64                  `json:"uptime,omitempty"`
	Timestamp   time.Time              `json:"timestamp"`
	Services    map[string]bool        `json:"services,omitempty"`
	Details     map[string]interface{} `json:"details,omitempty"`
}

// ErrorResponse represents an error response from the API
type ErrorResponse struct {
	Status    string              `json:"status"`
	Error     string              `json:"error"`
	Message   string              `json:"message"`
	Details   string              `json:"details,omitempty"`
	RequestID string              `json:"request_id,omitempty"`
	Errors    map[string][]string `json:"errors,omitempty"`
}

// IsSuccess returns true if the response indicates a successful operation
func (r *StandardResponse) IsSuccess() bool {
	return r.Status == "success"
}

// GetError returns an error if the response indicates a failure
func (r *StandardResponse) GetError() error {
	if r.Status == "error" {
		return &APIError{
			Status:    r.Status,
			ErrorType: r.Error,
			Message:   r.Message,
			Details:   r.Details,
		}
	}
	return nil
}