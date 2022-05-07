package bucket

import (
	"context"
	"fmt"
	"os"
	"time"

	"cloud.google.com/go/storage"
	"github.com/zicops/zicops-course-query/constants"
	"github.com/zicops/zicops-course-query/graph/model"
	"github.com/zicops/zicops-course-query/helpers"
	"google.golang.org/api/option"
)

// Client ....
type Client struct {
	projectID string
	client    *storage.Client
	bucket    *storage.BucketHandle
}

// NewStorageHandler return new database action
func NewStorageHandler() *Client {
	return &Client{projectID: "", client: nil}
}

// InitializeStorageClient ...........
func (sc *Client) InitializeStorageClient(ctx context.Context, projectID string) error {
	serviceAccountZicops := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	if serviceAccountZicops == "" {
		return fmt.Errorf("failed to get right credentials for course creator")
	}
	targetScopes := []string{
		"https://www.googleapis.com/auth/cloud-platform",
		"https://www.googleapis.com/auth/userinfo.email",
	}
	currentCreds, _, err := helpers.ReadCredentialsFile(ctx, serviceAccountZicops, targetScopes)
	if err != nil {
		return err
	}
	client, err := storage.NewClient(ctx, option.WithCredentials(currentCreds))
	if err != nil {
		return err
	}
	sc.client = client
	sc.projectID = projectID
	sc.bucket, _ = sc.CreateBucket(ctx, constants.COURSES_BUCKET)
	return nil
}

// CreateBucket  ...........
func (sc *Client) CreateBucket(ctx context.Context, bucketName string) (*storage.BucketHandle, error) {
	bkt := sc.client.Bucket(bucketName)
	exists, err := bkt.Attrs(ctx)
	if err != nil && exists == nil {
		if err := bkt.Create(ctx, sc.projectID, nil); err != nil {
			return nil, err
		}
	}
	return bkt, nil
}

// UploadToGCS ....
func (sc *Client) UploadToGCS(ctx context.Context, fileName string) (*storage.Writer, error) {
	bucketWriter := sc.bucket.Object(fileName).NewWriter(ctx)
	return bucketWriter, nil
}

func (sc *Client) GetSignedURLForObject(object string) string {
	opts := &storage.SignedURLOptions{
		Scheme:  storage.SigningSchemeV4,
		Method:  "GET",
		Expires: time.Now().Add(24 * time.Hour),
	}
	url, err := sc.bucket.SignedURL(object, opts)
	if err != nil {
		return ""
	}
	return url
}

func (sc *Client) GetSignedURLsForObjects(bucketPath string) []*model.SubtitleURL {
	opts := &storage.SignedURLOptions{
		Scheme:  storage.SigningSchemeV4,
		Method:  "GET",
		Expires: time.Now().Add(24 * time.Hour),
	}
	ctx := context.Background()

	objectsIter := sc.bucket.Objects(ctx, &storage.Query{
		Prefix:    bucketPath,
		Delimiter: "/",
	})
	// iterate over all objects in bucket
	var urls []*model.SubtitleURL

	for {
		obj, err := objectsIter.Next()
		if err != nil {
			break
		}
		objectBucketPath := fmt.Sprintf("%s/%s", obj.Bucket, obj.Name)
		url, err := sc.bucket.SignedURL(objectBucketPath, opts)
		if err != nil {
			break
		}
		language := ""
		if value, ok := obj.Metadata["language"]; ok {
			language = value
		}
		urls = append(urls, &model.SubtitleURL{URL: &url, Language: &language})
	}
	return urls
}
