# cloud-bench

This is a simple benchmarking tool for s3 and google storage. It will generate buffers of a fixed size with random payloads and then upload them to the specified provider.


## auth
In the s3 case you need to specify the AWS_ACCESS_KEY_ID and AWS_ACCESS_KEY in either the environment or on the command line.
In the google case you have to specify the project id and optionally the service account's json key

```
$ go run *.go s3 --id $AWS_ACCESS_KEY_ID --key $AWS_ACCESS_KEY
$ go run *.go google my-project-id -a service-account.json
```

