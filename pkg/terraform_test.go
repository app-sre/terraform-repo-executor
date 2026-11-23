package pkg

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/app-sre/terraform-repo-executor/pkg/vaultutil"
	"github.com/hashicorp/terraform-exec/tfexec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	accessKey   = "fake_access_key"
	secretKey   = "fake_secret_key"
	region      = "us-east-1"
	bucket      = "app-sre"
	repoName    = "a-repo"
	repoURL     = "https://gitlab.myinstance.com/some-gl-group/project_a"
	repoPath    = "prod/networking"
	repoRef     = "d82b3cb292d91ec2eb26fc282d751555088819f3"
	awsCredPath = "terraform/creds/prod-account"
	tfVersion   = "1.5.7"
)

var repoWithoutExplicitBucketSettings = Repo{
	Name:   repoName,
	URL:    repoURL,
	Path:   repoPath,
	Ref:    repoRef,
	Delete: false,
	AWSCreds: vaultutil.VaultSecret{
		Path:    awsCredPath,
		Version: 4,
	},
	RequireFips: false,
	TfVersion:   tfVersion,
}

func TestExtractTfCreds(t *testing.T) {
	exampleTfCreds := vaultutil.VaultKvData{
		AwsAccessKeyID:     accessKey,
		AwsSecretAccessKey: secretKey,
		AwsRegion:          region,
		AwsBucket:          bucket,
	}
	t.Run("credentials are extracted successfully when all fields present", func(t *testing.T) {
		expected := TfCreds{
			AccessKey: accessKey,
			SecretKey: secretKey,
			Region:    region,
			Bucket:    bucket,
		}

		creds, err := extractTfCreds(exampleTfCreds, repoWithoutExplicitBucketSettings)

		assert.Nil(t, err)
		assert.Equal(t, expected, creds)
	})

	t.Run("bucket credentials are extracted successfully when explicitly set in the repo config", func(t *testing.T) {
		explicitRegion := "us-west-1"
		explicitBucket := "sre-of-apps"

		expected := TfCreds{
			AccessKey: accessKey,
			SecretKey: secretKey,
			Region:    explicitRegion,
			Bucket:    explicitBucket,
		}

		repoWithExplicitBucketSettings := Repo{
			Name:   repoName,
			URL:    repoURL,
			Path:   repoPath,
			Ref:    repoRef,
			Delete: false,
			AWSCreds: vaultutil.VaultSecret{
				Path:    awsCredPath,
				Version: 4,
			},
			RequireFips: false,
			TfVersion:   tfVersion,
			Bucket:      explicitBucket,
			Region:      explicitRegion,
		}

		creds, err := extractTfCreds(exampleTfCreds, repoWithExplicitBucketSettings)

		assert.Nil(t, err)
		assert.Equal(t, expected, creds)
	})

	t.Run("error out when vault secret is missing some required keys", func(t *testing.T) {
		secretMissingKeys := vaultutil.VaultKvData{
			AwsAccessKeyID:     accessKey,
			AwsSecretAccessKey: secretKey,
		}

		_, err := extractTfCreds(secretMissingKeys, repoWithoutExplicitBucketSettings)

		assert.Error(t, err)
	})

}

