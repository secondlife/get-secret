package main

import (
	"context"
	"log"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
)

var (
	once sync.Once
	cfg aws.Config
)

// GetAwsConfig returns a singleton AWS config for use with AWS services.
func GetAwsConfig() aws.Config {
	once.Do(func() {
		log.Println("creating AWS config")
		var err error
		cfg, err = config.LoadDefaultConfig(
			context.TODO(),
			config.WithDefaultRegion("us-west-2"),
		)
		if err != nil {
			panic(err)
		}
		log.Printf("config created for %s", cfg.Region)
	})
	return cfg
}
