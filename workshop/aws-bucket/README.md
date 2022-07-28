# Play with AWS Buckets

This example creates an AWS bucket, lists the buckets on your account and then deletes the test bucket.

Install the aws mixin
```
porter mixin install aws
```

# Credentials

```
$ porter credentials generate aws
```

For each credential, select `environment variable` as the source. This is what your credentials file should look like:

```yaml
name: aws
credentials:
- name: AWS_ACCESS_KEY_ID
  source:
    env: AWS_ACCESS_KEY_ID
- name: AWS_SECRET_ACCESS_KEY
  source:
    env: AWS_SECRET_ACCESS_KEY
```

# Try it out

## Create a bucket
```console
$ porter install --credential-set aws

installing porter-aws-bucket...
executing porter install configuration from /cnab/app/porter.yaml
Create Bucket
Starting operation: Create Bucket
{
    "Location": "/porter-aws-mixin-test"
}
Finished operation: Create Bucket
execution completed successfully!
```

## List buckets
```console
$ porter invoke --action list --credential-set aws

invoking custom action list on porter-aws-bucket...
executing porter list configuration from /cnab/app/porter.yaml
List Buckets
Starting operation: List Buckets
[
    "blog.sweetgeek.net",
    "porter-aws-mixin-test",
    "sweetgeek.net"
]
Finished operation: List Buckets
execution completed successfully!
```

## Delete a bucket
```console
$ porter uninstall --credential-set aws

uninstalling porter-aws-bucket...
executing porter uninstall configuration from /cnab/app/porter.yaml
Delete Bucket
execution completed successfully!
```
