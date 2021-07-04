package main

import "testing"

const UnknownSource SecretSource = 50

// MockProvider captures input/output for test assertions.
type MockProvider struct {
	Input     *GetSecretInput
	Output    GetSecretOutput
	OutputErr *error
}

func CreateMockProvider() *MockProvider {
	secret := "secret"
	return &MockProvider{
		Output:    GetSecretOutput{String: &secret},
		Input:     nil,
		OutputErr: nil,
	}
}

func (p *MockProvider) GetSecret(i GetSecretInput) (*GetSecretOutput, error) {
	p.Input = &i
	return &p.Output, nil
}

func TestRunWithNoArguments(t *testing.T) {
	args := []string{"get-secret"}
	if run(args, CreateMockProvider()) != 2 {
		t.Errorf("run() did not exit with status 2 when no parameter were given.")
	}
}

func TestRunUsesSecretsManager(t *testing.T) {
	args := []string{"get-secret", "secret"}
	provider := CreateMockProvider()
	status := run(args, provider)

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
	status = run(args, provider)

	if provider.Input.Version != "version" {
		t.Errorf("run() did not pass version")
	}
}

func TestRunUsesParameterStore(t *testing.T) {
	args := []string{"get-secret", "--ssm", "secret"}
	provider := CreateMockProvider()
	status := run(args, provider)

	if provider.Input.Source != ParameterStore {
		t.Errorf("run() did not use SSM Parameter Store when passed --ssm.")
	}

	if status != 0 {
		t.Errorf("run() did not exit successfully.")
	}
}
