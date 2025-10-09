package notifications

import (
	"context"
	"fmt"
	"log"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"google.golang.org/api/option"
)

// FCMService handles Firebase Cloud Messaging
type FCMService struct {
	client *messaging.Client
}

// NewFCMService creates a new FCM service
func NewFCMService(credentialsPath string) (*FCMService, error) {
	opt := option.WithCredentialsFile(credentialsPath)
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Firebase app: %w", err)
	}

	client, err := app.Messaging(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to get messaging client: %w", err)
	}

	return &FCMService{client: client}, nil
}

// NotificationPayload represents a push notification
type NotificationPayload struct {
	Title       string            `json:"title"`
	Body        string            `json:"body"`
	ImageURL    string            `json:"imageUrl,omitempty"`
	Icon        string            `json:"icon,omitempty"`
	ClickAction string            `json:"clickAction,omitempty"`
	Data        map[string]string `json:"data,omitempty"`
}

// SendToToken sends a notification to a specific device token
func (f *FCMService) SendToToken(ctx context.Context, token string, payload NotificationPayload) (string, error) {
	message := &messaging.Message{
		Token: token,
		Notification: &messaging.Notification{
			Title:    payload.Title,
			Body:     payload.Body,
			ImageURL: payload.ImageURL,
		},
		Data: payload.Data,
		Android: &messaging.AndroidConfig{
			Notification: &messaging.AndroidNotification{
				Icon:        payload.Icon,
				ClickAction: payload.ClickAction,
				Priority:    messaging.PriorityHigh,
			},
		},
		APNS: &messaging.APNSConfig{
			Payload: &messaging.APNSPayload{
				Aps: &messaging.Aps{
					Alert: &messaging.ApsAlert{
						Title: payload.Title,
						Body:  payload.Body,
					},
					Sound: "default",
					Badge: nil,
				},
			},
		},
	}

	response, err := f.client.Send(ctx, message)
	if err != nil {
		return "", fmt.Errorf("failed to send message: %w", err)
	}

	log.Printf("Successfully sent message: %s", response)
	return response, nil
}

// SendToTokens sends a notification to multiple device tokens
func (f *FCMService) SendToTokens(ctx context.Context, tokens []string, payload NotificationPayload) (*messaging.BatchResponse, error) {
	message := &messaging.MulticastMessage{
		Tokens: tokens,
		Notification: &messaging.Notification{
			Title:    payload.Title,
			Body:     payload.Body,
			ImageURL: payload.ImageURL,
		},
		Data: payload.Data,
		Android: &messaging.AndroidConfig{
			Notification: &messaging.AndroidNotification{
				Icon:        payload.Icon,
				ClickAction: payload.ClickAction,
				Priority:    messaging.PriorityHigh,
			},
		},
		APNS: &messaging.APNSConfig{
			Payload: &messaging.APNSPayload{
				Aps: &messaging.Aps{
					Alert: &messaging.ApsAlert{
						Title: payload.Title,
						Body:  payload.Body,
					},
					Sound: "default",
				},
			},
		},
	}

	response, err := f.client.SendMulticast(ctx, message)
	if err != nil {
		return nil, fmt.Errorf("failed to send multicast message: %w", err)
	}

	log.Printf("Successfully sent %d messages, %d failures", response.SuccessCount, response.FailureCount)
	return response, nil
}

// SendToTopic sends a notification to a topic
func (f *FCMService) SendToTopic(ctx context.Context, topic string, payload NotificationPayload) (string, error) {
	message := &messaging.Message{
		Topic: topic,
		Notification: &messaging.Notification{
			Title:    payload.Title,
			Body:     payload.Body,
			ImageURL: payload.ImageURL,
		},
		Data: payload.Data,
		Android: &messaging.AndroidConfig{
			Notification: &messaging.AndroidNotification{
				Icon:        payload.Icon,
				ClickAction: payload.ClickAction,
				Priority:    messaging.PriorityHigh,
			},
		},
		APNS: &messaging.APNSConfig{
			Payload: &messaging.APNSPayload{
				Aps: &messaging.Aps{
					Alert: &messaging.ApsAlert{
						Title: payload.Title,
						Body:  payload.Body,
					},
					Sound: "default",
				},
			},
		},
	}

	response, err := f.client.Send(ctx, message)
	if err != nil {
		return "", fmt.Errorf("failed to send topic message: %w", err)
	}

	log.Printf("Successfully sent message to topic %s: %s", topic, response)
	return response, nil
}

// SubscribeToTopic subscribes tokens to a topic
func (f *FCMService) SubscribeToTopic(ctx context.Context, tokens []string, topic string) (*messaging.TopicManagementResponse, error) {
	response, err := f.client.SubscribeToTopic(ctx, tokens, topic)
	if err != nil {
		return nil, fmt.Errorf("failed to subscribe to topic: %w", err)
	}

	log.Printf("Successfully subscribed %d tokens to topic %s", len(tokens), topic)
	return response, nil
}

// UnsubscribeFromTopic unsubscribes tokens from a topic
func (f *FCMService) UnsubscribeFromTopic(ctx context.Context, tokens []string, topic string) (*messaging.TopicManagementResponse, error) {
	response, err := f.client.UnsubscribeFromTopic(ctx, tokens, topic)
	if err != nil {
		return nil, fmt.Errorf("failed to unsubscribe from topic: %w", err)
	}

	log.Printf("Successfully unsubscribed %d tokens from topic %s", len(tokens), topic)
	return response, nil
}