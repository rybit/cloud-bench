# cloud-bench

This is a simple benchmarking tool for s3 and google storage. It will generate buffers of a fixed size with random payloads and then upload them to the specified provider.


## auth
In the s3 case you need to specify the AWS_ACCESS_KEY_ID and AWS_ACCESS_KEY in either the environment or on the command line.
In the google case you have to specify the project id and optionally the service account's json key

```
$ go run *.go s3 --id $AWS_ACCESS_KEY_ID --key $AWS_ACCESS_KEY
$ go run *.go google my-project-id -a service-account.json
```

## test procedure
This program was used to upload files to S3 and GCS from EC2 and GCP. It was done both across clouds and in the same cloud. In both cases we used uploaded 10,000 files, 100 at a time. We chose 100 kb & 1000 kb files because the majority of files we handle are small, text doesn't take much space.

Machine specs:
- 2 core
- 4 GB memory
- large disks
- west coast region

The bucket in GCS was a multi-regional. The bucket in AWS was in the same region as the EC2 instance.

### results

Cross Cloud

| from | to  | concurrency | kb/file | # files | generation ns  | upload ns     |
|------|-----|-------------|---------|---------|----------------|---------------|
| gce  | s3  | 100         | 100     | 10000   | 5329173054488  | 1566737502449 |
| gce  | s3  | 100         | 1000    | 10000   | 61293681801836 | 2618097434728 |
| ec2  | gcs | 100         | 100     | 10000   | 3649485826597  | 3408973228791 |
| ec2  | gcs | 100         | 1000    | 10000   | 67847580371015 | 4151692912402 |

Same Cloud

| from | to  | concurrency | kb/file | # files | generation ns  | upload ns     |
|------|-----|-------------|---------|---------|----------------|---------------|
| gce  | gcs | 100         | 100     | 10000   | 3753495038963  | 3334422390625 |
| gce  | gcs | 100         | 1000    | 10000   | 59550670730615 | 3442640357291 |
| ec2  | s3  | 100         | 100     | 10000   | 6363562062251  | 751838563352  |
| ec2  | s3  | 100         | 1000    | 10000   | 69511973302665 | 1360465822324 |
