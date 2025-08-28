package gmail

import (
	"context"
	"email-parser-poc/internal/domain/entities"
	"email-parser-poc/internal/ports/outgoing"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type gmailRepository struct {
	client        *http.Client
	tokenProvider outgoing.TokenProvider
}

func NewGmailRepository(tokenProvider outgoing.TokenProvider) outgoing.EmailRepository {

	return &gmailRepository{
		client:        &http.Client{},
		tokenProvider: tokenProvider,
	}
}

func (r *gmailRepository) FetchEmails(ctx context.Context, filter entities.EmailFilter) (*entities.EmailList, error) {
	var allEmails []entities.EmailMessage
	pageToken := filter.PageToken
	filter.OnlyPromotional = true

	for {
		messageIDs, nextPageToken, err := r.getMessageIDs(ctx, filter.MaxResults, pageToken)
		if err != nil {
			return nil, err
		}

		fmt.Printf("✓ Got %d message IDs from Gmail API\n", len(messageIDs)) // DEBUG

		for _, msgID := range messageIDs {
			if len(allEmails) >= filter.MaxResults {
				break
			}

			fmt.Printf("Processing email ID: %s\n", msgID) // DEBUG

			email, err := r.getEmailContent(ctx, msgID)
			if err != nil {
				fmt.Printf("❌ Failed to get email %s: %v\n", msgID, err)
				continue
			}
			if filter.OnlyPromotional && !email.IsPromotional {
				continue // Skip non-promotional emails
			}

			fmt.Printf("✓ Successfully processed email: %s\n", email.Subject) // DEBUG
			allEmails = append(allEmails, *email)
		}

		fmt.Printf("Total emails processed: %d\n", len(allEmails)) // DEBUG

		if nextPageToken == "" || len(allEmails) >= filter.MaxResults {
			break
		}

		pageToken = nextPageToken
	}

	return &entities.EmailList{
		Emails:        allEmails,
		NextPageToken: "",
		TotalCount:    len(allEmails),
	}, nil
}

func (r *gmailRepository) getMessageIDs(ctx context.Context, maxResults int, pageToken string) ([]string, string, error) {
	apiURL := "https://gmail.googleapis.com/gmail/v1/users/me/messages"
	params := url.Values{}
	params.Add("maxResults", fmt.Sprintf("%d", min(maxResults, 500)))

	if pageToken != "" {
		params.Add("pageToken", pageToken)
	}

	fullURL := apiURL + "?" + params.Encode()

	fmt.Printf("🔍 Making request to: %s\n", fullURL) // DEBUG

	req, err := http.NewRequestWithContext(ctx, "GET", fullURL, nil)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+r.tokenProvider.GetAccessToken())

	resp, err := r.client.Do(req)
	if err != nil {
		fmt.Printf("❌ Request failed: %v\n", err) // DEBUG
		return nil, "", fmt.Errorf("failed to get message list: %w", err)
	}
	defer resp.Body.Close()

	fmt.Printf("📡 Response status: %s\n", resp.Status) // DEBUG

	// Read the full response body for debugging
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read response body: %w", err)
	}

	fmt.Printf("📄 Response body: %s\n", string(bodyBytes)) // DEBUG

	// Check for HTTP errors
	if resp.StatusCode != 200 {
		return nil, "", fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var listResp MessagesListResponse
	if err := json.Unmarshal(bodyBytes, &listResp); err != nil {
		return nil, "", fmt.Errorf("failed to decode message list: %w", err)
	}

	var messageIDs []string
	for _, msg := range listResp.Messages {
		messageIDs = append(messageIDs, msg.ID)
	}

	return messageIDs, listResp.NextPageToken, nil
}
func (r *gmailRepository) getEmailContent(ctx context.Context, messageID string) (*entities.EmailMessage, error) {
	apiURL := fmt.Sprintf("https://gmail.googleapis.com/gmail/v1/users/me/messages/%s?format=full", messageID)

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+r.tokenProvider.GetAccessToken())

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get message: %w", err)
	}
	defer resp.Body.Close()

	var gmailMsg GmailMessage
	if err := json.NewDecoder(resp.Body).Decode(&gmailMsg); err != nil {
		return nil, fmt.Errorf("failed to decode message: %w", err)
	}

	email := r.parseGmailMessage(gmailMsg)
	return &email, nil
}

