package utils

import (
	"os"
	"time"
)

var Now = time.Now

const BaseGcsUrl = "https://storage.googleapis.com"

var BucketName = os.Getenv("BUCKET_NAME")
