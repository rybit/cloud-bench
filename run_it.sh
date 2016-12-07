#!/bin/bash

bin="./cloud-bench-linux-0.0.2"

KEY_ID="---- YOUR VALUE ----"
ACCESS_KEY="---- YOUR VALUE ----"
SERVICE_FILE="---- YOUR VALUE ----"

echo "STARTING S3: 2MB && 100 blobs"
time $bin s3 --id $KEY_ID --key $ACCESS_KEY  --kb 2048  --num 100 -c 30
echo
echo "STARTING GOOGLE: 2MB && 100 blobs"
time $bin google netlify-cdn -a $SERVICE_FILE --kb 2048  --num 100 -c 30
echo
echo "STARTING S3: 2MB && 1000 blobs"
time $bin s3 --id $KEY_ID --key $ACCESS_KEY   --kb 2048  --num 1000 -c 30
echo
echo "STARTING GOOGLE: 2MB && 1000 blobs"
time $bin google netlify-cdn -a $SERVICE_FILE --kb 2048  --num 1000 -c 30
echo
echo "STARTING S3: 10MB && 100 blobs"
time $bin s3 --id $KEY_ID --key $ACCESS_KEY   --kb 10240 --num 100 -c 30
echo
echo "STARTING GOOGLE: 10MB && 100 blobs"
time $bin google netlify-cdn -a $SERVICE_FILE --kb 10240 --num 100 -c 30
echo
echo "STARTING S3: 10MB && 1000 blobs"
time $bin s3 --id $KEY_ID --key $ACCESS_KEY   --kb 10240 --num 1000 -c 30
echo
echo "STARTING GOOGLE: 10MB && 1000 blobs"
time $bin google netlify-cdn -a $SERVICE_FILE --kb 10240 --num 1000 -c 30

