package contact

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

type ContactRequest struct {
	Name    string `json:"name"`
	Email   string `json:"email"`
	Subject string `json:"subject"`
	Message string `json:"message"`
}

func SendContactEmail(req ContactRequest) error {
	apiKey := os.Getenv("RESEND_API_KEY")
	toEmail := os.Getenv("CONTACT_EMAIL_TO")
	fromEmail := os.Getenv("CONTACT_EMAIL_FROM")

	if apiKey == "" {
		return fmt.Errorf("RESEND_API_KEY is not configured")
	}

	payload := map[string]interface{}{
		"from":     fromEmail,
		"to":       []string{toEmail},
		"reply_to": req.Email,
		"subject":  fmt.Sprintf("[GeoPulse Contact] %s - %s", req.Subject, req.Name),
		"html": fmt.Sprintf(`
			<h2>New Contact Request / Quota Extension</h2>
			<p><strong>Name:</strong> %s</p>
			<p><strong>Email:</strong> %s</p>
			<p><strong>Subject:</strong> %s</p>
			<p><strong>Message:</strong></p>
			<blockquote style="background: #18181b; color: #f4f4f5; padding: 12px; border-left: 4px solid #06b6d4;">
				%s
			</blockquote>
		`, req.Name, req.Email, req.Subject, req.Message),
	}

	jsonBytes, _ := json.Marshal(payload)
	httpReq, err := http.NewRequest("POST", "https://api.resend.com/emails", bytes.NewBuffer(jsonBytes))
	if err != nil {
		return err
	}

	httpReq.Header.Set("Authorization", "Bearer "+apiKey)
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("resend API returned status %d", resp.StatusCode)
	}

	return nil
}