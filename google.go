package main

import (
	"context"

	"fmt"

	"cloud.google.com/go/storage"
	"github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

var jsonCredsFile string

func GoogleCmd() *cobra.Command {

	gcsCmd := &cobra.Command{
		Use:   "google",
		Short: "google <project_id>",
		Run:   uploadToGoogle,
	}

	gcsCmd.Flags().StringVarP(&jsonCredsFile, "auth", "a", "", "the service role json config file")

	return gcsCmd
}

func uploadToGoogle(cmd *cobra.Command, args []string) {
	if len(args) != 1 {
		logrus.Fatal("Must provide the project id to use")
	}

	opts := []option.ClientOption{}
	if jsonCredsFile != "" {
		opts = append(opts, option.WithServiceAccountFile(jsonCredsFile))
	}

	projectID := args[0]
	ctx := context.Background()
	client, err := storage.NewClient(ctx, opts...)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to configure storage client")
	}
	logrus.Info("Connected to google API")

	logrus.Info("Listing out buckets")
	found := false
	iter := client.Buckets(ctx, projectID)
	iter.Prefix = bucket

	for b, err := iter.Next(); err != iterator.Done && err == nil; b, err = iter.Next() {
		fmt.Printf("bucket %+v\n", b)
		fmt.Printf("err %+v\n", err)
		if b.Name == bucket {
			found = true
			break
		}
	}
	if err != nil && err != iterator.Done {
		logrus.WithError(err).Fatal("Failed to walk buckets")
	}

	b := client.Bucket(bucket)
	if !found {
		if err := b.Create(ctx, projectID, nil); err != nil {
			logrus.WithError(err).Fatalf("Failed to create bucket %s in project %s", bucket, projectID)
		}
	}

	results := uploadData(func(key string, data []byte) error {
		obj := b.Object(key)
		w := obj.NewWriter(ctx)
		if _, err := fmt.Fprintf(w, string(data)); err != nil {
			return err
		}

		return w.Close()
	})

	displayResults(results)
}
