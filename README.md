# get-secret 

**get-secret** is a small program that gets secrets from AWS Secrets Manager
and parameters from SSM Parameter Store.

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
  --conf           load secrets from configuration file ("-" for stdin)
  -v               show verbose logging 

configuration file example:
  # source_path        destination_path                   owner group    permissions source_service
  /mitra/myapp/secrets /etc/secrets-internal/secrets.json root  www-data 0640
  /mitra/myapp/param   /etc/secrets-internal/param.txt    root  www-data 0640        ssm`
```

### Notes

If you are attempting to **get-secrets** on a machine with AWS credentials from
the environment, such as when using aws sso or awsume, then you must set
`AWS_SDK_LOAD_CONFIG` to a truthy value for credentials loading to work. See
[sdk-for-go's session documentation][session] for more information.

[session]: https://docs.aws.amazon.com/sdk-for-go/api/aws/session/
