---
title: aws mixin
description: Run Amazon commands using the aws CLI.
---

<img src="/images/mixins/aws.svg" class="mixin-logo"/>

Run Amazon commands using the [aws CLI](https://docs.aws.amazon.com/cli/latest/reference/index.html#cli-aws).

Source: https://github.com/getporter/aws-mixin

### Install or Upgrade
```
porter mixin install aws --version v1.0.0-rc.1
```

### Mixin Syntax

See the [AWS CLI Command Reference](https://docs.aws.amazon.com/cli/latest/reference/index.html#cli-aws) for the supported commands

```yaml
aws:
  description: "Description of the command"
  service: SERVICE
  operation: OP
  arguments:
  - arg1
  - arg2
  flags:
    FLAGNAME: FLAGVALUE
    REPEATED_FLAG:
    - FLAGVALUE1
    - FLAGVALUE2
  suppress-output: false
  outputs:
  - name: NAME
    jsonPath: JSONPATH
```

### Suppress Output

The `suppress-output` field controls whether output from the mixin should be
prevented from printing to the console. By default this value is false, using
Porter's default behavior of hiding known sensitive values. When 
`suppress-output: true` all output from the mixin (stderr and stdout) are hidden.

Step outputs (below) are still collected when output is suppressed. This allows
you to prevent sensitive data from being exposed while still collecting it from
a command and using it in your bundle.

### Outputs

The mixin supports `jsonpath` outputs.


#### JSON Path

The `jsonPath` output treats stdout like a json document and applies the expression, saving the result to the output.

```yaml
outputs:
- name: NAME
  jsonPath: JSONPATH
```

For example, if the `jsonPath` expression was `$[*].id` and the command sent the following to stdout: 

```json
[
  {
    "id": "1085517466897181794",
    "name": "my-vm"
  }
]
```

Then then output would have the following contents:

```json
["1085517466897181794"]
```

---

### Examples

The [Buckets Example](https://github.com/getporter/aws-mixin/tree/master/examples/buckets) provides a full working bundle demonstrating how to use this mixin.

#### Provision a VM

```yaml
aws:
  description: "Provision VM"
  service: ec2
  operation: run-instances
  flags:
    image-id: ami-xxxxxxxx
    instance-type: t2.micro
```

#### Create a Bucket

```yaml
aws:
  description: "Create Bucket"
  service: s3api
  operation: create-bucket
  flags:
    bucket: my-buckkit
    region: us-east-1
```