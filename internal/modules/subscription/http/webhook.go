package http

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/M1ralai/notly-api/internal/common/utils"
	"github.com/M1ralai/notly-api/internal/modules/subscription/domain"
	"github.com/M1ralai/notly-api/internal/modules/subscription/service"
)

type RevenueCatWebhookEvent struct {
	ID                    string   `json:"id"`
	Type                  string   `json:"type"`
	AppUserID             string   `json:"app_user_id"`
	OriginalAppUserID     string   `json:"original_app_user_id"`
	Aliases               []string `json:"aliases,omitempty"`
	ProductID             string   `json:"product_id"`
	EntitlementID         string   `json:"entitlement_id,omitempty"`
	EntitlementIDs        []string `json:"entitlement_ids,omitempty"`
	PeriodType            string   `json:"period_type,omitempty"`
	PurchasedAtMs         int64    `json:"purchased_at_ms,omitempty"`
	ExpirationAtMs        int64    `json:"expiration_at_ms,omitempty"`
	Environment           string   `json:"environment,omitempty"`
	Store                 string   `json:"store,omitempty"`
	TransactionID         string   `json:"transaction_id,omitempty"`
	OriginalTransactionID string   `json:"original_transaction_id,omitempty"`
	CancelReason          string   `json:"cancel_reason,omitempty"`
}

type RevenueCatWebhookPayload struct {
	APIVersion string                  `json:"api_version"`
	Event      RevenueCatWebhookEvent `json:"event"`
}

type WebhookHandler struct {
	service service.Service
}

func NewWebhookHandler(service service.Service) *WebhookHandler {
	return &WebhookHandler{service: service}
}

// HandleRevenueCat processes incoming webhook events from RevenueCat.
func (wh *WebhookHandler) HandleRevenueCat(w http.ResponseWriter, r *http.Request) {
	// Verify Authorization header if configured
	secret := os.Getenv("REVENUECAT_WEBHOOK_AUTH_HEADER")
	if secret == "" {
		secret = os.Getenv("REVENUECAT_WEBHOOK_SECRET")
	}
	if secret != "" {
		authHeader := r.Header.Get("Authorization")
		if authHeader != secret && authHeader != "Bearer "+secret {
			log.Printf("⚠ RevenueCat webhook unauthorized attempt: header=%s", authHeader)
			utils.ReturnError(w, "UNAUTHORIZED", "Webhook authorization failed", "")
			return
		}
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		utils.ReturnError(w, "BAD_REQUEST", "Failed to read request body", err.Error())
		return
	}
	defer r.Body.Close()

	var payload RevenueCatWebhookPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		log.Printf("⚠ Failed to unmarshal RevenueCat webhook: %v", err)
		utils.ReturnError(w, "INVALID_BODY", "Invalid JSON payload", err.Error())
		return
	}

	event := payload.Event
	log.Printf("📥 RevenueCat Webhook received: type=%s app_user_id=%s product_id=%s store=%s",
		event.Type, event.AppUserID, event.ProductID, event.Store)

	// Try to resolve user ID from app_user_id
	userID, err := parseUserID(event.AppUserID)
	if err != nil && event.OriginalAppUserID != "" {
		userID, err = parseUserID(event.OriginalAppUserID)
	}

	if err != nil || userID <= 0 {
		// Log and respond 200 OK so RevenueCat does not keep retrying unrecognized anonymous events
		log.Printf("ℹ Ignoring RevenueCat event for non-numeric app_user_id=%s (err: %v)", event.AppUserID, err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ignored", "reason": "unrecognized user id"})
		return
	}

	provider := mapStoreToProvider(event.Store)
	status := mapEventTypeToStatus(event.Type)
	plan := mapProductIDToPlan(event.ProductID)

	var expiresAt *time.Time
	if event.ExpirationAtMs > 0 {
		t := time.UnixMilli(event.ExpirationAtMs)
		expiresAt = &t
	}

	env := strings.ToLower(event.Environment)
	if env == "sandbox" {
		env = "sandbox"
	} else if env == "" {
		env = "production"
	}

	input := service.UpsertEntitlementInput{
		UserID:                userID,
		Provider:              provider,
		ProductID:             event.ProductID,
		Plan:                  plan,
		Status:                status,
		TransactionID:         event.TransactionID,
		OriginalTransactionID: event.OriginalTransactionID,
		ExpiresAt:             expiresAt,
		Environment:           env,
		RawPayload:            body,
	}

	res, err := wh.service.UpsertEntitlement(r.Context(), input)
	if err != nil {
		log.Printf("❌ Failed to upsert entitlement for user %d: %v", userID, err)
		utils.ReturnError(w, "INTERNAL_ERROR", "Failed to process entitlement update", err.Error())
		return
	}

	log.Printf("✓ Entitlement updated for user %d: is_premium=%v status=%s plan=%s",
		userID, res.IsPremium, res.Status, res.PremiumPlan)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":     "ok",
		"is_premium": res.IsPremium,
		"plan":       res.PremiumPlan,
	})
}

func parseUserID(appUserID string) (int, error) {
	clean := strings.TrimSpace(appUserID)
	// If prefixed like "user_123" or similar, strip prefix
	if idx := strings.LastIndex(clean, "_"); idx != -1 {
		part := clean[idx+1:]
		if id, err := strconv.Atoi(part); err == nil {
			return id, nil
		}
	}
	return strconv.Atoi(clean)
}

func mapStoreToProvider(store string) string {
	switch strings.ToUpper(strings.TrimSpace(store)) {
	case "APP_STORE", "MAC_APP_STORE":
		return domain.ProviderApple
	case "PLAY_STORE":
		return domain.ProviderGoogle
	case "STRIPE":
		return domain.ProviderStripe
	case "RC_BILLING":
		return domain.ProviderRCBilling
	case "PROMOTIONAL":
		return "promotional"
	default:
		return domain.ProviderRevenueCat
	}
}

func mapEventTypeToStatus(eventType string) string {
	switch strings.ToUpper(strings.TrimSpace(eventType)) {
	case "INITIAL_PURCHASE", "RENEWAL", "UNCANCELLATION", "PRODUCT_CHANGE", "NON_RENEWING_PURCHASE":
		return domain.StatusActive
	case "CANCELLATION":
		return domain.StatusCancelled
	case "EXPIRATION":
		return domain.StatusExpired
	case "BILLING_ISSUE":
		return domain.StatusBillingIssue
	case "REVOCATION":
		return domain.StatusRevoked
	default:
		return domain.StatusActive
	}
}

func mapProductIDToPlan(productID string) string {
	lower := strings.ToLower(productID)
	if strings.Contains(lower, "year") || strings.Contains(lower, "annual") {
		return domain.PlanYearly
	}
	if strings.Contains(lower, "life") {
		return domain.PlanLifetime
	}
	return domain.PlanMonthly
}
