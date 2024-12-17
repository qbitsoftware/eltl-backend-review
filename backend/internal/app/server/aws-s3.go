package server

import (
	"context"
	"errors"
	"log"
	"mime/multipart"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"
)

type Bucket struct {
	S3Client *s3.Client
}

func newS3Client() *Bucket {
	if os.Getenv(awsAccessKey) == "" {
		log.Printf("No aws access key set!")
	}
	if os.Getenv(awsSecretKey) == "" {
		log.Printf("No aws secret key set!")
	}
	if os.Getenv(awsRegion) == "" {
		log.Printf("No aws region set!")
	}
	if os.Getenv(awsUploadBucket) == "" {
		log.Printf("No aws upload bucket set!")
	}
	if os.Getenv(awsRetrieveBucket) == "" {
		log.Printf("No aws retrieve bucket set!")
	}
	if os.Getenv(awsRetrieveBucketUrl) == "" {
		log.Printf("No aws retrieve bucket url set!")
	}
	if os.Getenv(awsUploadBucketUrl) == "" {
		log.Printf("No aws upload bucket url set!")
	}

	client := s3.New(s3.Options{
		Region:      os.Getenv(awsRegion),
		Credentials: aws.NewCredentialsCache(credentials.NewStaticCredentialsProvider(strings.Trim(os.Getenv(awsAccessKey), " "), strings.Trim(os.Getenv(awsSecretKey), " "), "")),
	})
	return &Bucket{
		S3Client: client,
	}
}

func (b *Bucket) UploadFile(ctx context.Context, bucketName string, objectKey string, file multipart.File) error {
	_, err := b.S3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(objectKey),
		Body:   file,
	})
	if err != nil {
		var apiErr smithy.APIError
		if errors.As(err, &apiErr) && apiErr.ErrorCode() == "EntityTooLarge" {
			log.Printf("Error while uploading object to %s. The object is too large.\n"+
				"To upload objects larger than 5GB, use the S3 console (160GB max)\n"+
				"or the multipart upload API (5TB max).", bucketName)
		} else {
			log.Printf("Couldn't upload file %v to %v:%v. Here's why: %v\n",
				"test", bucketName, objectKey, err)
		}
	} else {
		err = s3.NewObjectExistsWaiter(b.S3Client).Wait(
			ctx, &s3.HeadObjectInput{Bucket: aws.String(bucketName), Key: aws.String(objectKey)}, time.Minute)
		if err != nil {
			log.Printf("Failed attempt to wait for object %s to exist.\n", objectKey)
		}
	}
	return err
}

func (b Bucket) ListObjects(ctx context.Context, prefix string) ([]types.Object, error) {
	var err error
	var output *s3.ListObjectsV2Output
	var objects []types.Object
	input := &s3.ListObjectsV2Input{
		Bucket: aws.String(os.Getenv(awsRetrieveBucket)),
		Prefix: aws.String(prefix),
	}
	objectPaginator := s3.NewListObjectsV2Paginator(b.S3Client, input)
	for objectPaginator.HasMorePages() {
		output, err = objectPaginator.NextPage(ctx)
		if err != nil {
			var noBucket *types.NoSuchBucket
			if errors.As(err, &noBucket) {
				log.Printf("Bucket %s does not exist.\n", strings.Trim(os.Getenv(awsRetrieveBucket), " "))
				err = noBucket
			}
			break
		} else {
			objects = append(objects, output.Contents...)
		}
	}
	return objects, err
}

func (b Bucket) DeleteObject(ctx context.Context, key string, versionId string, bypassGovernance bool) (bool, error) {
	deleted := false
	bucketName := os.Getenv(awsUploadBucketUrl)
	input := &s3.DeleteObjectInput{
		Bucket: aws.String(strings.Trim(os.Getenv(awsRetrieveBucketUrl), " ")),
		Key:    aws.String(key),
	}
	if versionId != "" {
		input.VersionId = aws.String(versionId)
	}
	if bypassGovernance {
		input.BypassGovernanceRetention = aws.Bool(true)
	}
	_, err := b.S3Client.DeleteObject(ctx, input)
	if err != nil {
		var noKey *types.NoSuchKey
		var apiErr *smithy.GenericAPIError
		if errors.As(err, &noKey) {
			log.Printf("Object %s does not exist in %s.\n", key, bucketName)
			err = noKey
		} else if errors.As(err, &apiErr) {
			switch apiErr.ErrorCode() {
			case "AccessDenied":
				log.Printf("Access denied: cannot delete object %s from %s.\n", key, bucketName)
				err = nil
			case "InvalidArgument":
				if bypassGovernance {
					log.Printf("You cannot specify bypass governance on a bucket without lock enabled.")
					err = nil
				}
			}
		}
	} else {
		err = s3.NewObjectNotExistsWaiter(b.S3Client).Wait(
			ctx, &s3.HeadObjectInput{Bucket: aws.String(bucketName), Key: aws.String(key)}, time.Minute)
		if err != nil {
			log.Printf("Failed attempt to wait for object %s in bucket %s to be deleted.\n", key, bucketName)
		} else {
			deleted = true
		}
	}
	return deleted, err
}
