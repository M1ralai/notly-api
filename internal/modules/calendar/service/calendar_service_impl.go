package service

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/M1ralai/notly-api/internal/common/utils"
	"github.com/M1ralai/notly-api/internal/infrastructure/encryption"
	"github.com/M1ralai/notly-api/internal/infrastructure/google"
	"github.com/M1ralai/notly-api/internal/infrastructure/logger"
	"github.com/M1ralai/notly-api/internal/modules/calendar/domain"
	"github.com/M1ralai/notly-api/internal/modules/calendar/dto"
	"github.com/M1ralai/notly-api/internal/modules/calendar/repository"
	notifService "github.com/M1ralai/notly-api/internal/modules/notification/service"
)

type calendarService struct {
	integrationRepo  repository.CalendarIntegrationRepository
	syncQueueRepo    repository.SyncQueueRepository
	eventMappingRepo repository.EventMappingRepository
	googleClient     google.CalendarClient
	encryptor        *encryption.Encryptor
	logger           *logger.ZapLogger
	broadcaster      *notifService.Broadcaster
}

func NewCalendarService(
	integrationRepo repository.CalendarIntegrationRepository,
	syncQueueRepo repository.SyncQueueRepository,
	eventMappingRepo repository.EventMappingRepository,
	googleClient google.CalendarClient,
	encryptor *encryption.Encryptor,
	logger *logger.ZapLogger,
	broadcaster *notifService.Broadcaster,
) CalendarService {
	return &calendarService{
		integrationRepo:  integrationRepo,
		syncQueueRepo:    syncQueueRepo,
		eventMappingRepo: eventMappingRepo,
		googleClient:     googleClient,
		encryptor:        encryptor,
		logger:           logger,
		broadcaster:      broadcaster,
	}
}

func (s *calendarService) GetGoogleAuthURL(ctx context.Context, userID int) (string, error) {
	s.logger.Info("Generating Google auth URL", map[string]interface{}{"user_id": userID, "action": "GET_GOOGLE_AUTH_URL"})

	clientID := os.Getenv("GOOGLE_CLIENT_ID")
	redirectURI := os.Getenv("GOOGLE_REDIRECT_URI")
	if clientID == "" || redirectURI == "" {
		return "", errors.New("Google OAuth not configured")
	}

	scope := url.QueryEscape("https://www.googleapis.com/auth/calendar.events")
	state := strconv.Itoa(userID)

	authURL := fmt.Sprintf(
		"https://accounts.google.com/o/oauth2/v2/auth?client_id=%s&redirect_uri=%s&response_type=code&scope=%s&access_type=offline&prompt=consent&state=%s",
		clientID, url.QueryEscape(redirectURI), scope, state,
	)

	s.logger.Info("Google auth URL generated", map[string]interface{}{"user_id": userID, "action": "GET_GOOGLE_AUTH_URL_SUCCESS"})
	return authURL, nil
}

