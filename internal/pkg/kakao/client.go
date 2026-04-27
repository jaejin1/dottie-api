package kakao

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const userMeURL = "https://kapi.kakao.com/v2/user/me"

type KakaoUser struct {
	ID           int64
	Nickname     string
	ProfileImage string
}

type Client interface {
	GetUser(ctx context.Context, accessToken string) (*KakaoUser, error)
}

type HTTPClient struct {
	httpClient *http.Client
}

func NewClient() *HTTPClient {
	return &HTTPClient{httpClient: &http.Client{}}
}

func (c *HTTPClient) GetUser(ctx context.Context, accessToken string) (*KakaoUser, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, userMeURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("kakao api error %d: %s", resp.StatusCode, string(body))
	}

	var raw struct {
		ID           int64 `json:"id"`
		KakaoAccount struct {
			Profile struct {
				Nickname        string `json:"nickname"`
				ProfileImageURL string `json:"profile_image_url"`
			} `json:"profile"`
		} `json:"kakao_account"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, err
	}

	return &KakaoUser{
		ID:           raw.ID,
		Nickname:     raw.KakaoAccount.Profile.Nickname,
		ProfileImage: raw.KakaoAccount.Profile.ProfileImageURL,
	}, nil
}
