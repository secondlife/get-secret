package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/aws/aws-sdk-go/service/ssm"
)

const usage = `usage: get-secret [--ssm|--secretsmanager|--conf] NAME [VERSION]

Fetch values from AWS Secrets Manager and SSM Parameter Store.

positional arguments:
  NAME     secret, parameter or conf file name
  VERSION  secret version, used by Secrets Manager only. Default = AWSCURRENT

optional arguments:
  --secretsmanager use AWS Secrets Manager (default)
  --ssm            use SSM Parameter Store
  --conf           load secrets from configuration file ("-" for stdin)
  --env-conf       load secrets from environment variable
  -v               show verbose logging 
  
configuration file example:
  # source_path        destination_path                   owner group    permissions source_service
  /mitra/myapp/secrets /etc/secrets-internal/secrets.json root  www-data 0640
  /mitra/myapp/param   /etc/secrets-internal/param.txt    root  www-data 0640        ssm
`

type SecretProvider interface {
	GetSecret(i GetSecretInput) ([]byte, error)
}

type SecretsManagerProvider struct{}
type ParameterStoreProvider struct{}
type CombinedProvider struct{}

func (p *SecretsManagerProvider) GetSecret(i GetSecretInput) ([]byte, error) {
	svc := secretsmanager.New(GetAwsSession())
	input := &secretsmanager.GetSecretValueInput{
		SecretId:     aws.String(i.Name),
		VersionStage: aws.String(i.Version),
	}
	log.Printf("secretsmanager: getting %s", i.Name)
	res, err := svc.GetSecretValue(input)

	if err != nil {
		return nil, err
	}

	if res.SecretString != nil {
		return []byte(*res.SecretString), nil
	}
	return res.SecretBinary, nil
}

func (p *ParameterStoreProvider) GetSecret(i GetSecretInput) ([]byte, error) {
	svc := ssm.New(GetAwsSession())
	input := &ssm.GetParameterInput{
		Name:           aws.String(i.Name),
		WithDecryption: aws.Bool(true),
	}
	log.Printf("ssm: getting %s", i.Name)
	res, err := svc.GetParameter(input)

	if err != nil {
		return nil, err
	}

	return []byte(*res.Parameter.Value), nil
}

func (p *CombinedProvider) GetSecret(i GetSecretInput) ([]byte, error) {
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

const (
	AWSCURRENT     = "AWSCURRENT"
	SecretsManager = "secretsmanager"
	ParameterStore = "ssm"
	Combined       = "combined"
)

type GetSecretInput struct {
	Name    string
	Version string
	Source  string
}

type CopySecretInput struct {
	Name   string
	Dest   string
	Uid    string
	Gid    string
	Mode   int64
	Source string
}

// ConfLineFromStr parses whitespace separated CopySecretInput fields from a string.
// This is used as part of get-secret configuration file parsing.
func CopySecretInputFromStr(s string) (*CopySecretInput, error) {
	fields := strings.Fields(s)

	if len(fields) == 0 {
		return nil, nil
	}

	if len(fields) < 5 {
		return nil, fmt.Errorf("ConfLineFromStr: Incorrect number of fields. Need: NAME DST OWNER GROUP PERMISSIONS [SOURCE]")
	}

	// Default to secretsmanager
	source := SecretsManager
	if len(fields) > 5 {
		source = fields[5]
	}

	// Parse octal
	mode, err := strconv.ParseInt(fields[4], 8, 64)
	if err != nil {
		return nil, err
	}

	// Lookup user by name or id
	username := fields[2]
	owner, err := user.Lookup(username)
	switch err.(type) {
	case nil:
		break
	case user.UnknownUserError:
		if owner, err = user.LookupId(username); err != nil {
			if _, ok := err.(*strconv.NumError); ok {
				return nil, user.UnknownUserError(username)
			}
			return nil, err
		}
	default:
		return nil, err
	}

	// Lookup group by name or id
	groupname := fields[3]
	group, err := user.LookupGroup(groupname)
	switch err.(type) {
	case nil:
		break
	case user.UnknownGroupError:
		if group, err = user.LookupGroupId(groupname); err != nil {
			if _, ok := err.(*strconv.NumError); ok {
				return nil, user.UnknownGroupError(groupname)
			}
			return nil, err
		}
	default:
		return nil, err
	}

	return &CopySecretInput{
		Name:   fields[0],
		Dest:   fields[1],
		Uid:    owner.Uid,
		Gid:    group.Gid,
		Mode:   mode,
		Source: source,
	}, nil
}

type SecretLoader struct {
	provider SecretProvider
}

func (s *SecretLoader) FromEnvConf(name string) error {
	conf := os.Getenv(name)
	return s.FromConf(strings.NewReader(conf))
}

func (s *SecretLoader) FromFileConf(path string) error {
	log.Printf("Loading configuration from %s", path)
	if path == "-" {
		return s.FromConf(os.Stdin)
	} else {
		r, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("Unable to open %s", path)
		}
		defer r.Close()
		return s.FromConf(r)
	}
}