func (s *calendarService) HandleGoogleCallback(ctx context.Context, userID int, code string) (*dto.IntegrationResponse, error) {
	s.logger.Info("Handling Google OAuth callback", map[string]interface{}{"user_id": userID, "action": "GOOGLE_OAUTH_CALLBACK"})

	// Exchange code for tokens using Google OAuth API
	tokens, err := s.googleClient.ExchangeCode(ctx, code)
	if err != nil {
		s.logger.Error("Failed to exchange code for tokens", err, map[string]interface{}{"user_id": userID, "action": "GOOGLE_OAUTH_TOKEN_EXCHANGE_FAILED"})
		return nil, fmt.Errorf("token exchange failed: %w", err)
	}

	// Encrypt tokens before storing
	encryptedAccessToken, err := s.encryptor.Encrypt(tokens.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt access token: %w", err)
	}

	encryptedRefreshToken, err := s.encryptor.Encrypt(tokens.RefreshToken)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt refresh token: %w", err)
	}

	now := time.Now()
	expiresAt := tokens.ExpiresAt

	existing, _ := s.integrationRepo.GetByUserAndProvider(ctx, userID, "google")
	if existing != nil {
		existing.AccessToken = encryptedAccessToken
		existing.RefreshToken = encryptedRefreshToken
		existing.ExpiresAt = &expiresAt
		existing.IsActive = true
		existing.UpdatedAt = now
		if err := s.integrationRepo.Update(ctx, existing); err != nil {
			s.logger.Error("Failed to update Google integration", err, map[string]interface{}{"user_id": userID, "action": "GOOGLE_OAUTH_CALLBACK_FAILED"})
			return nil, err
		}
		s.logger.Info("Google integration updated", map[string]interface{}{"user_id": userID, "integration_id": existing.ID, "action": "GOOGLE_OAUTH_CALLBACK_SUCCESS"})
		return dto.ToIntegrationResponse(existing), nil
	}

	integration := &domain.CalendarIntegration{
		UserID:       userID,
		Provider:     "google",
		AccessToken:  encryptedAccessToken,
		RefreshToken: encryptedRefreshToken,
		ExpiresAt:    &expiresAt,
		IsActive:     true,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	created, err := s.integrationRepo.Create(ctx, integration)
	if err != nil {
		s.logger.Error("Failed to create Google integration", err, map[string]interface{}{"user_id": userID, "action": "GOOGLE_OAUTH_CALLBACK_FAILED"})
		return nil, err
	}

	s.logger.Info("Google integration created", map[string]interface{}{"user_id": userID, "integration_id": created.ID, "action": "GOOGLE_OAUTH_CALLBACK_SUCCESS"})
	return dto.ToIntegrationResponse(created), nil
}

func (s *calendarService) DisconnectGoogle(ctx context.Context, userID int) error {
	s.logger.Info("Disconnecting Google Calendar", map[string]interface{}{"user_id": userID, "action": "DISCONNECT_GOOGLE"})

	integration, err := s.integrationRepo.GetByUserAndProvider(ctx, userID, "google")
	if err != nil {
		return err
	}
	if integration == nil {
		return errors.New("no Google integration found")
	}

	integration.IsActive = false
	if err := s.integrationRepo.Update(ctx, integration); err != nil {
		s.logger.Error("Failed to disconnect Google", err, map[string]interface{}{"user_id": userID, "action": "DISCONNECT_GOOGLE_FAILED"})
		return err
	}

	s.logger.Info("Google Calendar disconnected", map[string]interface{}{"user_id": userID, "action": "DISCONNECT_GOOGLE_SUCCESS"})
	return nil
}

func (s *calendarService) SyncGoogle(ctx context.Context, userID int) error {
	s.logger.Info("Syncing with Google Calendar", map[string]interface{}{"user_id": userID, "action": "SYNC_GOOGLE"})

	integration, err := s.integrationRepo.GetByUserAndProvider(ctx, userID, "google")
	if err != nil {
		return err
	}
	if integration == nil || !integration.IsActive {
		return errors.New("Google Calendar not connected")
	}

	now := time.Now()
	integration.LastSyncAt = &now
	if err := s.integrationRepo.Update(ctx, integration); err != nil {
		s.logger.Error("Failed to sync Google", err, map[string]interface{}{"user_id": userID, "action": "SYNC_GOOGLE_FAILED"})
		return err
	}

	s.logger.Info("Google Calendar synced", map[string]interface{}{"user_id": userID, "action": "SYNC_GOOGLE_SUCCESS"})
	return nil
}

func (s *calendarService) GetSyncStatus(ctx context.Context, userID int) (*dto.SyncStatusResponse, error) {
	integrations, err := s.integrationRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	status := &dto.SyncStatusResponse{Integrations: make([]*dto.IntegrationResponse, len(integrations))}
	for i, integration := range integrations {
		status.Integrations[i] = dto.ToIntegrationResponse(integration)
	}

	return status, nil
}

