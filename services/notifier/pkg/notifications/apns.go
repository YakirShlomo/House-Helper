package notifications

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/sideshow/apns2"
	"github.com/sideshow/apns2/certificate"
	"github.com/sideshow/apns2/payload"
	"github.com/sideshow/apns2/token"
)

// APNSService handles Apple Push Notification Service
type APNSService struct {
	client *apns2.Client
}

// NewAPNSServiceWithCertificate creates a new APNS service with certificate
func NewAPNSServiceWithCertificate(certPath, keyPath string, production bool) (*APNSService, error) {
	cert, err := certificate.FromPemFile(certPath, keyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load certificate: %w", err)
	}

	client := apns2.NewClient(cert)
	if production {
		client = client.Production()
	} else {
		client = client.Development()
	}

	return &APNSService{client: client}, nil
}

// NewAPNSServiceWithToken creates a new APNS service with auth token
func NewAPNSServiceWithToken(keyPath, keyID, teamID string, production bool) (*APNSService, error) {
	authKey, err := token.AuthKeyFromFile(keyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load auth key: %w", err)
	}

	tokenAuth := &token.Token{
		AuthKey: authKey,
		KeyID:   keyID,
		TeamID:  teamID,
	}

	client := apns2.NewTokenClient(tokenAuth)
	if production {
		client = client.Production()
	} else {
		client = client.Development()
	}

	return &APNSService{client: client}, nil
}

// APNSPayload represents an APNS notification payload
type APNSPayload struct {
	Title           string            `json:"title"`
	Body            string            `json:"body"`
	Badge           *int              `json:"badge,omitempty"`
	Sound           string            `json:"sound,omitempty"`
	Category        string            `json:"category,omitempty"`
	ThreadID        string            `json:"threadId,omitempty"`
	CustomData      map[string]string `json:"customData,omitempty"`
	MutableContent  bool              `json:"mutableContent,omitempty"`
	ContentState    map[string]string `json:"contentState,omitempty"`
	TargetContentID string            `json:"targetContentId,omitempty"`
}

// SendNotification sends a notification to a device token
func (a *APNSService) SendNotification(ctx context.Context, bundleID, deviceToken string, apnsPayload APNSPayload) (*apns2.Response, error) {
	notification := &apns2.Notification{
		DeviceToken: deviceToken,
		Topic:       bundleID,
	}

	// Build the payload
	p := payload.NewPayload()
	if apnsPayload.Title != "" {
		p.Alert(apnsPayload.Title)
	}
	if apnsPayload.Body != "" {
		p.AlertBody(apnsPayload.Body)
	}
	if apnsPayload.Badge != nil {
		p.Badge(*apnsPayload.Badge)
	}
	if apnsPayload.Sound != "" {
		p.Sound(apnsPayload.Sound)
	}
	if apnsPayload.Category != "" {
		p.Category(apnsPayload.Category)
	}
	if apnsPayload.ThreadID != "" {
		p.ThreadID(apnsPayload.ThreadID)
	}
	if apnsPayload.MutableContent {
		p.MutableContent()
	}
	if apnsPayload.TargetContentID != "" {
		p.Custom("target-content-id", apnsPayload.TargetContentID)
	}

	// Add custom data
	for key, value := range apnsPayload.CustomData {
		p.Custom(key, value)
	}

	// Add content state for live activities
	if len(apnsPayload.ContentState) > 0 {
		p.Custom("content-state", apnsPayload.ContentState)
	}

	notification.Payload = p

	response, err := a.client.PushWithContext(ctx, notification)
	if err != nil {
		return nil, fmt.Errorf("failed to send notification: %w", err)
	}

	if response.StatusCode != 200 {
		log.Printf("APNS error: %d - %s", response.StatusCode, response.Reason)
		return response, fmt.Errorf("APNS error: %s", response.Reason)
	}

	log.Printf("Successfully sent APNS notification to %s", deviceToken)
	return response, nil
}

// SendSilentNotification sends a silent notification for background updates
func (a *APNSService) SendSilentNotification(ctx context.Context, bundleID, deviceToken string, customData map[string]string) (*apns2.Response, error) {
	notification := &apns2.Notification{
		DeviceToken: deviceToken,
		Topic:       bundleID,
		Priority:    apns2.PriorityLow,
	}

	p := payload.NewPayload().ContentAvailable()

	// Add custom data
	for key, value := range customData {
		p.Custom(key, value)
	}

	notification.Payload = p

	response, err := a.client.PushWithContext(ctx, notification)
	if err != nil {
		return nil, fmt.Errorf("failed to send silent notification: %w", err)
	}

	if response.StatusCode != 200 {
		log.Printf("APNS silent notification error: %d - %s", response.StatusCode, response.Reason)
		return response, fmt.Errorf("APNS error: %s", response.Reason)
	}

	log.Printf("Successfully sent APNS silent notification to %s", deviceToken)
	return response, nil
}

// SendLiveActivity sends a live activity notification
func (a *APNSService) SendLiveActivity(ctx context.Context, bundleID, deviceToken string, activityPayload APNSPayload) (*apns2.Response, error) {
	notification := &apns2.Notification{
		DeviceToken: deviceToken,
		Topic:       bundleID + ".push-type.liveactivity",
		Priority:    apns2.PriorityHigh,
		PushType:    apns2.PushTypeLiveActivity,
	}

	p := payload.NewPayload()

	// Add live activity specific fields
	if activityPayload.Title != "" {
		p.Alert(activityPayload.Title)
	}
	if activityPayload.Body != "" {
		p.AlertBody(activityPayload.Body)
	}
	if len(activityPayload.ContentState) > 0 {
		p.Custom("content-state", activityPayload.ContentState)
	}
	if activityPayload.TargetContentID != "" {
		p.Custom("target-content-id", activityPayload.TargetContentID)
	}

	// Add timestamp
	p.Custom("timestamp", time.Now().Unix())

	notification.Payload = p

	response, err := a.client.PushWithContext(ctx, notification)
	if err != nil {
		return nil, fmt.Errorf("failed to send live activity: %w", err)
	}

	if response.StatusCode != 200 {
		log.Printf("APNS live activity error: %d - %s", response.StatusCode, response.Reason)
		return response, fmt.Errorf("APNS error: %s", response.Reason)
	}

	log.Printf("Successfully sent APNS live activity to %s", deviceToken)
	return response, nil
}

// ValidateToken validates an APNS device token format
func ValidateToken(deviceToken string) error {
	if len(deviceToken) != 64 {
		return fmt.Errorf("invalid token length: expected 64 characters, got %d", len(deviceToken))
	}

	// Check if token contains only hex characters
	for _, c := range deviceToken {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return fmt.Errorf("invalid token format: contains non-hex character")
		}
	}

	return nil
}

// Close closes the APNS client connection
func (a *APNSService) Close() error {
	if a.client != nil {
		a.client.CloseIdleConnections()
	}
	return nil
}
