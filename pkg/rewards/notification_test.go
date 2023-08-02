package rewards

import (
	"os"
	"testing"
)

func TestNotifyEmail(t *testing.T) {

	os.Setenv("EMAIL_REDIRECT_URI", "YOUR_REDIRECT_URI")
	os.Setenv("CLIENT_ID", "YOUR_CLIENT_ID")
	os.Setenv("CLIENT_SECRET", "YOUR_CLIENT_SECRET")
	// Define test cases
	tests := []struct {
		to      string
		subject string
		content string
	}{
		{"jvwfreire@gmail.com", "Test Subject", "Test Content"},
		// Add more test cases as needed
	}

	// this is actually the end user flow, however we can't test this automatically. The solution is to mock access token values
	accessToken, err := GetAccessToken()
	if err != nil {
		t.Errorf("Error getting access token: %v", err)
	}

	for _, test := range tests {
		NotifyEmail(test.to, test.subject, test.content, accessToken)
		// You may want to add more detailed assertions here to check the behavior of NotifyEmail
		// Depending on the behavior of the Zoho API and what you want to test, this may require
		// mocking the HTTP client or other dependencies
	}
}
