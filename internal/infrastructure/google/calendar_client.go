package google

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"
)

// Tokens represents OAuth tokens from Google
type Tokens struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token,omitempty"`
	TokenType    string    `json:"token_type"`
	ExpiresIn    int       `json:"expires_in"`
	ExpiresAt    time.Time `json:"-"`
}

// ReminderOverride represents a custom reminder override for Google Calendar
type ReminderOverride struct {
	Method  string `json:"method"` // "email" or "popup"
	Minutes int    `json:"minutes"`
}

// Reminders represents the reminders object for Google Calendar API
type Reminders struct {
	UseDefault bool               `json:"useDefault"`
	Overrides  []ReminderOverride `json:"overrides,omitempty"`
}

// CalendarEvent represents an event to create in Google Calendar
type CalendarEvent struct {
	Summary     string
	Description string
	Date        time.Time  // For all-day events
	StartTime   *time.Time // For timed events
	EndTime     *time.Time // For timed events
	Recurrence  []string   // RRULE strings
	Reminders   *Reminders // Custom reminders
}

// CalendarClient interface for Google Calendar operations
type CalendarClient interface {
	ExchangeCode(ctx context.Context, code string) (*Tokens, error)
	RefreshToken(ctx context.Context, refreshToken string) (*Tokens, error)
	CreateAllDayEvent(ctx context.Context, accessToken string, event *CalendarEvent) (string, error)
	CreateTimedEvent(ctx context.Context, accessToken string, event *CalendarEvent) (string, error)
	UpdateEvent(ctx context.Context, accessToken string, eventID string, event *CalendarEvent) error
	DeleteEvent(ctx context.Context, accessToken string, eventID string) error
}

type calendarClient struct {
	clientID     string
	clientSecret string
	redirectURI  string
	httpClient   *http.Client
}

// NewCalendarClient creates a new Google Calendar client
func NewCalendarClient() (CalendarClient, error) {
	clientID := os.Getenv("GOOGLE_CLIENT_ID")
	clientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")
	redirectURI := os.Getenv("GOOGLE_REDIRECT_URI")

	if clientID == "" || clientSecret == "" || redirectURI == "" {
		return nil, errors.New("Google OAuth environment variables not configured")
	}

	return &calendarClient{
		clientID:     clientID,
		clientSecret: clientSecret,
		redirectURI:  redirectURI,
		httpClient:   &http.Client{Timeout: 30 * time.Second},
	}, nil
}

// ExchangeCode exchanges authorization code for access and refresh tokens
func (c *calendarClient) ExchangeCode(ctx context.Context, code string) (*Tokens, error) {
	data := url.Values{
		"code":          {code},
		"client_id":     {c.clientID},
		"client_secret": {c.clientSecret},
		"redirect_uri":  {c.redirectURI},
		"grant_type":    {"authorization_code"},
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://oauth2.googleapis.com/token", bytes.NewBufferString(data.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("token exchange failed: %s", string(body))
	}

	var tokens Tokens
	if err := json.NewDecoder(resp.Body).Decode(&tokens); err != nil {
		return nil, err
	}

	tokens.ExpiresAt = time.Now().Add(time.Duration(tokens.ExpiresIn) * time.Second)
	return &tokens, nil
}

// RefreshToken refreshes an expired access token
func (c *calendarClient) RefreshToken(ctx context.Context, refreshToken string) (*Tokens, error) {
	data := url.Values{
		"refresh_token": {refreshToken},
		"client_id":     {c.clientID},
		"client_secret": {c.clientSecret},
		"grant_type":    {"refresh_token"},
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://oauth2.googleapis.com/token", bytes.NewBufferString(data.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("token refresh failed: %s", string(body))
	}

	var tokens Tokens
	if err := json.NewDecoder(resp.Body).Decode(&tokens); err != nil {
		return nil, err
	}

	tokens.ExpiresAt = time.Now().Add(time.Duration(tokens.ExpiresIn) * time.Second)
	return &tokens, nil
}

// CreateAllDayEvent creates an all-day event in the user's primary calendar
func (c *calendarClient) CreateAllDayEvent(ctx context.Context, accessToken string, event *CalendarEvent) (string, error) {
	// Format date for all-day event (YYYY-MM-DD)
	dateStr := event.Date.Format("2006-01-02")
	nextDay := event.Date.AddDate(0, 0, 1).Format("2006-01-02")

	payload := map[string]interface{}{
		"summary":     event.Summary,
		"description": event.Description,
		"start": map[string]string{
			"date": dateStr,
		},
		"end": map[string]string{
			"date": nextDay,
		},
	}

	if event.Reminders != nil {
		payload["reminders"] = event.Reminders
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://www.googleapis.com/calendar/v3/calendars/primary/events", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("create event failed: %s", string(respBody))
	}

	var result struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	return result.ID, nil
}

// DeleteEvent deletes an event from the user's primary calendar
func (c *calendarClient) DeleteEvent(ctx context.Context, accessToken string, eventID string) error {
	req, err := http.NewRequestWithContext(ctx, "DELETE", "https://www.googleapis.com/calendar/v3/calendars/primary/events/"+eventID, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// 204 No Content = success, 404 = already deleted (ok), 410 = gone (ok)
	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusNotFound && resp.StatusCode != http.StatusGone {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("delete event failed: %s", string(body))
	}

	return nil
}

// CreateTimedEvent creates a timed event in the user's primary calendar
func (c *calendarClient) CreateTimedEvent(ctx context.Context, accessToken string, event *CalendarEvent) (string, error) {
	if event.StartTime == nil || event.EndTime == nil {
		return "", errors.New("start and end times are required for timed events")
	}

	payload := map[string]interface{}{
		"summary":     event.Summary,
		"description": event.Description,
		"start": map[string]string{
			"dateTime": event.StartTime.Format(time.RFC3339),
			"timeZone": event.StartTime.Location().String(),
		},
		"end": map[string]string{
			"dateTime": event.EndTime.Format(time.RFC3339),
			"timeZone": event.EndTime.Location().String(),
		},
	}

	if len(event.Recurrence) > 0 {
		payload["recurrence"] = event.Recurrence
	}

	if event.Reminders != nil {
		payload["reminders"] = event.Reminders
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://www.googleapis.com/calendar/v3/calendars/primary/events", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("create timed event failed: %s", string(respBody))
	}

	var result struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	return result.ID, nil
}

// UpdateEvent updates an existing event in the user's primary calendar
func (c *calendarClient) UpdateEvent(ctx context.Context, accessToken string, eventID string, event *CalendarEvent) error {
	payload := map[string]interface{}{
		"summary":     event.Summary,
		"description": event.Description,
	}

	if event.Date.IsZero() && event.StartTime != nil && event.EndTime != nil {
		// Timed event
		payload["start"] = map[string]string{
			"dateTime": event.StartTime.Format(time.RFC3339),
			"timeZone": event.StartTime.Location().String(),
		}
		payload["end"] = map[string]string{
			"dateTime": event.EndTime.Format(time.RFC3339),
			"timeZone": event.EndTime.Location().String(),
		}
	} else if !event.Date.IsZero() {
		// All-day event
		dateStr := event.Date.Format("2006-01-02")
		nextDay := event.Date.AddDate(0, 0, 1).Format("2006-01-02")
		payload["start"] = map[string]string{"date": dateStr}
		payload["end"] = map[string]string{"date": nextDay}
	}

	if event.Reminders != nil {
		payload["reminders"] = event.Reminders
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "PATCH", "https://www.googleapis.com/calendar/v3/calendars/primary/events/"+eventID, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("update event failed: %s", string(respBody))
	}

	return nil
}
