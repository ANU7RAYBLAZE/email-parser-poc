package entities

import "time"

type EmailMessage struct {
	ID            string            `json:"id"`
	Subject       string            `json:"subject"`
	From          string            `json:"from"`
	To            string            `json:"to"`
	Date          time.Time         `json:"date"`
	Body          string            `json:"body"`
	Headers       map[string]string `json:"headers"`
	IsPromotional bool              `json:"is_promotional"`
}
type EmailList struct {
	Emails        []EmailMessage `json:"emails"`
	NextPageToken string         `json:"next_page_token,omitempty"`
	TotalCount    int            `json:"total_count"`
}

type EmailFilter struct {
	MaxResults      int    `json:"max_results"`
	Query           string `json:"query,omitempty"`
	PageToken       string `json:"page_token,omitempty"`
	OnlyPromotional bool   `json:"only_promotional"`
}
