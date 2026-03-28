package anima

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// A2ATaskStatus represents the status of an A2A task.
type A2ATaskStatus string

const (
	A2ATaskStatusPending    A2ATaskStatus = "pending"
	A2ATaskStatusRunning    A2ATaskStatus = "running"
	A2ATaskStatusCompleted  A2ATaskStatus = "completed"
	A2ATaskStatusFailed     A2ATaskStatus = "failed"
	A2ATaskStatusCancelled  A2ATaskStatus = "cancelled"
)

// A2ATask represents an agent-to-agent task.
type A2ATask struct {
	ID        string                 `json:"id"`
	AgentID   string                 `json:"agentId"`
	Status    A2ATaskStatus          `json:"status"`
	Input     any                    `json:"input,omitempty"`
	Output    any                    `json:"output,omitempty"`
	Error     *A2ATaskError          `json:"error,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt string                 `json:"createdAt"`
	UpdatedAt string                 `json:"updatedAt"`
}

// A2ATaskError represents an error that occurred during task execution.
type A2ATaskError struct {
	Code    string `json:"code,omitempty"`
	Message string `json:"message"`
}

// SubmitA2ATaskParams contains the parameters for submitting an A2A task.
type SubmitA2ATaskParams struct {
	Input    any                    `json:"input"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// A2ATaskListParams contains parameters for listing A2A tasks.
type A2ATaskListParams struct {
	ListParams
	Status A2ATaskStatus
}

// ToQuery converts A2ATaskListParams to URL query values.
func (p A2ATaskListParams) ToQuery() url.Values {
	q := p.ListParams.ToQuery()
	if p.Status != "" {
		q.Set("status", string(p.Status))
	}
	return q
}

// A2AService provides methods for agent-to-agent (A2A) communication.
type A2AService struct {
	client *httpClient
}

// newA2AService creates a new A2AService.
func newA2AService(c *httpClient) *A2AService {
	return &A2AService{client: c}
}

// SubmitTask submits a new task to an agent via the A2A protocol.
func (s *A2AService) SubmitTask(ctx context.Context, agentID string, params SubmitA2ATaskParams) (*A2ATask, error) {
	task, err := Do[A2ATask](ctx, s.client, http.MethodPost, fmt.Sprintf("/agents/%s/a2a/tasks", agentID), params, nil)
	if err != nil {
		return nil, err
	}
	return &task, nil
}

// GetTask retrieves an A2A task by ID.
func (s *A2AService) GetTask(ctx context.Context, agentID string, taskID string) (*A2ATask, error) {
	task, err := Do[A2ATask](ctx, s.client, http.MethodGet, fmt.Sprintf("/agents/%s/a2a/tasks/%s", agentID, taskID), nil, nil)
	if err != nil {
		return nil, err
	}
	return &task, nil
}

// ListTasks returns a paginated list of A2A tasks for an agent.
func (s *A2AService) ListTasks(ctx context.Context, agentID string, params *A2ATaskListParams) (*Page[A2ATask], error) {
	var q url.Values
	if params != nil {
		q = params.ToQuery()
	}
	page, err := Do[Page[A2ATask]](ctx, s.client, http.MethodGet, fmt.Sprintf("/agents/%s/a2a/tasks", agentID), nil, q)
	if err != nil {
		return nil, err
	}
	return &page, nil
}

// ListTasksAutoPaging returns an iterator that automatically paginates through all A2A tasks.
func (s *A2AService) ListTasksAutoPaging(agentID string, params *A2ATaskListParams) *ListIterator[A2ATask] {
	return NewListIterator(func(ctx context.Context, cursor string) (*Page[A2ATask], error) {
		p := &A2ATaskListParams{}
		if params != nil {
			*p = *params
		}
		p.Cursor = cursor
		return s.ListTasks(ctx, agentID, p)
	})
}

// CancelTask cancels an in-progress A2A task.
func (s *A2AService) CancelTask(ctx context.Context, agentID string, taskID string) (*A2ATask, error) {
	task, err := Do[A2ATask](ctx, s.client, http.MethodPost, fmt.Sprintf("/agents/%s/a2a/tasks/%s/cancel", agentID, taskID), nil, nil)
	if err != nil {
		return nil, err
	}
	return &task, nil
}
