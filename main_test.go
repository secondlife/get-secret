package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"testing"
	"text/template"
)

// MockProvider captures input/output for test assertions.
type MockProvider struct {
	Input   *GetSecretInput
	Output  []byte
	Secrets map[string]string
}

func CreateMockProvider() *MockProvider {
	return &MockProvider{
		Input: nil,
		Secrets: map[string]string{
			"secret-secretsmanager": "secret",
			"secret-ssm":            "secret-from-ssm",
		},
	}
}

func (p *MockProvider) GetSecret(i GetSecretInput) ([]byte, error) {
	p.Input = &i
	if secret, ok := p.Secrets[fmt.Sprintf("%s-%s", i.Name, i.Source)]; ok {
		return []byte(secret), nil
	}
	return nil, errors.New("Secret not found")
}

func TestRunWithNoArguments(t *testing.T) {
	buf := new(bytes.Buffer)
	flag.CommandLine.SetOutput(buf)
	args := []string{"get-secret"}
	if run(args, buf, CreateMockProvider()) != 2 {
		t.Errorf("run() did not exit with status 2 when no parameter were given.")
	}
}

func TestRunUsesSecretsManager(t *testing.T) {
	args := []string{"get-secret", "secret"}
	provider := CreateMockProvider()
	buf := new(bytes.Buffer)
	status := run(args, buf, provider)

	if provider.Input.Source != SecretsManager {
		t.Errorf("run() did not use Secrets Manager by default.")
	}

	if provider.Input.Version != "AWSCURRENT" {
		t.Errorf("run() did not pass AWSCURRENT version by default")
	}

	if status != 0 {
		t.Errorf("run() did not exit successfully.")
	}

	args = []string{"get-secret", "secret", "version"}
	buf = new(bytes.Buffer)
	status = run(args, buf, provider)

	if status != 0 {
		t.Errorf("run() did not exit successfully.")
	}

	if provider.Input.Version != "version" {
		t.Errorf("run() did not pass version")
	}
}

func TestRunUsesParameterStore(t *testing.T) {
	args := []string{"get-secret", "--ssm", "secret"}
	provider := CreateMockProvider()
	buf := new(bytes.Buffer)
	status := run(args, buf, provider)

	if provider.Input.Source != ParameterStore {
		t.Errorf("run() did not use SSM Parameter Store when passed --ssm.")
	}

	if status != 0 {
		t.Errorf("run() did not exit successfully.")
	}
}

func TestFromConf(t *testing.T) {
	provider := &MockProvider{
		Input: nil,
		Secrets: map[string]string{
			"secret1-secretsmanager": "secret1",
			"secret2-ssm":            "secret2",
		},
	}
	loader := &SecretLoader{provider}

	usr, err := user.Current()
	if err != nil {
		t.Error(err)
	}

	confTmpl := `
secret1 {{.Temp}}/secret1 {{.User}} {{.Group}} 0644
secret2 {{.Temp}}/secret2 {{.User}} {{.Group}} 0644 ssm`

	tmpl, err := template.New("conf").Parse(confTmpl)
	if err != nil {
		t.Error(err)
	}

	tmp, err := os.MkdirTemp(os.TempDir(), "get-secret-test-*")
	if err != nil {
		t.Error(err)
	}

	username := usr.Username
	test_user := os.Getenv("TEST_USER")
	if test_user != "" {
		username = test_user
	}

	buf := new(bytes.Buffer)
	err = tmpl.Execute(buf, struct {
		Temp  string
		User  string
		Group string
	}{
		Temp:  tmp,
		User:  username,
		Group: usr.Gid,
	})
	if err != nil {
		t.Error(err)
	}

	defer func() {
		if err = os.RemoveAll(tmp); err != nil {
			panic(err)
		}
	}()

	// Write conf to a file
	conf := filepath.Join(tmp, "secrets.conf")
	if err = os.WriteFile(conf, buf.Bytes(), 0644); err != nil {
		t.Error(err)
	}

	if err = loader.FromFileConf(conf); err != nil {
		t.Error(err)
	}

	for _, f := range []string{"secret1", "secret2"} {
		s, err := os.ReadFile(filepath.Join(tmp, f))
		if err != nil {
			t.Error(err)
		}
		if string(s) != f {
			t.Errorf("Expected %s to contain %s but got %s", f, f, s)
		}
	}
}
