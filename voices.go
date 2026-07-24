package anima

import (
	"context"
	"net/http"
	"net/url"
)

// Voice represents a voice in the catalog. Vendor-neutral: the underlying
// provider/model is never exposed — only descriptive metadata and a proxied
// preview URL.
type Voice struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Gender      string   `json:"gender"`
	Accent      string   `json:"accent,omitempty"`
	Age         string   `json:"age,omitempty"`
	Descriptors []string `json:"descriptors"`
	UseCases    []string `json:"useCases"`
	Language    string   `json:"language"`
	// SampleURL is a vendor-neutral preview URL under the API host — the client
	// never touches the provider CDN. Empty until a sample clip is generated.
	SampleURL string `json:"sampleUrl,omitempty"`
}

// VoiceList wraps a list of voices.
type VoiceList struct {
	Voices []Voice `json:"voices"`
}

// ListVoicesParams contains parameters for filtering the voice catalog.
type ListVoicesParams struct {
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

// List returns available voices, optionally filtered by gender or language.
func (s *VoicesService) List(ctx context.Context, params ListVoicesParams) (*VoiceList, error) {
	q := url.Values{}
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
