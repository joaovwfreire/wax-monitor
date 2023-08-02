package rewards

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
)

func NotifyEmail(to, subject, content, accessToken string) {
	payload := map[string]interface{}{
		"fromAddress": "notify@breeders.zone",
		"toAddress":   to,
		"subject":     subject,
		"content":     content,
		"askReceipt":  "yes",
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		fmt.Println("Error: ", err)
		return
	}

	req, err := http.NewRequest("POST", "https://mail.zoho.com/api/accounts/<accountId>/messages", bytes.NewReader(payloadBytes))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")

	accessToken, err := GetAccessToken()
	if err != nil {
		fmt.Println("Error getting access token:", err)
		return
	}

	req.Header.Set("Authorization", "Zoho-oauthtoken "+accessToken)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return
	}

	fmt.Println("Response:", string(body))
}

func GetAccessToken() (string, error) {

	redirectUri := os.Getenv("EMAIL_REDIRECT_URI")
	clientId := os.Getenv("CLIENT_ID")
	// Redirect URL to get the authorization code
	authURL := fmt.Sprintf("https://accounts.zoho.com/oauth/v2/auth?response_type=code&client_id=%s&scope=ZohoMail.messages.CREATE&redirect_uri=%s", url.QueryEscape(clientId), url.QueryEscape(redirectUri))
	fmt.Println("Please visit the URL to authorize:", authURL)
	// Wait for user input with the authorization code
	fmt.Print("Enter the authorization code:")
	var authCode string
	fmt.Scanln(&authCode)

	// Exchange the authorization code for an access token
	tokenURL := "https://accounts.zoho.com/oauth/v2/token"
	data := url.Values{
		"code":          {authCode},
		"client_id":     {clientId},
		"client_secret": {os.Getenv("CLIENT_SECRET")},
		"redirect_uri":  {redirectUri},
		"scope":         {"ZohoMail.messages.CREATE"},
		"grant_type":    {"authorization_code"},
	}
	resp, err := http.PostForm(tokenURL, data)
	if err != nil {
		fmt.Println("Error getting token:", err)
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return "", err
	}

	var result map[string]interface{}
	json.Unmarshal(body, &result)
	//expiryTimestampInMs := result["expires_in "].(int64)
	accessToken := result["access_token"].(string)
	//refreshToken := result["refresh_token"].(string)
	//tokenType := result["token_type"].(string)
	fmt.Println("Access token obtained:", accessToken)

	return accessToken, nil
}

func refreshAuthToken() {}

func NotifyTelegram() {}
