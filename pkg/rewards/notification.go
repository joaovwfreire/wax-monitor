package rewards

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

var currentAccessToken string
var currentRefreshToken string

type ZohoConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURI  string
	AccountID    string
	AccessToken  string
	RefreshToken string
}

type AccountDataResponse struct {
	Status Status  `json:"status"`
	Data   []Datum `json:"data"`
}

type Datum struct {
	Country             string           `json:"country"`
	LastLogin           int64            `json:"lastLogin"`
	ActiveSyncEnabled   bool             `json:"activeSyncEnabled"`
	IncomingBlocked     bool             `json:"incomingBlocked"`
	Language            string           `json:"language"`
	Type                string           `json:"type"`
	ExtraStorage        Address          `json:"extraStorage"`
	IncomingUserName    string           `json:"incomingUserName"`
	EmailAddress        []EmailAddress   `json:"emailAddress"`
	MailboxStatus       string           `json:"mailboxStatus"`
	PopBlocked          bool             `json:"popBlocked"`
	EncryptedZuid       string           `json:"encryptedZuid"`
	UsedStorage         int64            `json:"usedStorage"`
	SpamcheckEnabled    bool             `json:"spamcheckEnabled"`
	IMAPAccessEnabled   bool             `json:"imapAccessEnabled"`
	TimeZone            string           `json:"timeZone"`
	AccountCreationTime int64            `json:"accountCreationTime"`
	Zuid                int64            `json:"zuid"`
	WebBlocked          bool             `json:"webBlocked"`
	PlanStorage         int64            `json:"planStorage"`
	FirstName           string           `json:"firstName"`
	AccountID           string           `json:"accountId"`
	Sequence            int64            `json:"sequence"`
	MailboxAddress      string           `json:"mailboxAddress"`
	LastPasswordReset   int64            `json:"lastPasswordReset"`
	TfaEnabled          bool             `json:"tfaEnabled"`
	Status              bool             `json:"status"`
	LastName            string           `json:"lastName"`
	AccountDisplayName  string           `json:"accountDisplayName"`
	Role                string           `json:"role"`
	Gender              string           `json:"gender"`
	AccountName         string           `json:"accountName"`
	DisplayName         string           `json:"displayName"`
	IsLogoExist         bool             `json:"isLogoExist"`
	URI                 string           `json:"URI"`
	PrimaryEmailAddress string           `json:"primaryEmailAddress"`
	Enabled             bool             `json:"enabled"`
	MailboxCreationTime int64            `json:"mailboxCreationTime"`
	BasicStorage        string           `json:"basicStorage"`
	LastClient          string           `json:"lastClient"`
	AllowedStorage      int64            `json:"allowedStorage"`
	SendMailDetails     []SendMailDetail `json:"sendMailDetails"`
	PopFetchTime        int64            `json:"popFetchTime"`
	Address             Address          `json:"address"`
	PlanType            int64            `json:"planType"`
	UserExpiry          int64            `json:"userExpiry"`
	PopAccessEnabled    bool             `json:"popAccessEnabled"`
	IMAPBlocked         bool             `json:"imapBlocked"`
	IamUserRole         string           `json:"iamUserRole"`
	OutgoingBlocked     bool             `json:"outgoingBlocked"`
	PolicyID            PolicyID         `json:"policyId"`
}

type Address struct {
}

type EmailAddress struct {
	IsAlias     bool   `json:"isAlias"`
	IsPrimary   bool   `json:"isPrimary"`
	MailID      string `json:"mailId"`
	IsConfirmed bool   `json:"isConfirmed"`
}

type PolicyID struct {
	The1082700000243868974 string `json:"1082700000243868974"`
	Zoid                   int64  `json:"zoid"`
}

type SendMailDetail struct {
	SendMailID         string `json:"sendMailId"`
	DisplayName        string `json:"displayName"`
	ServerName         string `json:"serverName"`
	SignatureID        string `json:"signatureId"`
	ServerPort         int64  `json:"serverPort"`
	UserName           string `json:"userName"`
	ConnectionType     string `json:"connectionType"`
	Mode               string `json:"mode"`
	Validated          bool   `json:"validated"`
	FromAddress        string `json:"fromAddress"`
	SMTPConnection     int64  `json:"smtpConnection"`
	ValidationRequired bool   `json:"validationRequired"`
	ValidationState    int64  `json:"validationState"`
	Status             bool   `json:"status"`
}

type Status struct {
	Code        int64  `json:"code"`
	Description string `json:"description"`
}

func NewZohoConfig() *ZohoConfig {
	return &ZohoConfig{
		ClientID:     os.Getenv("CLIENT_ID"),
		ClientSecret: os.Getenv("CLIENT_SECRET"),
		RedirectURI:  os.Getenv("EMAIL_REDIRECT_URI"),
	}
}

func openURL(url string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "mac":
		cmd = exec.Command("open", url)
	default:
		return os.ErrNotExist
	}

	return cmd.Start()
}

func (z *ZohoConfig) authenticate(scope string) {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
	currentAccessToken, err = z.getAccessToken(scope)
	if err != nil {
		fmt.Println("Error initializing access token:", err)
		return
	}
}

func (z *ZohoConfig) initialize() (string, error) {
	z.authenticate("accounts")

	// TODO: Make an API call to Zoho Mail using z.AccessToken to get the AccountID.
	// Parse the response and set z.AccountID.
	req, err := http.NewRequest("GET", "https://mail.zoho.com/api/accounts", nil)
	if err != nil {
		fmt.Println("Error creating request:", err) // <-- Added this log
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Zoho-oauthtoken "+z.AccessToken)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err) // <-- Added this log
		return "", err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err) // <-- Added this log
		return "", err
	}

	var responseMapping AccountDataResponse
	err = json.Unmarshal(body, &responseMapping)
	if err != nil {
		fmt.Println("Error unmarshalling response body:", err) // <-- Added this log
		return "", err
	}

	// uses first account id in response
	z.AccountID = responseMapping.Data[0].AccountID

	fmt.Println("Response body:", string(body))

	return string(body), err
}