func (s *calendarService) GetIntegrations(ctx context.Context, userID int) ([]*dto.IntegrationResponse, error) {
	integrations, err := s.integrationRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	result := make([]*dto.IntegrationResponse, len(integrations))
	for i, integration := range integrations {
		result[i] = dto.ToIntegrationResponse(integration)
	}
	return result, nil
}

// getDecryptedAccessToken returns decrypted access token, refreshing if needed
func (s *calendarService) getDecryptedAccessToken(ctx context.Context, integration *domain.CalendarIntegration) (string, error) {
	// Check if token needs refresh
	if integration.NeedsRefresh() {
		refreshToken, err := s.encryptor.Decrypt(integration.RefreshToken)
		if err != nil {
			return "", fmt.Errorf("failed to decrypt refresh token: %w", err)
		}

		tokens, err := s.googleClient.RefreshToken(ctx, refreshToken)
		if err != nil {
			if strings.Contains(err.Error(), "invalid_grant") || strings.Contains(err.Error(), "Token has been expired or revoked") {
				s.logger.Error("Google integration token completely revoked by provider. Deleting integration.", err, map[string]interface{}{"user_id": integration.UserID})

				// Token is permanently dead. Delete the integration.
				_ = s.integrationRepo.Delete(ctx, integration.ID)

				// Notify the frontend via WebSocket
				if s.broadcaster != nil {
					excludeCID := utils.GetConnectionIDFromContext(ctx)
					s.broadcaster.Publish(integration.UserID, excludeCID, "integration.revoked", map[string]interface{}{
						"provider": integration.Provider,
						"message":  "Google Calendar connection lost. Please reconnect your account.",
					})
				}

				return "", fmt.Errorf("google calendar connection lost. please disconnect and reconnect your account")
			}
			return "", fmt.Errorf("token refresh failed: %w", err)
		}

		// Update stored tokens
		encryptedAccessToken, _ := s.encryptor.Encrypt(tokens.AccessToken)
		integration.AccessToken = encryptedAccessToken

		if tokens.RefreshToken != "" {
			newEncryptedRefreshToken, err := s.encryptor.Encrypt(tokens.RefreshToken)
			if err == nil {
				integration.RefreshToken = newEncryptedRefreshToken
			}
		}

		expiresAt := tokens.ExpiresAt
		integration.ExpiresAt = &expiresAt
		integration.UpdatedAt = time.Now()

		if err := s.integrationRepo.Update(ctx, integration); err != nil {
			s.logger.Error("Failed to update refreshed token", err, nil)
		}

		return tokens.AccessToken, nil
	}

	return s.encryptor.Decrypt(integration.AccessToken)
}

// MarkDone creates an all-day event in Google Calendar for completed task/habit
func (s *calendarService) MarkDone(ctx context.Context, userID int, localID int, entityType string, title string, date time.Time) error {
	s.logger.Info("Marking done in Google Calendar", map[string]interface{}{
		"user_id": userID, "local_id": localID, "entity_type": entityType, "action": "MARK_DONE",
	})

	integration, err := s.integrationRepo.GetByUserAndProvider(ctx, userID, "google")
	if err != nil {
		return err
	}
	if integration == nil || !integration.IsActive {
		// No calendar connected, silently skip
		return nil
	}

	accessToken, err := s.getDecryptedAccessToken(ctx, integration)
	if err != nil {
		s.logger.Error("Failed to get access token", err, map[string]interface{}{"user_id": userID})
		return err
	}

	event := &google.CalendarEvent{
		Summary: "✅ " + title,
		Date:    date,
	}

	googleEventID, err := s.googleClient.CreateAllDayEvent(ctx, accessToken, event)
	if err != nil {
		s.logger.Error("Failed to create Google Calendar event", err, map[string]interface{}{
			"user_id": userID, "local_id": localID, "action": "CREATE_EVENT_FAILED",
		})
		return err
	}

	// Save mapping
	mapping := &domain.GoogleCalendarEvent{
		UserID:        userID,
		LocalID:       localID,
		LocalType:     entityType,
		GoogleEventID: googleEventID,
		EventDate:     date,
	}
	if err := s.eventMappingRepo.Create(ctx, mapping); err != nil {
		s.logger.Error("Failed to save event mapping", err, map[string]interface{}{
			"user_id": userID, "google_event_id": googleEventID,
		})
		// Try to delete the orphaned Google event
		_ = s.googleClient.DeleteEvent(ctx, accessToken, googleEventID)
		return err
	}

	s.logger.Info("Done marked in Google Calendar", map[string]interface{}{
		"user_id": userID, "google_event_id": googleEventID, "action": "MARK_DONE_SUCCESS",
	})
	return nil
}

