package pkg

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

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
	client, err := initVaultClient(vaultMock.URL, roleId, secretId)
	assert.Nil(t, err)
	assert.Equal(t, mockedToken, client.Token())
}
