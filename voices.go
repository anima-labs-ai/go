package anima

import (
	"context"
	"net/http"
	"net/url"
)

// Voice represents an available voice for AI agent phone calls.
type Voice struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Provider    string `json:"provider"`
	Tier        string `json:"tier"`
	Gender      string `json:"gender,omitempty"`
	Language    string `json:"language"`
	Accent      string `json:"accent,omitempty"`
	Style       string `json:"style,omitempty"`
	AgeRange    string `json:"ageRange,omitempty"`
	Description string `json:"description,omitempty"`
	PreviewURL  string `json:"previewUrl,omitempty"`
}

// VoiceList wraps a list of voices.
type VoiceList struct {
	Voices []Voice `json:"voices"`
}

// ListVoicesParams contains parameters for filtering the voice catalog.
type ListVoicesParams struct {
	Tier     string
	Gender   string
	Language string
}

// VoicesService provides methods for browsing the voice catalog.
type VoicesService struct {
	client *httpClient
}

// newVoicesService creates a new VoicesService.
func newVoicesService(c *httpClient) *VoicesService {
	return &VoicesService{client: c}
}

// List returns available voices, optionally filtered by tier, gender, or language.
func (s *VoicesService) List(ctx context.Context, params ListVoicesParams) (*VoiceList, error) {
	q := url.Values{}
	if params.Tier != "" {
		q.Set("tier", params.Tier)
	}
	if params.Gender != "" {
		q.Set("gender", params.Gender)
	}
	if params.Language != "" {
		q.Set("language", params.Language)
	}
	list, err := Do[VoiceList](ctx, s.client, http.MethodGet, "/voice/catalog", nil, q)
	if err != nil {
		return nil, err
	}
	return &list, nil
}
