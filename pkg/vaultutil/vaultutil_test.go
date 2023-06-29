package vaultutil

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	vault "github.com/hashicorp/vault/api"
	"github.com/stretchr/testify/assert"
)

func TestInitVaultClient(t *testing.T) {
	mockedToken := "65b74ffd-842c-fd43-1386-f7d7006e520a"
	vaultMock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.Path, "auth/approle/login")
		body, err := io.ReadAll(r.Body)
		assert.Nil(t, err)
		assert.Equal(t, `{"role_id":"foo","secret_id":"bar"}`, string(body))
		fmt.Fprintf(w, `{"auth": {"client_token": "%s"}}`, mockedToken)
	}))
	defer vaultMock.Close()

	roleId := "foo"
	secretId := "bar"
	client, err := InitVaultClient(vaultMock.URL, roleId, secretId)
	assert.Nil(t, err)
	assert.Equal(t, mockedToken, client.Token())
}

func TestGetVaultTfSecret(t *testing.T) {
	mockedData := `{
		"data": {
			"data": {
			  	"aws_access_key_id": "foo",
				"aws_secret_access_key": "bar",
				"region": "weast",
				"bucket": "head"
			},
			"metadata": {}
		}
}`
	vaultMock := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.Path, "v1/terraform/data/stage")
		assert.Equal(t, "3", r.URL.Query().Get("version"))
		fmt.Fprint(w, mockedData)
	}))
	defer vaultMock.Close()

	client, _ := vault.NewClient(&vault.Config{
		Address: vaultMock.URL,
	})

	actual, err := GetVaultTfSecret(client, VaultSecret{
		Path:    "terraform/stage",
		Version: 3,
	}, KV_V2)
	assert.Nil(t, err)

	expected := VaultKvData{
		"aws_access_key_id":     "foo",
		"aws_secret_access_key": "bar",
		"region":                "weast",
		"bucket":                "head",
	}

	assert.Equal(t, expected, actual)
}