func (z *ZohoConfig) sendEmail(to, subject, content string) error {
	z.authenticate("messages")

	payload := map[string]interface{}{
		"fromAddress": "dev@breeders.zone",
		"toAddress":   to,
		"subject":     subject,
		"content":     content,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		fmt.Println("Error marshalling payload:", err) // <-- Added this log
		return err
	}

	reqUrl := fmt.Sprintf("https://mail.zoho.com/api/accounts/%s/messages", z.AccountID)

	req, err := http.NewRequest("POST", reqUrl, bytes.NewReader(payloadBytes))
	if err != nil {
		fmt.Println("Error creating request:", err) // <-- Added this log
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Zoho-oauthtoken "+z.AccessToken)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err) // <-- Added this log
		return err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err) // <-- Added this log
		return err
	}

	fmt.Printf("HTTP Response Status: %s, Body: %s\n", res.Status, string(body)) // <-- Added this log

	var response map[string]interface{}
	json.Unmarshal(body, &response)
	if status, ok := response["status"].(map[string]interface{}); ok {
		if code, ok := status["code"].(float64); ok && code == 404 {
			if data, ok := response["data"].(map[string]interface{}); ok {
				if errorCode, ok := data["errorCode"].(string); ok && errorCode == "INVALID_OAUTHTOKEN" {
					return fmt.Errorf("INVALID_OAUTHTOKEN")
				}
			}
		}
	}

	fmt.Println("Finished sendEmail function...") // <-- Added this log
	return nil
}

// NotifyEmail remains a top-level function, but now leverages the ZohoConfig struct
func NotifyEmail(to, subject, content string) {
	config := NewZohoConfig()

	_, err := config.initialize()
	if err != nil {
		fmt.Println("Error during initialization:", err)
		return
	}

	fmt.Println("Sleeping for 5 seconds...")
	time.Sleep(5 * time.Second)

	err = config.sendEmail(to, subject, content)
	if err != nil {
		fmt.Println("Error sending email:", err)
	}
}

var authCodeChan = make(chan string, 1)

func startAuthCodeListener() {
	listener, err := net.Listen("tcp", "localhost:8080") // Assuming redirectUri is http://localhost:8080/
	if err != nil {
		fmt.Println("Error starting listener:", err)
		return
	}
	defer listener.Close()

	conn, err := listener.Accept()
	if err != nil {
		fmt.Println("Error accepting connection:", err)
		return
	}
	defer conn.Close()

	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		fmt.Println("Error reading from connection:", err)
		return
	}

	request := string(buffer[:n])
	startIndex := strings.Index(request, "code=")
	if startIndex == -1 {
		fmt.Println("Authorization code not found in request")
		return
	}
	startIndex += 5
	endIndex := strings.Index(request[startIndex:], "&")
	if endIndex == -1 {
		endIndex = strings.Index(request[startIndex:], " ")
	}
	authCode := request[startIndex : startIndex+endIndex]

	authCodeChan <- authCode
}

func (z *ZohoConfig) getAccessToken(scope string) (string, error) {
	redirectUri := os.Getenv("EMAIL_REDIRECT_URI")
	clientId := os.Getenv("CLIENT_ID")

	// Start the listener to capture the authorization code
	go startAuthCodeListener()
	var authURL string
	// Redirect URL to get the authorization code
	if scope == "accounts" {
		authURL = fmt.Sprintf("https://accounts.zoho.com/oauth/v2/auth?response_type=code&client_id=%s&scope=ZohoMail.accounts.READ&redirect_uri=%s", url.QueryEscape(clientId), url.QueryEscape(redirectUri))
	} else if scope == "messages" {
		authURL = fmt.Sprintf("https://accounts.zoho.com/oauth/v2/auth?response_type=code&client_id=%s&scope=ZohoMail.messages.CREATE&redirect_uri=%s", url.QueryEscape(clientId), url.QueryEscape(redirectUri))
	}

	fmt.Println("Opening default browser to authorize:", authURL)

	if err := openURL(authURL); err != nil {
		log.Fatal(err)
	}
	time.Sleep(2 * time.Second)
	// Wait for the listener to capture the authorization code
	fmt.Print("Waiting for authorization code...")
	authCode := <-authCodeChan

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
	accessTokenValue, ok := result["access_token"]
	if !ok {
		return "", fmt.Errorf("access_token not found in response")
	}

	accessToken, ok := accessTokenValue.(string)
	if !ok {
		return "", fmt.Errorf("access_token is not a string")
	}
	fmt.Println("Access token obtained succesfully")
	z.AccessToken = accessToken

	return accessToken, nil
}

func (z *ZohoConfig) updateToken() (string, error) {
	clientId := os.Getenv("CLIENT_ID")
	clientSecret := os.Getenv("CLIENT_SECRET")
	redirectUri := os.Getenv("EMAIL_REDIRECT_URI")

	data := url.Values{
		"refresh_token": {z.RefreshToken},
		"client_id":     {clientId},
		"client_secret": {clientSecret},
		"redirect_uri":  {redirectUri},
		"grant_type":    {"refresh_token"},
	}

	resp, err := http.PostForm("https://accounts.zoho.com/oauth/v2/token", data)
	if err != nil {
		fmt.Println("Error refreshing token:", err)
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
	accessToken := result["access_token"].(string)
	fmt.Println("Refreshed access token obtained succesfully")
	z.AccessToken = accessToken

	return accessToken, nil
}

func NotifyTelegram() {}
