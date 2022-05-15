package main

import (
	"log"
	"os"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
)

var (
	once sync.Once
	sess *session.Session
)

// Singleton AWS Session
func GetAwsSession() *session.Session {
	once.Do(func() {
		log.Println("creating AWS session")
		var err error
		sess, err = CreateAwsSession()
		if err != nil {
			panic(err)
		}
		log.Printf("session created in %s", *sess.Config.Region)
	})
	return sess
}

func CreateAwsSession() (*session.Session, error) {
	// Attempt to resolve the AWS region. Region may not be automatically
	// resolved in some environments such as in a container running on ECS.
	region, regionPresent := os.LookupEnv("AWS_REGION")
	if !regionPresent {
		region, regionPresent = os.LookupEnv("AWS_DEFAULT_REGION")
	}
	if !regionPresent {
		sess, err := session.NewSession()
		if err != nil {
			return sess, err
		}
		meta := ec2metadata.New(sess)
		if meta.Available() {
			region, _ = meta.Region()
		}
	}

	return session.NewSession(&aws.Config{
		Region: aws.String(region),
	})
}