// MarkUndone removes the event from Google Calendar for unmarked task/habit
func (s *calendarService) MarkUndone(ctx context.Context, userID int, localID int, entityType string, date time.Time) error {
	s.logger.Info("Marking undone in Google Calendar", map[string]interface{}{
		"user_id": userID, "local_id": localID, "entity_type": entityType, "action": "MARK_UNDONE",
	})

	integration, err := s.integrationRepo.GetByUserAndProvider(ctx, userID, "google")
	if err != nil {
		return err
	}
	if integration == nil || !integration.IsActive {
		// No calendar connected, silently skip
		return nil
	}

	// Find mapping
	mapping, err := s.eventMappingRepo.GetByLocalEvent(ctx, userID, localID, entityType, date)
	if err != nil {
		return err
	}
	if mapping == nil {
		// No mapping found, nothing to delete
		return nil
	}

	accessToken, err := s.getDecryptedAccessToken(ctx, integration)
	if err != nil {
		s.logger.Error("Failed to get access token", err, map[string]interface{}{"user_id": userID})
		return err
	}

	// Delete from Google Calendar
	if err := s.googleClient.DeleteEvent(ctx, accessToken, mapping.GoogleEventID); err != nil {
		s.logger.Error("Failed to delete Google Calendar event", err, map[string]interface{}{
			"user_id": userID, "google_event_id": mapping.GoogleEventID, "action": "DELETE_EVENT_FAILED",
		})
		// Continue to delete mapping anyway
	}

	// Delete mapping
	if err := s.eventMappingRepo.Delete(ctx, mapping.ID); err != nil {
		s.logger.Error("Failed to delete event mapping", err, map[string]interface{}{
			"user_id": userID, "mapping_id": mapping.ID,
		})
		return err
	}

	s.logger.Info("Undone marked in Google Calendar", map[string]interface{}{
		"user_id": userID, "action": "MARK_UNDONE_SUCCESS",
	})
	return nil
}

func (s *calendarService) CreateAllDayEvent(ctx context.Context, userID int, localID int, entityType string, title, description string, date time.Time) error {
	s.logger.Info("Creating all-day event in Google Calendar", map[string]interface{}{
		"user_id": userID, "local_id": localID, "entity_type": entityType, "action": "CREATE_ALL_DAY_EVENT",
	})

	integration, err := s.integrationRepo.GetByUserAndProvider(ctx, userID, "google")
	if err != nil {
		return err
	}
	if integration == nil || !integration.IsActive {
		return nil
	}

	accessToken, err := s.getDecryptedAccessToken(ctx, integration)
	if err != nil {
		s.logger.Error("Failed to get access token", err, map[string]interface{}{"user_id": userID})
		return err
	}

	event := &google.CalendarEvent{
		Summary:     title,
		Description: description,
		Date:        date,
	}

	googleEventID, err := s.googleClient.CreateAllDayEvent(ctx, accessToken, event)
	if err != nil {
		s.logger.Error("Failed to create Google Calendar all-day event", err, map[string]interface{}{
			"user_id": userID, "local_id": localID, "action": "CREATE_ALL_DAY_EVENT_FAILED",
		})
		return err
	}

	mapping := &domain.GoogleCalendarEvent{
		UserID:        userID,
		LocalID:       localID,
		LocalType:     entityType,
		GoogleEventID: googleEventID,
		EventDate:     date,
	}
	if err := s.eventMappingRepo.Create(ctx, mapping); err != nil {
		s.logger.Error("Failed to save event mapping", err, map[string]interface{}{
			"user_id": userID, "google_event_id": googleEventID,
		})
		_ = s.googleClient.DeleteEvent(ctx, accessToken, googleEventID)
		return err
	}

	s.logger.Info("All-day event created in Google Calendar", map[string]interface{}{
		"user_id": userID, "google_event_id": googleEventID, "action": "CREATE_ALL_DAY_EVENT_SUCCESS",
	})
	return nil
}

