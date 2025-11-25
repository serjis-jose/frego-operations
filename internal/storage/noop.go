package storage

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

var ErrUploaderDisabled = errors.New("storage: uploader disabled")

type noopUploader struct{}

// NewNoopUploader returns an uploader that reports a disabled storage backend.
func NewNoopUploader() DocumentUploader {
	return noopUploader{}
}

func (noopUploader) UploadPartyDocument(ctx context.Context, tenantID, tenantName string, partyID uuid.UUID, partyName, docType string, payload DocumentPayload) (DocumentLocation, error) {
	return DocumentLocation{}, ErrUploaderDisabled
}

func (noopUploader) DownloadDocument(ctx context.Context, region, key string) (DocumentDownload, error) {
	return DocumentDownload{}, ErrUploaderDisabled
}
