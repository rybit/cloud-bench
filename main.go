package main

import (
	"time"

	"math/rand"

	"fmt"

	"sync"

	"github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"
)

var maxConcurrent int
var fileSize int64
var fileCount int32
var seed int
var bucket string

func main() {
	root := cobra.Command{}

	root.AddCommand(S3Cmd(), GoogleCmd())
	root.PersistentFlags().IntVarP(&maxConcurrent, "concurrent", "c", 10, "the max number of concurrent files to upload")
	root.PersistentFlags().Int64Var(&fileSize, "kb", 100, "the size of the data to create and upload")
	root.PersistentFlags().Int32VarP(&fileCount, "num", "n", 1, "the number of files to upload")
	root.PersistentFlags().IntVarP(&seed, "seed", "s", 0, "the seed to use for random")
	root.PersistentFlags().StringVarP(&bucket, "bucket", "b", "nf-bench-test", "name of the bucket to upload to")

	if seed != 0 {
		rand.Seed(int64(seed))
	}

	if err := root.Execute(); err != nil {
		logrus.Fatal("Failed to execute command")
	}
}

type result struct {
	Size       int64
	UploadTime time.Duration
	GenTime    time.Duration
}

const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func getDataBuffer(bytes int64) []byte {
	res := make([]byte, bytes)
	for i := range res {
		res[i] = chars[rand.Intn(len(chars))]
	}

	return res
}

func displayResults(res []*result) {
	total := time.Duration(0)
	uploads := int64(0)
	genTotal := time.Duration(0)
	bs := int64(0)

	for _, r := range res {
		if r == nil {
			continue
		}
		uploads++
		bs += r.Size
		total += r.UploadTime
		genTotal += r.GenTime
	}
	avg := time.Duration(total.Nanoseconds() / uploads)
	avgGenTime := time.Duration(genTotal.Nanoseconds() / uploads)
	bpsec := float32(bs) / float32(total.Seconds())

	fmt.Printf("Completed %d uploads (%d bytes) in %s, avg %s\n", uploads, bs, total.String(), avg.String())
	fmt.Printf("Average: %d nsec/file, %.2f bytes/sec\n", avg.Nanoseconds(), bpsec)
	fmt.Printf("Generation time was %s, avg %s", genTotal.String(), avgGenTime.String())
}

type uploadFunc func(name string, data []byte) error

func uploadData(u uploadFunc) []*result {
	globalStart := time.Now()

	sem := make(chan bool, maxConcurrent)
	wg := new(sync.WaitGroup)
	shared := new(sharedErr)
	results := make([]*result, fileCount)
	bytesToMake := fileSize * 1024

	logrus.Infof("Starting to upload %d files of %d bytes", fileCount, bytesToMake)
	for i := range results {
		wg.Add(1)
		fileID := i
		go func() {
			sem <- true
			defer func() {
				wg.Done()
				<-sem
			}()
			key := string(getDataBuffer(20))
			l := logrus.WithField("worker_id", key)
			if shared.hasError() {
				l.Debug("Skipping because of previous error")
				return
			}

			l.Info("generating data")
			genStart := time.Now()
			data := getDataBuffer(bytesToMake)
			genDur := time.Since(genStart)
			l.Infof("generated data in %s", genDur.String())
			start := time.Now()

			if err := u(key, data); err != nil {
				l.WithError(err).Error("Error while uploading file")
				return
			}

			dur := time.Since(start)
			l.Infof("Finished uploading file %d/%d in %s", fileID+1, len(results), dur.String())
			results[fileID] = &result{
				Size:       bytesToMake,
				UploadTime: dur,
				GenTime:    genDur,
			}
		}()
	}

	logrus.Infof("Launched workers")
	wg.Wait()
	dur := time.Since(globalStart)
	logrus.Infof("Completed workers in %s", dur.String())
	return results
}