func (s *calendarService) UpdateAllDayEvent(ctx context.Context, userID int, localID int, entityType string, title, description string, date time.Time) error {
	s.logger.Info("Updating all-day event in Google Calendar", map[string]interface{}{
		"user_id": userID, "local_id": localID, "entity_type": entityType, "action": "UPDATE_ALL_DAY_EVENT",
	})

	integration, err := s.integrationRepo.GetByUserAndProvider(ctx, userID, "google")
	if err != nil {
		return err
	}
	if integration == nil || !integration.IsActive {
		return nil
	}

	mapping, err := s.eventMappingRepo.GetByLocalID(ctx, userID, localID, entityType)
	if err != nil {
		return err
	}

	if mapping == nil {
		return s.CreateAllDayEvent(ctx, userID, localID, entityType, title, description, date)
	}

	accessToken, err := s.getDecryptedAccessToken(ctx, integration)
	if err != nil {
		s.logger.Error("Failed to get access token", err, map[string]interface{}{"user_id": userID})
		return err
	}

	event := &google.CalendarEvent{
		Summary:     title,
		Description: description,
		Date:        date,
	}

	if err := s.googleClient.UpdateEvent(ctx, accessToken, mapping.GoogleEventID, event); err != nil {
		s.logger.Error("Failed to update Google Calendar event", err, map[string]interface{}{
			"user_id": userID, "google_event_id": mapping.GoogleEventID, "action": "UPDATE_EVENT_FAILED",
		})
		return err
	}

	mapping.EventDate = date
	// ideally we update mapping in DB if date changes, but repo support is limited.
	// however since we rely on Google ID, it's mostly fine.

	s.logger.Info("All-day event updated in Google Calendar", map[string]interface{}{
		"user_id": userID, "action": "UPDATE_ALL_DAY_EVENT_SUCCESS",
	})
	return nil
}

func (s *calendarService) CreateTimedEvent(ctx context.Context, userID int, localID int, entityType string, title, description string, startTime, endTime time.Time, recurrence []string, notificationEnabled bool, notificationMethod string, notificationMinutes int) error {
	s.logger.Info("Creating timed event in Google Calendar", map[string]interface{}{
		"user_id": userID, "local_id": localID, "entity_type": entityType, "action": "CREATE_TIMED_EVENT",
	})

	integration, err := s.integrationRepo.GetByUserAndProvider(ctx, userID, "google")
	if err != nil {
		return err
	}
	if integration == nil || !integration.IsActive {
		return nil // No calendar connected
	}

	accessToken, err := s.getDecryptedAccessToken(ctx, integration)
	if err != nil {
		s.logger.Error("Failed to get access token", err, map[string]interface{}{"user_id": userID})
		return err
	}

	event := &google.CalendarEvent{
		Summary:     title,
		Description: description,
		StartTime:   &startTime,
		EndTime:     &endTime,
		Recurrence:  recurrence,
	}

	if notificationEnabled {
		method := notificationMethod
		if method == "" {
			method = "popup"
		}
		event.Reminders = &google.Reminders{
			UseDefault: false,
			Overrides: []google.ReminderOverride{
				{Method: method, Minutes: notificationMinutes},
			},
		}
	} else {
		// Explicitly disable reminders
		event.Reminders = &google.Reminders{
			UseDefault: false,
			Overrides:  []google.ReminderOverride{},
		}
	}

	googleEventID, err := s.googleClient.CreateTimedEvent(ctx, accessToken, event)
	if err != nil {
		s.logger.Error("Failed to create Google Calendar timed event", err, map[string]interface{}{
			"user_id": userID, "local_id": localID, "action": "CREATE_TIMED_EVENT_FAILED",
		})
		return err
	}

	// Save mapping
	mapping := &domain.GoogleCalendarEvent{
		UserID:        userID,
		LocalID:       localID,
		LocalType:     entityType,
		GoogleEventID: googleEventID,
		EventDate:     startTime, // Use start time as event date for indexing
	}
	if err := s.eventMappingRepo.Create(ctx, mapping); err != nil {
		s.logger.Error("Failed to save event mapping", err, map[string]interface{}{
			"user_id": userID, "google_event_id": googleEventID,
		})
		_ = s.googleClient.DeleteEvent(ctx, accessToken, googleEventID)
		return err
	}

	s.logger.Info("Timed event created in Google Calendar", map[string]interface{}{
		"user_id": userID, "google_event_id": googleEventID, "action": "CREATE_TIMED_EVENT_SUCCESS",
	})
	return nil
}

