# Terraform-Repo Executor

An application for performing Terraform operations on targeted git repositories.

Terraform Repo executor takes input from a corresponding [Qontract Reconcile integration](https://github.com/app-sre/qontract-reconcile/blob/master/reconcile/terraform_repo.py) and uses that input to manage the lifecycle of a repository of raw HCL/Terraform definitions through App Interface.

## Configuration

## Environment Variables

* **Required**
  * `VAULT_ADDR` - http address of Vault instance to retrieve/write secrets to
  * `VAULT_ROLE_ID` - used for [AppRole auth](https://developer.hashicorp.com/vault/docs/auth/approle)
  * `VAULT_SECRET_ID`- used for [AppRole auth](https://developer.hashicorp.com/vault/docs/auth/approle)
* **Optional**
  * `CONFIG_FILE` - input/config file location, defaults to `/config.yaml`
  * `WORKDIR` - working directory for tf operations, defaults to `/tf-repo`
  * `VAULT_TF_KV_VERSION` - defaults to `KV_V2`. Specifies which version of the [Vault `kv` secrets engine to use for reading/writing secrets](https://developer.hashicorp.com/vault/docs/secrets/kv)

## Config file

The application processes the yaml/json defined at `CONFIG_FILE` for determining targets. [The schema for this file is defined in QR](https://github.com/app-sre/qontract-reconcile/blob/master/reconcile/terraform_repo.py#L56).

* `dry-run`: *boolean* - if `true`, the application executes `terraform plan`; if `false`, the application executes `terraform apply`.
* `repos`: *list(Repo)* - a list of tf-repo targets. Below attributes comprise a tf-repo object:
  * `repository`: *string* - URL of Git repository
  * `name`: *string* - custom name for the repository, used as an identifier throughout the application
  * `ref`: *string* - commit sha in the repository to be targeted
  * `project_path`: *string* - Terraform Git repositories can include multiple Terraform root modules in one repo so this path defines [where the provider and other required files for this repo are located](https://developer.hashicorp.com/terraform/language/providers/configuration)
  * `delete`: *boolean* - if `true`, the application will execute the Terraform action with the [`destroy` flag](https://developer.hashicorp.com/terraform/cli/commands/destroy) set
  * `require_fips`: *boolean* - if `true` then the executor will validate the generated plan to ensure that AWS is using FIPS endpoints
  * `bucket`: *string* - optional S3 bucket name to store Terraform state in. If not specified then the executor will try to extract this from `aws_creds` Vault secret
  * `bucket_path`: *string* - optional path of where to store specific Terraform state files in `bucket`
  * `region`: *string* - optional AWS region of where the `bucket` is stored
  * `tf_version`: *string* - required, determines which tf binary to run, full enumeration in [schemas](https://github.com/app-sre/qontract-schemas/blob/main/schemas/aws/terraform-repo-1.yml#L37-L40)
  * `aws_creds`: *AWSCreds* - reference to a Vault secret including credentials for accessing the [S3 state backend for Terraform](https://developer.hashicorp.com/terraform/language/settings/backends/s3). Attributes defined below:
    * `path`: *string* - path to the secret in the vault. For KV v2, do not include the hidden `data` path segment
    * `version`: *integer* - for KV2 engine, defines which version of secret to read, ignored for KV1 engines as they don't have a concept of secret versioning
  * `variables`: *Variables* - optionally defines Vault paths to [read inputs, write outputs to](https://developer.hashicorp.com/terraform/language/values)
    * `inputs`: *Inputs*
      * `path`: *string* - path in vault to read from
      * `version`: *integer* - which version of secret to read (ignored for KV1 vault)
    * `outputs`: *Outputs*
      * `path`: *string* - path in vault to write to

Note that this file is auto generated by the Qontract Reconcile integration.

### Example

```yaml
dry_run: true
repos: 
- repository: https://gitlab.myinstance.com/some-gl-group/project_a
  name: foo-foo
  ref: d82b3cb292d91ec2eb26fc282d751555088819f3
  project_path: prod/networking
  delete: false
  aws_creds:
    path: terraform/creds/prod-acount
    version: 4
  variables:
    inputs:
      path: terraform/inputs/foo-foo
    outputs:
      path: terraform/outputs/foo-foo
- repository: https://gitlab.myinstance.com/another-gl-group/project_b
  name: bar-bar
  ref: 47ef09135da2d158ede78dbbe8c59de1775a274c
  project_path: stage/rds
  delete: false
  aws_creds:
    path: terraform/creds/stage-account
    version: 1
  bucket: bar-bar-backend
  bucket_path: bar
  region: us-east-1
```