func TestWriteTemplate(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "template")

	assert.Nil(t, err)

	// create the directory that would normally be created when cloning the repo
	repoDir := fmt.Sprintf("%s/%s/%s", tmpDir,
		repoWithoutExplicitBucketSettings.Name,
		repoWithoutExplicitBucketSettings.Path)
	err = os.MkdirAll(repoDir, FolderPerm)

	assert.Nil(t, err)

	defer os.RemoveAll(tmpDir)

	t.Run("generate a backend file successfully", func(t *testing.T) {
		backendCreds := TfCreds{
			AccessKey: accessKey,
			SecretKey: secretKey,
			Region:    region,
			Key:       "a-repo-tf-repo.tfstate",
			Bucket:    bucket,
		}
		body := `access_key = "{{.AccessKey}}"
			{{- "\n"}}secret_key = "{{.SecretKey}}"
			{{- "\n"}}region = "{{.Region}}"
			{{- "\n"}}key = "{{.Key}}"
			{{- "\n"}}bucket = "{{.Bucket}}"`

		err := WriteTemplate(backendCreds, body, fmt.Sprintf("%s/%s/%s/%s", tmpDir, repoWithoutExplicitBucketSettings.Name, repoWithoutExplicitBucketSettings.Path, BackendFile))

		assert.Nil(t, err)

		expected, err := os.ReadFile("../test/data/s3.tfbackend")
		assert.Nil(t, err)

		output, err := os.ReadFile(fmt.Sprintf("%s/%s", repoDir, BackendFile))
		assert.Nil(t, err)

		assert.Equal(t, string(expected), string(output))
	})

	t.Run("generate an input tfvars file successfully", func(t *testing.T) {
		inputSecret := vaultutil.VaultKvData{
			"foo": "bar",
			"bar": "bell",
		}
		body := `{{ range $k, $v := . }}{{ $k }} = "{{ $v }}"{{- "\n"}}{{ end }}`

		err := WriteTemplate(inputSecret, body, fmt.Sprintf("%s/%s/%s/%s", tmpDir, repoWithoutExplicitBucketSettings.Name, repoWithoutExplicitBucketSettings.Path, InputVarsFile))

		assert.Nil(t, err)

		expected, err := os.ReadFile("../test/data/input.auto.tfvars")
		assert.Nil(t, err)

		output, err := os.ReadFile(fmt.Sprintf("%s/%s", repoDir, InputVarsFile))
		assert.Nil(t, err)

		assert.Equal(t, string(expected), string(output))
	})
}

const fipsPlanJSONStdoutCanary = "fips-compliance-check-plan-json-canary"

func writeFakeTerraformBinary(t *testing.T, dir string) string {
	t.Helper()

	script := fmt.Sprintf(`#!/bin/sh
case "$1" in
version)
	if [ "$2" = "-json" ]; then
		printf '%%s\n' '{"terraform_version":"1.5.7","platform":"linux_amd64","provider_selections":{},"terraform_outdated":false}'
		exit 0
	fi
	echo "Terraform v1.5.7"
	exit 0
	;;
show)
	if [ "$2" = "-json" ]; then
		cat <<'PLAN_EOF'
{
  "format_version": "1.2",
  "terraform_version": "1.5.7",
  "sensitive_plan_leak_canary": %q,
  "configuration": {
    "provider_config": {
      "aws": {
        "name": "aws",
        "expressions": {
          "use_fips_endpoint": {
            "constant_value": true
          }
        }
      }
    }
  }
}
PLAN_EOF
		exit 0
	fi
	;;
esac
exit 1
`, fipsPlanJSONStdoutCanary)

	binaryPath := filepath.Join(dir, "terraform")
	require.NoError(t, os.WriteFile(binaryPath, []byte(script), 0o755))
	return binaryPath
}

func TestFipsComplianceCheckDoesNotWritePlanJSONToStdout(t *testing.T) {
	workDir := t.TempDir()
	fakeTerraform := writeFakeTerraformBinary(t, workDir)

	planFile := filepath.Join(workDir, "tfplan")
	require.NoError(t, os.WriteFile(planFile, []byte{}, 0o644))

	tf, err := tfexec.NewTerraform(workDir, fakeTerraform)
	require.NoError(t, err)
	tf.SetEnv(map[string]string{})

	var stdoutCapture strings.Builder
	tf.SetStdout(&stdoutCapture)

	repo := Repo{
		Name:        repoName,
		RequireFips: true,
	}

	executor := &Executor{}
	err = executor.fipsComplianceCheck(repo, planFile, tf)
	require.NoError(t, err)
	assert.NotContains(t, stdoutCapture.String(), fipsPlanJSONStdoutCanary)
}
