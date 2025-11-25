package storage

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"path"
	"path/filepath"
	"strings"
	"time"
	"unicode"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/google/uuid"
)

// S3Config defines the fields required to initialise an S3-backed uploader.
type S3Config struct {
	Bucket          string
	Region          string
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	UsePathStyle    bool
	KeyPrefix       string
}

// S3Uploader uploads party documents to an S3 compatible object store.
type S3Uploader struct {
	client *s3.Client
	bucket string
	prefix string
	region string
}

// NewS3Uploader creates a configured uploader or returns an error if initialisation fails.
func NewS3Uploader(ctx context.Context, cfg S3Config) (*S3Uploader, error) {
	if strings.TrimSpace(cfg.Bucket) == "" {
		return nil, fmt.Errorf("storage: s3 uploader: bucket is required")
	}
	if strings.TrimSpace(cfg.Region) == "" {
		return nil, fmt.Errorf("storage: s3 uploader: region is required")
	}

	loadOptions := []func(*awsconfig.LoadOptions) error{
		awsconfig.WithRegion(cfg.Region),
	}

	if strings.TrimSpace(cfg.Endpoint) != "" {
		endpoint := strings.TrimSpace(cfg.Endpoint)
		loadOptions = append(loadOptions, awsconfig.WithEndpointResolverWithOptions(
			aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
				return aws.Endpoint{
					PartitionID:   "aws",
					URL:           endpoint,
					SigningRegion: cfg.Region,
				}, nil
			}),
		))
	}

	if strings.TrimSpace(cfg.AccessKeyID) != "" && strings.TrimSpace(cfg.SecretAccessKey) != "" {
		provider := credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.SecretAccessKey, "")
		loadOptions = append(loadOptions, awsconfig.WithCredentialsProvider(provider))
	}

	awsCfg, err := awsconfig.LoadDefaultConfig(ctx, loadOptions...)
	if err != nil {
		return nil, fmt.Errorf("storage: s3 uploader: load config: %w", err)
	}

	clientOptions := func(o *s3.Options) {
		o.UsePathStyle = cfg.UsePathStyle
	}

	client := s3.NewFromConfig(awsCfg, clientOptions)

	prefix := strings.Trim(strings.TrimSpace(cfg.KeyPrefix), "/")

	return &S3Uploader{
		client: client,
		bucket: cfg.Bucket,
		prefix: prefix,
		region: cfg.Region,
	}, nil
}

// UploadPartyDocument uploads the provided payload and returns the resulting object key.
func (u *S3Uploader) UploadPartyDocument(ctx context.Context, tenantID, tenantName string, partyID uuid.UUID, partyName, docType string, payload DocumentPayload) (DocumentLocation, error) {
	if len(payload.Data) == 0 {
		return DocumentLocation{}, fmt.Errorf("storage: s3 uploader: payload is empty")
	}
	fileName := sanitizeFileName(payload.FileName)
	if fileName == "" {
		fileName = "document.bin"
	}

	key := u.buildObjectKey(tenantID, tenantName, partyID, partyName, docType, fileName)

	contentType := strings.TrimSpace(payload.ContentType)
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	_, err := u.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(u.bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(payload.Data),
		ContentType: aws.String(contentType),
		Metadata: map[string]string{
			"doc_type": docType,
		},
	})
	if err != nil {
		return DocumentLocation{}, fmt.Errorf("storage: s3 uploader: put object: %w", err)
	}

	return DocumentLocation{
		Key:    key,
		Region: u.region,
	}, nil
}

// DownloadDocument retrieves a document by key ensuring it resides in the configured region.
func (u *S3Uploader) DownloadDocument(ctx context.Context, region, key string) (DocumentDownload, error) {
	key = strings.TrimSpace(key)
	if key == "" {
		return DocumentDownload{}, fmt.Errorf("storage: s3 uploader: key is required")
	}

	if trimmed := strings.TrimSpace(region); trimmed != "" && !strings.EqualFold(trimmed, strings.TrimSpace(u.region)) {
		return DocumentDownload{}, fmt.Errorf("storage: s3 uploader: mismatched region '%s'", trimmed)
	}

	output, err := u.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(u.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		var notFound *s3types.NoSuchKey
		if errors.As(err, &notFound) {
			return DocumentDownload{}, ErrObjectNotFound
		}
		return DocumentDownload{}, fmt.Errorf("storage: s3 uploader: get object: %w", err)
	}
	defer output.Body.Close()

	data, err := io.ReadAll(output.Body)
	if err != nil {
		return DocumentDownload{}, fmt.Errorf("storage: s3 uploader: read object: %w", err)
	}

	contentType := "application/octet-stream"
	if output.ContentType != nil {
		if trimmed := strings.TrimSpace(*output.ContentType); trimmed != "" {
			contentType = trimmed
		}
	}
	var size int64
	if output.ContentLength != nil {
		size = *output.ContentLength
	}

	return DocumentDownload{
		Data:        data,
		ContentType: contentType,
		Size:        size,
	}, nil
}

func (u *S3Uploader) buildObjectKey(tenantID, tenantName string, partyID uuid.UUID, partyName, docType, fileName string) string {
	tenantSlug := slugifyName(tenantName, tenantID)
	partySlug := slugifyName(partyName, partyID.String())
	docSlug := slugifyName(docType, "document")

	unique := uuid.New().String()
	safeFile := sanitizeFileName(fileName)
	if safeFile == "" {
		safeFile = fmt.Sprintf("%s.bin", docSlug)
	}
	objectName := fmt.Sprintf("%s-%s", unique, safeFile)
	now := time.Now().UTC()

	parts := []string{
		"tenants",
		fmt.Sprintf("%s-%s", tenantSlug, tenantID),
		"parties",
		fmt.Sprintf("%s-%s", partySlug, partyID.String()),
		docSlug,
		fmt.Sprintf("%04d", now.Year()),
		fmt.Sprintf("%02d", now.Month()),
		fmt.Sprintf("%02d", now.Day()),
		objectName,
	}

	if u.prefix != "" {
		parts = append([]string{u.prefix}, parts...)
	}

	return path.Join(parts...)
}

func sanitizePathToken(value string) string {
	clean := strings.TrimSpace(value)
	clean = strings.ReplaceAll(clean, "..", "")
	clean = strings.ReplaceAll(clean, "/", "")
	clean = strings.ReplaceAll(clean, "\\", "")
	if clean == "" {
		return "unknown"
	}
	return clean
}

func sanitizeFileName(name string) string {
	clean := strings.TrimSpace(name)
	clean = filepath.Base(clean)
	clean = strings.ReplaceAll(clean, "..", "")
	if clean == "." || clean == string(filepath.Separator) || clean == "" {
		return ""
	}
	return clean
}

func slugifyName(value, fallback string) string {
	input := strings.TrimSpace(value)
	if input == "" {
		input = fallback
	}
	input = strings.ToLower(input)

	var builder strings.Builder
	lastDash := false
	for _, r := range input {
		switch {
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			builder.WriteRune(r)
			lastDash = false
		case r == ' ' || r == '-' || r == '_' || r == '.':
			if !lastDash && builder.Len() > 0 {
				builder.WriteRune('-')
				lastDash = true
			}
		default:
			// skip unsafe characters
		}
	}

	slug := strings.Trim(builder.String(), "-")
	if slug == "" {
		slug = sanitizePathToken(fallback)
	}
	if slug == "" {
		slug = "tenant"
	}
	return slug
}
