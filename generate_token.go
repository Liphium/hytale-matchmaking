package main

import (
	"fmt"
	"log"
	"net/url"
	"time"

	"github.com/Liphium/hytale-auth-manager/util"

	"github.com/gofiber/fiber/v2"
)

const (
	DeviceAuthURL   = "https://oauth.accounts.hytale.com/oauth2/device/auth"
	TokenURL        = "https://oauth.accounts.hytale.com/oauth2/token"
	ProfilesURL     = "https://account-data.hytale.com/my-account/get-profiles"
	GameSessionURL  = "https://sessions.hytale.com/game-session/new"
	ClientID        = "hytale-server"
	Scope           = "openid offline auth:server"
	DefaultInterval = 5 * time.Second
)

type DeviceCodeResponse struct {
	DeviceCode              string `json:"device_code"`
	UserCode                string `json:"user_code"`
	VerificationURI         string `json:"verification_uri"`
	VerificationURIComplete string `json:"verification_uri_complete"`
	ExpiresIn               int    `json:"expires_in"`
	Interval                int    `json:"interval"`
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
	Error        string `json:"error,omitempty"`
}

type ProfilesResponse struct {
	Owner    string    `json:"owner"`
	Profiles []Profile `json:"profiles"`
}

type Profile struct {
	UUID     string `json:"uuid"`
	Username string `json:"username"`
}

type GameSessionResponse struct {
	SessionToken  string `json:"sessionToken"`
	IdentityToken string `json:"identityToken"`
	ExpiresAt     string `json:"expiresAt"`
}

type GameSessionRequest struct {
	UUID string `json:"uuid"`
}

func generateToken(c *fiber.Ctx) error {
	// Step 1: Request Device Code
	deviceCode, err := requestDeviceCode()
	if err != nil {
		log.Printf("Error requesting device code: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to request device code",
		})
	}

	log.Printf("Device Code obtained. User code: %s", deviceCode.UserCode)
	log.Printf("User should visit: %s", deviceCode.VerificationURIComplete)

	// Start polling in a goroutine
	go pollForToken(deviceCode)

	// Redirect user to the verification URL
	return c.Redirect(deviceCode.VerificationURIComplete, fiber.StatusFound)
}

func requestDeviceCode() (*DeviceCodeResponse, error) {

	// OAuth endpoints require application/x-www-form-urlencoded
	data := url.Values{}
	data.Set("client_id", ClientID)
	data.Set("scope", Scope)

	deviceCode, err := util.PostForm[DeviceCodeResponse](DeviceAuthURL, data, util.Headers{})
	if err != nil {
		return nil, err
	}

	return &deviceCode, nil
}

func pollForToken(deviceCode *DeviceCodeResponse) {
	interval := time.Duration(deviceCode.Interval) * time.Second
	if interval == 0 {
		interval = DefaultInterval
	}

	expiresAt := time.Now().Add(time.Duration(deviceCode.ExpiresIn) * time.Second)

	log.Printf("Starting token polling (interval: %v)", interval)

	for time.Now().Before(expiresAt) {
		time.Sleep(interval)

		tokenResp, err := pollTokenEndpoint(deviceCode.DeviceCode)
		if err != nil {
			log.Printf("Error polling token endpoint: %v", err)
			continue
		}

		if tokenResp.Error == "authorization_pending" {
			log.Printf("Waiting for user authorization...")
			continue
		}

		if tokenResp.Error != "" {
			log.Printf("Token error: %s", tokenResp.Error)
			return
		}

		if tokenResp.AccessToken != "" {
			log.Printf("✓ Access token obtained successfully!")
			handleTokenSuccess(tokenResp.AccessToken, tokenResp.RefreshToken)
			return
		}
	}

	log.Printf("Device code expired without authorization")
}

func pollTokenEndpoint(deviceCode string) (*TokenResponse, error) {
	// OAuth endpoints require application/x-www-form-urlencoded
	data := url.Values{}
	data.Set("client_id", ClientID)
	data.Set("grant_type", "urn:ietf:params:oauth:grant-type:device_code")
	data.Set("device_code", deviceCode)

	tokenResp, err := util.PostFormAllowErrors[TokenResponse](TokenURL, data, util.Headers{})
	if err != nil {
		return nil, err
	}

	return &tokenResp, nil
}

func handleTokenSuccess(accessToken, refreshToken string) {
	// Step 4: Get Available Profiles
	profiles, err := getProfiles(accessToken)
	if err != nil {
		log.Printf("Error getting profiles: %v", err)
		return
	}

	if len(profiles.Profiles) == 0 {
		log.Printf("No profiles found for this account")
		return
	}

	// Use the first profile
	profile := profiles.Profiles[0]
	log.Printf("Using profile: %s (UUID: %s)", profile.Username, profile.UUID)

	// Store the token
	token := Token{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		Account:      profile.Username,
		UUID:         profile.UUID,
	}

	addToken(token)
	log.Printf("✓ Token stored successfully for account: %s", profile.Username)
}

func getProfiles(accessToken string) (*ProfilesResponse, error) {
	headers := util.Headers{
		"Authorization": fmt.Sprintf("Bearer %s", accessToken),
	}

	profiles, err := util.Get[ProfilesResponse](ProfilesURL, headers)
	if err != nil {
		return nil, err
	}

	return &profiles, nil
}

func createGameSession(accessToken, uuid string) (*GameSessionResponse, error) {
	headers := util.Headers{
		"Authorization": fmt.Sprintf("Bearer %s", accessToken),
	}

	requestBody := GameSessionRequest{
		UUID: uuid,
	}

	session, err := util.Post[GameSessionResponse](GameSessionURL, requestBody, headers)
	if err != nil {
		return nil, err
	}

	return &session, nil
}
