package auth

import (
	"context"
	"fmt"

	firebase "firebase.google.com/go/v4"
	firebaseAuth "firebase.google.com/go/v4/auth"
	"google.golang.org/api/option"
)

type FirebaseService struct {
	client *firebaseAuth.Client
}

func NewFirebaseService(credentialsPath string) (*FirebaseService, error) {
	opt := option.WithCredentialsFile(credentialsPath)
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		return nil, err
	}
	client, err := app.Auth(context.Background())
	if err != nil {
		return nil, err
	}
	return &FirebaseService{client: client}, nil
}

func (s *FirebaseService) VerifyIDToken(ctx context.Context, idToken string) (uid, email, name string, err error) {
	token, err := s.client.VerifyIDToken(ctx, idToken)
	if err != nil {
		return "", "", "", fmt.Errorf("firebase verify failed: %w", err)
	}
	email = getString(token.Claims, "email")
	name = getString(token.Claims, "name")
	return token.UID, email, name, nil
}

func getString(claims map[string]interface{}, key string) string {
	if v, ok := claims[key].(string); ok {
		return v
	}
	return ""
}
