# get-secret

**get-secret** is a small program that gets secrets from AWS Secrets Manager
and parameters from SSM Parameter Store. It can be useful in containers and
other environments where a self-contained tool for working with AWS secrets
is desired.

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
  --env-conf       load secrets from environment variable
  -v               show verbose logging

configuration file example:
  # source_path        destination_path                   owner group    permissions source_service
  /mitra/myapp/secrets /etc/secrets-internal/secrets.json root  www-data 0640
  /mitra/myapp/param   /etc/secrets-internal/param.txt    root  www-data 0640        ssm
```