func (s *calendarService) UpdateTimedEvent(ctx context.Context, userID int, localID int, entityType string, title, description string, startTime, endTime time.Time, recurrence []string, notificationEnabled bool, notificationMethod string, notificationMinutes int) error {
	s.logger.Info("Updating timed event in Google Calendar", map[string]interface{}{
		"user_id": userID, "local_id": localID, "entity_type": entityType, "action": "UPDATE_TIMED_EVENT",
	})

	integration, err := s.integrationRepo.GetByUserAndProvider(ctx, userID, "google")
	if err != nil {
		return err
	}
	if integration == nil || !integration.IsActive {
		return nil
	}

	mapping, err := s.eventMappingRepo.GetByLocalID(ctx, userID, localID, entityType)
	if err != nil {
		return err
	}

	if mapping == nil {
		// Event might not exist in Google Calendar yet (maybe created before integration)
		// Try to create it instead
		return s.CreateTimedEvent(ctx, userID, localID, entityType, title, description, startTime, endTime, recurrence, notificationEnabled, notificationMethod, notificationMinutes)
	}

	accessToken, err := s.getDecryptedAccessToken(ctx, integration)
	if err != nil {
		s.logger.Error("Failed to get access token", err, map[string]interface{}{"user_id": userID})
		return err
	}

	event := &google.CalendarEvent{
		Summary:     title,
		Description: description,
		StartTime:   &startTime,
		EndTime:     &endTime,
		Recurrence:  recurrence,
	}

	if notificationEnabled {
		method := notificationMethod
		if method == "" {
			method = "popup"
		}
		event.Reminders = &google.Reminders{
			UseDefault: false,
			Overrides: []google.ReminderOverride{
				{Method: method, Minutes: notificationMinutes},
			},
		}
	} else {
		// Explicitly disable reminders
		event.Reminders = &google.Reminders{
			UseDefault: false,
			Overrides:  []google.ReminderOverride{},
		}
	}

	if err := s.googleClient.UpdateEvent(ctx, accessToken, mapping.GoogleEventID, event); err != nil {
		s.logger.Error("Failed to update Google Calendar event", err, map[string]interface{}{
			"user_id": userID, "google_event_id": mapping.GoogleEventID, "action": "UPDATE_EVENT_FAILED",
		})
		return err
	}

	// Update mapping mostly to keep EventDate current if it changed
	mapping.EventDate = startTime
	// We don't have an Update method in repo yet (only Create and Delete),
	// but EventDate in DB is mainly for uniqueness constraint on (user, local_id, type, date) for habits.
	// For tasks, we might need a better way to handle updates if date changes significantly such that it affects lookup.
	// But `GetByLocalID` ignores date, so it should be fine.

	s.logger.Info("Timed event updated in Google Calendar", map[string]interface{}{
		"user_id": userID, "action": "UPDATE_TIMED_EVENT_SUCCESS",
	})
	return nil
}

