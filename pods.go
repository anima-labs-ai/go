package anima

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// Pod represents an isolated compute pod for an agent.
type Pod struct {
	ID        string                 `json:"id"`
	AgentID   string                 `json:"agentId"`
	Name      string                 `json:"name"`
	Status    string                 `json:"status"`
	Image     string                 `json:"image,omitempty"`
	Resources *PodResources          `json:"resources,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt string                 `json:"createdAt"`
	UpdatedAt string                 `json:"updatedAt"`
}

// PodResources describes the compute resources allocated to a pod.
type PodResources struct {
	CPU    string `json:"cpu,omitempty"`
	Memory string `json:"memory,omitempty"`
	Disk   string `json:"disk,omitempty"`
}

// CreatePodParams contains the parameters for creating a pod.
type CreatePodParams struct {
	AgentID   string                 `json:"agentId"`
	Name      string                 `json:"name"`
	Image     string                 `json:"image,omitempty"`
	Resources *PodResources          `json:"resources,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// UpdatePodParams contains the parameters for updating a pod.
type UpdatePodParams struct {
	Name      string                 `json:"name,omitempty"`
	Image     string                 `json:"image,omitempty"`
	Resources *PodResources          `json:"resources,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// PodListParams contains parameters for listing pods.
type PodListParams struct {
	ListParams
	AgentID string
}

// ToQuery converts PodListParams to URL query values.
func (p PodListParams) ToQuery() url.Values {
	q := p.ListParams.ToQuery()
	if p.AgentID != "" {
		q.Set("agentId", p.AgentID)
	}
	return q
}

// PodList wraps a list of pods.
type PodList struct {
	Items []Pod `json:"items"`
}

// PodUsage represents resource usage statistics for a pod.
type PodUsage struct {
	PodID     string  `json:"podId"`
	CPU       float64 `json:"cpu"`
	Memory    float64 `json:"memory"`
	Disk      float64 `json:"disk"`
	Network   float64 `json:"network"`
	Uptime    int64   `json:"uptime"`
	CreatedAt string  `json:"createdAt"`
}

// PodsService provides methods for managing agent compute pods.
type PodsService struct {
	client *httpClient
}

// newPodsService creates a new PodsService.
func newPodsService(c *httpClient) *PodsService {
	return &PodsService{client: c}
}

// Create creates a new pod.
func (s *PodsService) Create(ctx context.Context, params CreatePodParams) (*Pod, error) {
	pod, err := Do[Pod](ctx, s.client, http.MethodPost, "/pods", params, nil)
	if err != nil {
		return nil, err
	}
	return &pod, nil
}

// List lists pods, optionally filtered by agent.
func (s *PodsService) List(ctx context.Context, params *PodListParams) (*PodList, error) {
	var q url.Values
	if params != nil {
		q = params.ToQuery()
	}
	list, err := Do[PodList](ctx, s.client, http.MethodGet, "/pods", nil, q)
	if err != nil {
		return nil, err
	}
	return &list, nil
}

// Get retrieves a pod by ID.
func (s *PodsService) Get(ctx context.Context, id string) (*Pod, error) {
	pod, err := Do[Pod](ctx, s.client, http.MethodGet, fmt.Sprintf("/pods/%s", id), nil, nil)
	if err != nil {
		return nil, err
	}
	return &pod, nil
}

// Update updates an existing pod.
func (s *PodsService) Update(ctx context.Context, id string, params UpdatePodParams) (*Pod, error) {
	pod, err := Do[Pod](ctx, s.client, http.MethodPut, fmt.Sprintf("/pods/%s", id), params, nil)
	if err != nil {
		return nil, err
	}
	return &pod, nil
}

// Delete deletes a pod.
func (s *PodsService) Delete(ctx context.Context, id string) error {
	_, err := Do[struct{}](ctx, s.client, http.MethodDelete, fmt.Sprintf("/pods/%s", id), nil, nil)
	return err
}

// Usage retrieves resource usage statistics for a pod.
func (s *PodsService) Usage(ctx context.Context, id string) (*PodUsage, error) {
	usage, err := Do[PodUsage](ctx, s.client, http.MethodGet, fmt.Sprintf("/pods/%s/usage", id), nil, nil)
	if err != nil {
		return nil, err
	}
	return &usage, nil
}
