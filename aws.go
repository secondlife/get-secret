package main

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"os"
	"sync"
)

var (
	once sync.Once
	sess *session.Session
)

// Singleton AWS Session
func GetAwsSession() *session.Session {
	once.Do(func() {
		sess = CreateAwsSession()
	})
	return sess
}

func CreateAwsSession() *session.Session {
	// Attempt to resolve the AWS region. Region may not be automatically
	// resolved in some environments such as in a container running on ECS.
	region, regionPresent := os.LookupEnv("AWS_REGION")
	if !regionPresent {
		region, regionPresent = os.LookupEnv("AWS_DEFAULT_REGION")
	}
	if !regionPresent {
		meta := ec2metadata.New(session.New())
		if meta.Available() {
			region, _ = meta.Region()
		}
	}

	return session.New(&aws.Config{
		Region: aws.String(region),
	})
}
