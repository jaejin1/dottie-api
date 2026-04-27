package firebase

import (
	"context"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"google.golang.org/api/option"
)

type Client struct {
	auth *auth.Client
}

func NewClient(credentialsJSON []byte) (*Client, error) {
	app, err := firebase.NewApp(context.Background(), nil,
		option.WithCredentialsJSON(credentialsJSON))
	if err != nil {
		return nil, err
	}
	authClient, err := app.Auth(context.Background())
	if err != nil {
		return nil, err
	}
	return &Client{auth: authClient}, nil
}

func (c *Client) VerifyIDToken(ctx context.Context, idToken string) (*auth.Token, error) {
	return c.auth.VerifyIDToken(ctx, idToken)
}

func (c *Client) CreateCustomToken(ctx context.Context, uid string) (string, error) {
	return c.auth.CustomToken(ctx, uid)
}