// Copy a secret from secretsmanager or ssm to a local file with the specified permissions.
func (s *SecretLoader) CopySecret(in CopySecretInput) error {
	// Pull secret
	res, err := s.provider.GetSecret(GetSecretInput{
		Name:    in.Name,
		Version: AWSCURRENT,
		Source:  in.Source,
	})
	if err != nil {
		return err
	}

	// Write to file
	f, err := os.Create(in.Dest)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err = f.Write(res); err != nil {
		return err
	}

	// Set permissions
	if err = chown(in.Dest, in.Uid, in.Gid); err != nil {
		return err
	}
	if err = os.Chmod(in.Dest, fs.FileMode(in.Mode)); err != nil {
		return err
	}

	return nil
}

func (s *SecretLoader) FromConf(conf io.Reader) error {
	lnNum := 0
	secretNum := 0
	scanner := bufio.NewScanner(conf)
	for scanner.Scan() {
		lnNum++
		ln := scanner.Text()
		log.Println(ln)
		// Skip comments
		if strings.HasPrefix(ln, "#") {
			continue
		}

		in, err := CopySecretInputFromStr(ln)
		if err != nil {
			return err
		}
		// Skip empty lines
		if in == nil {
			continue
		}

		if err = s.CopySecret(*in); err != nil {
			return err
		}

		secretNum++
	}

	if secretNum > 0 {
		log.Printf("Loaded %d secret(s)", secretNum)
	} else {
		log.Printf("No secrets to load")
	}

	return nil
}

func run(args []string, out io.Writer, provider SecretProvider) int {
	f := flag.NewFlagSet(args[0], flag.ExitOnError)
	f.Usage = func() {
		if _, err := flag.CommandLine.Output().Write([]byte(usage)); err != nil {
			panic(err)
		}
	}

	// Help text not supplied as we are using a custom usage function.
	useSsm := f.Bool("ssm", false, "")
	fileConf := f.Bool("conf", false, "")
	envConf := f.Bool("env-conf", false, "")
	verbose := f.Bool("v", false, "")
	f.Bool("secretsmanager", true, "")

	err := f.Parse(args[1:])
	if err != nil {
		log.Fatal(err)
	}

	if *verbose {
		log.SetOutput(out)
	} else {
		log.SetOutput(ioutil.Discard)
	}

	narg := f.NArg()
	if narg < 1 {
		f.Usage()
		return 2
	}

	name := f.Arg(0)

	if *envConf {
		l := &SecretLoader{provider}
		if err = l.FromEnvConf(name); err != nil {
			log.Fatal(err)
		}
	} else if *fileConf {
		l := &SecretLoader{provider}
		if err = l.FromFileConf(name); err != nil {
			log.Fatal(err)
		}
	} else {
		version := AWSCURRENT
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
			log.Fatal(err)
		}

		if _, err := out.Write(res); err != nil {
			log.Fatal(err)
		}
	}
	return 0
}

func main() {
	os.Exit(run(os.Args, os.Stdout, &CombinedProvider{}))
}
