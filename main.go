package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/aws/aws-sdk-go/service/ssm"
)

const usage = `usage: get-secret [--ssm|--secretsmanager] NAME [VERSION]

Fetch values from AWS Secrets Manager and SSM Parameter Store.

positional arguments:
  NAME     secret or parameter name
  VERSION  secret version, used by Secrets Manager only. Default = AWSCURRENT

optional arguments:
  --secretsmanager use AWS Secrets Manager (default)
  --ssm            use SSM Parameter Store`

type SecretProvider interface {
	GetSecret(i GetSecretInput) (*GetSecretOutput, error)
}

type SecretsManagerProvider struct{}
type ParameterStoreProvider struct{}
type CombinedProvider struct{}

func (p *SecretsManagerProvider) GetSecret(i GetSecretInput) (*GetSecretOutput, error) {
	svc := secretsmanager.New(GetAwsSession())
	input := &secretsmanager.GetSecretValueInput{
		SecretId:     aws.String(i.Name),
		VersionStage: aws.String(i.Version),
	}
	res, err := svc.GetSecretValue(input)

	if err != nil {
		return nil, err
	}

	return &GetSecretOutput{
		Binary: res.SecretBinary,
		String: res.SecretString,
	}, nil
}

func (p *ParameterStoreProvider) GetSecret(i GetSecretInput) (*GetSecretOutput, error) {
	svc := ssm.New(GetAwsSession())
	input := &ssm.GetParameterInput{
		Name:           aws.String(i.Name),
		WithDecryption: aws.Bool(true),
	}
	res, err := svc.GetParameter(input)

	if err != nil {
		return nil, err
	}

	return &GetSecretOutput{
		String: res.Parameter.Value,
	}, nil
}

func (p *CombinedProvider) GetSecret(i GetSecretInput) (*GetSecretOutput, error) {
	switch i.Source {
	case SecretsManager:
		return (&SecretsManagerProvider{}).GetSecret(i)
	case ParameterStore:
		return (&ParameterStoreProvider{}).GetSecret(i)
	case Combined:
		return nil, errors.New("CombinedProvider.GetSecret called recursively.")
	default:
		return nil, fmt.Errorf("Unknown SecretSource %v", i.Source)
	}
}

type SecretSource int

const (
	SecretsManager SecretSource = iota
	ParameterStore
	Combined
)

type GetSecretInput struct {
	Name    string
	Version string
	Source  SecretSource
}

type GetSecretOutput struct {
	String *string
	Binary []byte
}

func run(args []string, provider SecretProvider) int {
	f := flag.NewFlagSet(args[0], flag.ExitOnError)
	f.Usage = func() {
		fmt.Println(usage)
	}

	// Help text not supplied as we are using a custom usage function.
	useSsm := f.Bool("ssm", false, "")
	f.Bool("secretsmanager", true, "")

	err := f.Parse(args[1:])
	if err != nil {
		panic(err)
	}

	narg := f.NArg()
	if narg < 1 {
		f.Usage()
		return 2
	}

	name := f.Arg(0)
	version := "AWSCURRENT"
	source := SecretsManager

	if narg > 1 {
		version = f.Arg(1)
	}

	if *useSsm {
		source = ParameterStore
	}

	res, err := provider.GetSecret(GetSecretInput{
		Name:    name,
		Version: version,
		Source:  source,
	})

	if err != nil {
		fmt.Println(err.Error())
		return 1
	}

	if res.Binary != nil && len(res.Binary) > 0 {
		os.Stdout.Write(res.Binary)
	} else if res.String != nil {
		fmt.Println(*res.String)
	} else {
		fmt.Println("Secret response had no value.")
		return 1
	}
	return 0
}

func main() {
	os.Exit(run(os.Args, &CombinedProvider{}))
}