func (r *gmailRepository) parseGmailMessage(gmailMsg GmailMessage) entities.EmailMessage {
	email := entities.EmailMessage{
		ID:      gmailMsg.ID,
		Headers: make(map[string]string),
	}

	for _, header := range gmailMsg.Payload.Headers {
		email.Headers[header.Name] = header.Value

		switch strings.ToLower(header.Name) {
		case "subject":
			email.Subject = header.Value
		case "from":
			email.From = header.Value
		case "to":
			email.To = header.Value
		case "date":
			email.Date = r.parseDate(header.Value)
		}
	}

	email.Body = r.extractBody(gmailMsg.Payload)
	email.IsPromotional = r.classifyAsPromotional(&email)
	return email
}

func (r *gmailRepository) parseDate(dateStr string) time.Time {
	layouts := []string{
		time.RFC1123Z,
		time.RFC1123,
		"Mon, 2 Jan 2006 15:04:05 -0700",
		time.RFC3339,
	}

	for _, layout := range layouts {
		if t, err := time.Parse(layout, dateStr); err == nil {
			return t
		}
	}

	return time.Now()
}

func (r *gmailRepository) classifyAsPromotional(email *entities.EmailMessage) bool {
	promotionalKeywords := []string{
		"sale", "discount", "off", "deal", "offer", "save", "free",
		"limited time", "expires", "ending soon", "last chance",
		"25%", "50%", "percent", "promo", "coupon", "unsubscribe",
		"marketing email", "promotional", "newsletter",
	}
	content := strings.ToLower(email.Subject + " " + email.Body + " " + email.From)

	score := 0

	for _, keyword := range promotionalKeywords {
		if strings.Contains(content, keyword) {
			score++
		}
	}

	if _, exits := email.Headers["List-Unsubscribe"]; exits {
		score += 3
	}

	return score >= 2
}

func (r *gmailRepository) extractBody(payload Payload) string {
	if payload.Body.Data != "" {
		if decoded, err := r.decodeBase64URL(payload.Body.Data); err == nil {
			return decoded
		}
	}

	for _, part := range payload.Parts {
		if part.MimeType == "text/plain" || part.MimeType == "text/html" {
			if part.Body.Data != "" {
				if decoded, err := r.decodeBase64URL(part.Body.Data); err == nil {
					return decoded
				}
			}
		}

		if len(part.Parts) > 0 {
			nestedBody := r.extractBody(Payload{Parts: part.Parts})
			if nestedBody != "" {
				return nestedBody
			}
		}
	}

	return ""
}

func (r *gmailRepository) decodeBase64URL(data string) (string, error) {
	data = strings.ReplaceAll(data, "-", "+")
	data = strings.ReplaceAll(data, "_", "/")

	for len(data)%4 != 0 {
		data += "="
	}

	decoded, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return "", err
	}

	return string(decoded), nil
}

// Keep your existing Gmail API structs
type GmailMessage struct {
	ID           string   `json:"id"`
	ThreadID     string   `json:"threadId"`
	LabelIDs     []string `json:"labelIds"`
	Snippet      string   `json:"snippet"`
	HistoryID    string   `json:"historyId"`
	InternalDate string   `json:"internalDate"`
	Payload      Payload  `json:"payload"`
	SizeEstimate int      `json:"sizeEstimate"`
}

type Payload struct {
	PartID   string   `json:"partId"`
	MimeType string   `json:"mimeType"`
	Filename string   `json:"filename"`
	Headers  []Header `json:"headers"`
	Body     Body     `json:"body"`
	Parts    []Part   `json:"parts"`
}

type Header struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type Body struct {
	AttachmentID string `json:"attachmentId"`
	Size         int    `json:"size"`
	Data         string `json:"data"`
}

type Part struct {
	PartID   string   `json:"partId"`
	MimeType string   `json:"mimeType"`
	Filename string   `json:"filename"`
	Headers  []Header `json:"headers"`
	Body     Body     `json:"body"`
	Parts    []Part   `json:"parts"`
}

type MessageRef struct {
	ID       string `json:"id"`
	ThreadID string `json:"threadId"`
}

type MessagesListResponse struct {
	Messages      []MessageRef `json:"messages"`
	NextPageToken string       `json:"nextPageToken"`
}
