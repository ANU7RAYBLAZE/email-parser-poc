package token

import "email-parser-poc/internal/ports/outgoing"

type staticTokenProvider struct {
	accessToken string
}

func NewStaticTokenProvider(accessToken string) outgoing.TokenProvider {
	return &staticTokenProvider{
		accessToken: accessToken,
	}
}

func (p *staticTokenProvider) GetAccessToken() string {
	return p.accessToken
}
