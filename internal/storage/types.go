package storage

import (
	"context"

	"github.com/google/uuid"
)

// DocumentPayload represents a binary document ready for persistence.
type DocumentPayload struct {
	FileName    string
	ContentType string
	Data        []byte
}

// DocumentLocation captures the storage handle returned after a successful upload.
type DocumentLocation struct {
	Key    string
	Region string
}

// DocumentDownload holds the binary data and metadata fetched from storage.
type DocumentDownload struct {
	Data        []byte
	ContentType string
	Size        int64
}

// DocumentUploader persists and retrieves documents for parties.
type DocumentUploader interface {
	UploadPartyDocument(ctx context.Context, tenantID, tenantName string, partyID uuid.UUID, partyName, docType string, payload DocumentPayload) (DocumentLocation, error)
	DownloadDocument(ctx context.Context, region, key string) (DocumentDownload, error)
}
