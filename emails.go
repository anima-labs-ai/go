package anima

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// EmailListParams contains parameters for listing emails.
type EmailListParams struct {
	ListParams
	AgentID string
}

// ToQuery converts EmailListParams to URL query values.
func (p EmailListParams) ToQuery() url.Values {
	q := p.ListParams.ToQuery()
	if p.AgentID != "" {
		q.Set("agentId", p.AgentID)
	}
	return q
}

// UploadAttachmentParams contains parameters for uploading an attachment.
type UploadAttachmentParams struct {
	Filename  string `json:"filename"`
	MimeType  string `json:"mimeType"`
	SizeBytes int64  `json:"sizeBytes"`
}

// Attachment represents an uploaded attachment.
type Attachment struct {
	ID         string  `json:"id"`
	Filename   string  `json:"filename"`
	MimeType   string  `json:"mimeType"`
	SizeBytes  int64   `json:"sizeBytes"`
	StorageKey string  `json:"storageKey"`
	URL        *string `json:"url"`
	CreatedAt  string  `json:"createdAt"`
}

// AttachmentDownload contains a pre-signed URL for downloading an attachment.
type AttachmentDownload struct {
	URL       string `json:"url"`
	ExpiresAt string `json:"expiresAt"`
}

// EmailsService provides methods for managing emails and attachments.
type EmailsService struct {
	client *httpClient
}

// newEmailsService creates a new EmailsService.
func newEmailsService(c *httpClient) *EmailsService {
	return &EmailsService{client: c}
}

// List returns a paginated list of emails.
func (s *EmailsService) List(ctx context.Context, params *EmailListParams) (*Page[Message], error) {
	var q url.Values
	if params != nil {
		q = params.ToQuery()
	}
	page, err := Do[Page[Message]](ctx, s.client, http.MethodGet, "/email", nil, q)
	if err != nil {
		return nil, err
	}
	return &page, nil
}

// ListAutoPaging returns an iterator that automatically paginates through all emails.
func (s *EmailsService) ListAutoPaging(params *EmailListParams) *ListIterator[Message] {
	return NewListIterator(func(ctx context.Context, cursor string) (*Page[Message], error) {
		p := &EmailListParams{}
		if params != nil {
			*p = *params
		}
		p.Cursor = cursor
		return s.List(ctx, p)
	})
}

// UploadAttachment uploads an attachment to a message.
func (s *EmailsService) UploadAttachment(ctx context.Context, messageID string, params UploadAttachmentParams) (*Attachment, error) {
	body := struct {
		MessageID string `json:"messageId"`
		UploadAttachmentParams
	}{
		MessageID:              messageID,
		UploadAttachmentParams: params,
	}
	att, err := Do[Attachment](ctx, s.client, http.MethodPost, fmt.Sprintf("/messages/%s/attachments", messageID), body, nil)
	if err != nil {
		return nil, err
	}
	return &att, nil
}

// GetAttachmentURL retrieves a pre-signed download URL for an attachment.
func (s *EmailsService) GetAttachmentURL(ctx context.Context, attachmentID string) (*AttachmentDownload, error) {
	dl, err := Do[AttachmentDownload](ctx, s.client, http.MethodGet, fmt.Sprintf("/attachments/%s/download", attachmentID), nil, nil)
	if err != nil {
		return nil, err
	}
	return &dl, nil
}
