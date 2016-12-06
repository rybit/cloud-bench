# cloud-bench

This is a simple benchmarking tool for s3 and google storage. It will generate buffers of a fixed size with random payloads and then upload them to the specified provider.


```
$ go run *.go
Usage:
   [command]

   Available Commands:
     google      google <project_id>
       s3

       Flags:
         -b, --bucket string    name of the bucket to upload to (default "nf-bench-test")
         -c, --concurrent int   the max number of concurrent files to upload (default 10)
             --kb int           the size of the data to create and upload (default 100)
         -n, --num int32        the number of files to upload (default 1)
         -s, --seed int         the seed to use for random (default 0)

Use " [command] --help" for more information about a command.
```

## auth
In the s3 case you need to specify the AWS_ACCESS_KEY_ID and AWS_ACCESS_KEY in either the environment or on the command line.
In the google case you have to specify the project id and optionally the service account's json key

```
$ go run *.go s3 --id $AWS_ACCESS_KEY_ID --key $AWS_ACCESS_KEY
$ go run *.go google my-project-id -a service-account.json
```

## running
You can easily configure the size, concurrency and number of files. If you set a seed you can reliably upload the same files. By default it will not use any value, making each run unique

