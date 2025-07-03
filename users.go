package nexmonyx

import (
	"context"
	"fmt"
)

// GetUser retrieves a user by ID
func (s *UsersService) Get(ctx context.Context, id string) (*User, error) {
	var resp StandardResponse
	resp.Data = &User{}

	_, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   fmt.Sprintf("/api/v1/users/%s", id),
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if user, ok := resp.Data.(*User); ok {
		return user, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// GetUserByEmail retrieves a user by email
func (s *UsersService) GetByEmail(ctx context.Context, email string) (*User, error) {
	var resp StandardResponse
	resp.Data = &User{}

	_, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   fmt.Sprintf("/api/v1/users/email/%s", email),
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if user, ok := resp.Data.(*User); ok {
		return user, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// ListUsers retrieves a list of users
func (s *UsersService) List(ctx context.Context, opts *ListOptions) ([]*User, *PaginationMeta, error) {
	var resp PaginatedResponse
	var users []*User
	resp.Data = &users

	req := &Request{
		Method: "GET",
		Path:   "/api/v1/users",
		Result: &resp,
	}

	if opts != nil {
		req.Query = opts.ToQuery()
	}

	_, err := s.client.Do(ctx, req)
	if err != nil {
		return nil, nil, err
	}

	return users, resp.Meta, nil
}

// CreateUser creates a new user
func (s *UsersService) Create(ctx context.Context, user *User) (*User, error) {
	var resp StandardResponse
	resp.Data = &User{}

	_, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   "/api/v1/users",
		Body:   user,
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if created, ok := resp.Data.(*User); ok {
		return created, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// UpdateUser updates an existing user
func (s *UsersService) Update(ctx context.Context, id string, user *User) (*User, error) {
	var resp StandardResponse
	resp.Data = &User{}

	_, err := s.client.Do(ctx, &Request{
		Method: "PUT",
		Path:   fmt.Sprintf("/api/v1/users/%s", id),
		Body:   user,
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if updated, ok := resp.Data.(*User); ok {
		return updated, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// DeleteUser deletes a user
func (s *UsersService) Delete(ctx context.Context, id string) error {
	var resp StandardResponse

	_, err := s.client.Do(ctx, &Request{
		Method: "DELETE",
		Path:   fmt.Sprintf("/api/v1/users/%s", id),
		Result: &resp,
	})
	return err
}

// GetCurrentUser retrieves the currently authenticated user
func (s *UsersService) GetCurrent(ctx context.Context) (*User, error) {
	var resp StandardResponse
	resp.Data = &User{}

	_, err := s.client.Do(ctx, &Request{
		Method: "GET",
		Path:   "/api/v1/users/me",
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if user, ok := resp.Data.(*User); ok {
		return user, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// UpdateUserRole updates a user's role
func (s *UsersService) UpdateRole(ctx context.Context, id string, role string) (*User, error) {
	var resp StandardResponse
	resp.Data = &User{}

	body := map[string]interface{}{
		"role": role,
	}

	_, err := s.client.Do(ctx, &Request{
		Method: "PUT",
		Path:   fmt.Sprintf("/api/v1/users/%s/role", id),
		Body:   body,
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if user, ok := resp.Data.(*User); ok {
		return user, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// UpdateUserPermissions updates a user's permissions
func (s *UsersService) UpdatePermissions(ctx context.Context, id string, permissions []string) (*User, error) {
	var resp StandardResponse
	resp.Data = &User{}

	body := map[string]interface{}{
		"permissions": permissions,
	}

	_, err := s.client.Do(ctx, &Request{
		Method: "PUT",
		Path:   fmt.Sprintf("/api/v1/users/%s/permissions", id),
		Body:   body,
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if user, ok := resp.Data.(*User); ok {
		return user, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// UpdateUserPreferences updates a user's preferences
func (s *UsersService) UpdatePreferences(ctx context.Context, id string, preferences map[string]interface{}) (*User, error) {
	var resp StandardResponse
	resp.Data = &User{}

	_, err := s.client.Do(ctx, &Request{
		Method: "PUT",
		Path:   fmt.Sprintf("/api/v1/users/%s/preferences", id),
		Body:   preferences,
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if user, ok := resp.Data.(*User); ok {
		return user, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// ResetUserPassword sends a password reset email
func (s *UsersService) ResetPassword(ctx context.Context, email string) error {
	var resp StandardResponse

	body := map[string]interface{}{
		"email": email,
	}

	_, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   "/api/v1/users/reset-password",
		Body:   body,
		Result: &resp,
	})
	return err
}

// EnableUser enables a user account
func (s *UsersService) Enable(ctx context.Context, id string) (*User, error) {
	var resp StandardResponse
	resp.Data = &User{}

	_, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   fmt.Sprintf("/api/v1/users/%s/enable", id),
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if user, ok := resp.Data.(*User); ok {
		return user, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}

// DisableUser disables a user account
func (s *UsersService) Disable(ctx context.Context, id string) (*User, error) {
	var resp StandardResponse
	resp.Data = &User{}

	_, err := s.client.Do(ctx, &Request{
		Method: "POST",
		Path:   fmt.Sprintf("/api/v1/users/%s/disable", id),
		Result: &resp,
	})
	if err != nil {
		return nil, err
	}

	if user, ok := resp.Data.(*User); ok {
		return user, nil
	}
	return nil, fmt.Errorf("unexpected response type")
}
