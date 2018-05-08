package main

import (
	"os"
	"text/tabwriter"
	"time"

	"math/rand"

	"fmt"

	"sync"

	"sync/atomic"

	"github.com/Sirupsen/logrus"
	"github.com/spf13/cobra"
)

var maxConcurrent int
var fileSize int64
var filesDone int32
var fileCount int32
var seed int
var bucket, prefix string
var debug bool

type uploadFunc func(name string, data []byte)

type result struct {
	Size       int64
	UploadTime time.Duration
	GenTime    time.Duration
}

const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func main() {
	root := cobra.Command{}

	root.AddCommand(S3Cmd(), GoogleCmd())
	root.PersistentFlags().IntVarP(&maxConcurrent, "concurrent", "c", 10, "the max number of concurrent files to upload")
	root.PersistentFlags().Int64Var(&fileSize, "kb", 100, "the size of the data to create and upload")
	root.PersistentFlags().Int32VarP(&fileCount, "num", "n", 1, "the number of files to upload")
	root.PersistentFlags().IntVarP(&seed, "seed", "s", 0, "the seed to use for random")
	root.PersistentFlags().BoolVarP(&debug, "verbose", "v", false, "enable debug logging")
	root.PersistentFlags().StringVarP(&bucket, "bucket", "b", "nf-bench-test", "name of the bucket to upload to")
	root.PersistentFlags().StringVarP(&prefix, "prefix", "p", fmt.Sprintf("%d", time.Now().Unix()), "a prefix to use for this test run")

	if seed != 0 {
		rand.Seed(int64(seed))
	}

	if debug {
		logrus.SetLevel(logrus.DebugLevel)
	}

	if err := root.Execute(); err != nil {
		logrus.Fatal("Failed to execute command")
	}
}

func getDataBuffer(bytes int64) []byte {
	res := make([]byte, bytes)
	for i := range res {
		res[i] = chars[rand.Intn(len(chars))]
	}

	return res
}

func runTest(u uploadFunc) {
	wg := new(sync.WaitGroup)
	work := make(chan string, fileCount)
	results := make(chan *result, fileCount)
	for i := 0; i < maxConcurrent; i++ {
		wg.Add(1)
		log := logrus.WithField("worker_id", i)
		go performUploads(log, wg, work, results, u)
	}
	logrus.Infof("Started %d workers", maxConcurrent)

	complete := make(chan struct{})
	go displayResults(results, complete)

	for i := int32(0); i < fileCount; i++ {
		name := string(getDataBuffer(20))
		work <- fmt.Sprintf("%s/%s", prefix, name)
	}

	logrus.Infof("Enqueued all the work")
	close(work)

	logrus.Infof("Waiting for work to complete")
	wg.Wait()
	close(results)

	<-complete
}

func performUploads(log *logrus.Entry, wg *sync.WaitGroup, work chan string, results chan *result, u uploadFunc) {
	defer wg.Done()
	for fname := range work {
		bytesToMake := fileSize * 1024
		log.Debugf("generating data")

		genStart := time.Now()
		data := getDataBuffer(bytesToMake)
		genDur := time.Since(genStart)
		log.Debugf("generated data in %s", genDur.String())

		upstart := time.Now()
		u(fname, data)
		updur := time.Since(upstart)

		results <- &result{
			Size:       bytesToMake,
			UploadTime: updur,
			GenTime:    genDur,
		}
		fcount := atomic.AddInt32(&filesDone, 1)
		log.Infof("Finished uploading file %d/%d in %s", fcount, fileCount, updur.String())
	}
	log.Debug("Shutdown worker")
}

func displayResults(res chan *result, complete chan struct{}) {
	var numUploads int64
	var genTotal, upTotal time.Duration

	for r := range res {
		if r != nil {
			numUploads++
			genTotal = r.GenTime
			upTotal = r.UploadTime
		}
	}

	bytesSent := numUploads * fileSize * 1024
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', tabwriter.AlignRight|tabwriter.Debug)
	fmt.Println("-------------------------------------------------------------------------------")
	fmt.Fprintln(w, "num uploads\tbytes sent\tgen nanosec\tupload nanonsec")
	fmt.Fprintf(w, "% d\t% d\t% d\t% d\n", numUploads, bytesSent, genTotal.Nanoseconds(), upTotal.Nanoseconds())
	w.Flush()
	fmt.Println("-------------------------------------------------------------------------------")
	close(complete)
}
