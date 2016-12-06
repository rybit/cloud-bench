package main

import (
	"sync"

	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/spf13/cobra"
)

var awsRegion string
var awsID string
var awsSecret string
var useEnvCreds bool

func S3Cmd() *cobra.Command {
	s3Cmd := &cobra.Command{
		Use: "s3",
		Run: uploadToS3,
	}

	s3Cmd.Flags().StringVarP(&awsRegion, "region", "r", "us-west-1", "the region to upload to")
	s3Cmd.Flags().StringVarP(&awsID, "id", "i", "", "the aws_access_key_id")
	s3Cmd.Flags().StringVarP(&awsSecret, "key", "k", "", "the aws_secret_access_key")
	s3Cmd.Flags().BoolVarP(&useEnvCreds, "env", "e", false, "if we should use creds from the environment")

	return s3Cmd
}

type sharedErr struct {
	sync.Mutex
	err error
}

func (s *sharedErr) setError(err error) {
	s.Lock()
	if s.err == nil {
		s.err = err
	}
	s.Unlock()
}

func (s *sharedErr) hasError() bool {
	s.Lock()
	defer s.Unlock()

	return s.err != nil
}

func uploadToS3(cmd *cobra.Command, args []string) {
	var creds *credentials.Credentials
	if useEnvCreds {
		creds = credentials.NewEnvCredentials()
	} else {
		if awsID == "" || awsSecret == "" {
			logrus.Fatal("aws secret and id must be specified or use the env flag")
		}
		creds = credentials.NewStaticCredentials(awsID, awsSecret, "")
	}

	config := aws.Config{
		Credentials: creds,
		Region:      &awsRegion,
	}
	svc := s3.New(session.New(&config))
	logrus.Debug("connected to s3 service")

	if _, err := svc.HeadBucket(&s3.HeadBucketInput{Bucket: &bucket}); err != nil {
		logrus.Info("Assuming the bucket doesn't exist - going to try and build it")

		if _, err := svc.CreateBucket(&s3.CreateBucketInput{Bucket: &bucket}); err != nil {
			logrus.WithError(err).Fatalf("Failed to create bucket %s", bucket)
		}
	}

	logrus.Infof("Waiting for the bucket %s to exist", bucket)
	if err := svc.WaitUntilBucketExists(&s3.HeadBucketInput{Bucket: &bucket}); err != nil {
		logrus.WithError(err).Fatal("Failed to wait for bucket creation")
	}

	results := uploadData(func(key string, data []byte) error {
		l := logrus.WithField("worker_id", key)
		l.Info("Starting to upload to s3")
		buf := strings.NewReader(string(data))
		dataLen := int64(len(data))
		_, err := svc.PutObject(&s3.PutObjectInput{
			Body:          buf,
			ContentLength: &dataLen,
			Bucket:        &bucket,
			Key:           &key,
		})
		return err
	})

	displayResults(results)
}