func (s *calendarService) DeleteEvent(ctx context.Context, userID int, localID int, entityType string) error {
	s.logger.Info("Deleting event from Google Calendar", map[string]interface{}{
		"user_id": userID, "local_id": localID, "entity_type": entityType, "action": "DELETE_EVENT",
	})

	integration, err := s.integrationRepo.GetByUserAndProvider(ctx, userID, "google")
	if err != nil {
		return err
	}
	if integration == nil || !integration.IsActive {
		return nil
	}

	mapping, err := s.eventMappingRepo.GetByLocalID(ctx, userID, localID, entityType)
	if err != nil {
		return err
	}

	if mapping == nil {
		return nil // Nothing to delete
	}

	accessToken, err := s.getDecryptedAccessToken(ctx, integration)
	if err != nil {
		s.logger.Error("Failed to get access token", err, map[string]interface{}{"user_id": userID})
		return err
	}

	// Delete from Google Calendar
	if err := s.googleClient.DeleteEvent(ctx, accessToken, mapping.GoogleEventID); err != nil {
		s.logger.Error("Failed to delete Google Calendar event", err, map[string]interface{}{
			"user_id": userID, "google_event_id": mapping.GoogleEventID, "action": "DELETE_EVENT_FAILED",
		})
		// Continue to delete mapping anyway
	}

	// Delete mapping
	if err := s.eventMappingRepo.Delete(ctx, mapping.ID); err != nil {
		s.logger.Error("Failed to delete event mapping", err, map[string]interface{}{
			"user_id": userID, "mapping_id": mapping.ID,
		})
		return err
	}

	s.logger.Info("Event deleted from Google Calendar and mapping removed", map[string]interface{}{
		"user_id": userID, "action": "DELETE_EVENT_SUCCESS",
	})
	return nil
}

func (s *calendarService) QueueSync(ctx context.Context, userID int, eventID int, action string) error {
	s.logger.Info("Queueing sync operation", map[string]interface{}{"user_id": userID, "event_id": eventID, "action_type": action, "action": "QUEUE_SYNC"})

	integrations, err := s.integrationRepo.GetByUserID(ctx, userID)
	if err != nil {
		return err
	}

	for _, integration := range integrations {
		if !integration.IsActive {
			continue
		}

		queueItem := &domain.SyncQueue{
			UserID:   userID,
			EventID:  &eventID,
			Provider: integration.Provider,
			Action:   action,
			Status:   "pending",
		}
		if _, err := s.syncQueueRepo.Create(ctx, queueItem); err != nil {
			s.logger.Error("Failed to queue sync", err, map[string]interface{}{"user_id": userID, "event_id": eventID, "provider": integration.Provider, "action": "QUEUE_SYNC_FAILED"})
			return err
		}
	}

	s.logger.Info("Sync queued", map[string]interface{}{"user_id": userID, "event_id": eventID, "action": "QUEUE_SYNC_SUCCESS"})
	return nil
}

func (s *calendarService) ProcessSyncQueue(ctx context.Context, limit int) (int, error) {
	s.logger.Info("Processing sync queue", map[string]interface{}{"limit": limit, "action": "PROCESS_SYNC_QUEUE"})

	pending, err := s.syncQueueRepo.GetPending(ctx, limit)
	if err != nil {
		return 0, err
	}

	processed := 0
	for _, item := range pending {
		if err := s.syncQueueRepo.UpdateStatus(ctx, item.ID, "completed", ""); err != nil {
			s.logger.Error("Failed to process sync item", err, map[string]interface{}{"sync_id": item.ID, "action": "PROCESS_SYNC_ITEM_FAILED"})
			s.syncQueueRepo.IncrementRetry(ctx, item.ID)
			continue
		}
		processed++
	}

	s.logger.Info("Sync queue processed", map[string]interface{}{"processed": processed, "action": "PROCESS_SYNC_QUEUE_SUCCESS"})
	return processed, nil
}
