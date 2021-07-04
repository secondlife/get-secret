# get-secret 

**get-secret** is a small program that gets secrets from AWS Secrets Manager and parameters from SSM Parameter Store.

## Use

```text
usage: get-secret [--ssm|--secretsmanager] NAME [VERSION]

Fetch values from AWS Secrets Manager and SSM Parameter Store.

positional arguments:
  NAME     secret or parameter name
  VERSION  secret version, used by Secrets Manager only. Default = AWSCURRENT

optional arguments:
  --secretsmanager use AWS Secrets Manager (default)
  --ssm            use SSM Parameter Store
```

### Notes

If you are attempting to **get-secrets** on a machine with AWS credentials from the environment, such as when using bobafett or awsume, then you must set `AWS_SDK_LOAD_CONFIG` to a truthy value for credentials loading to work. See [sdk-for-go's session documentation][session] for more information.

[session]: https://docs.aws.amazon.com/sdk-for-go/api/aws/session/
