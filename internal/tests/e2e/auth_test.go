package e2e

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type AuthRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Token string `json:"token"`
}

func TestRegister(t *testing.T) {
	err := env.resetDB()
	require.NoError(t, err)
	bodyBytes, err := json.Marshal(AuthRequest{
		Username: "testUser",
		Password: "testPass",
	})
	require.NoError(t, err)

	resp, err := http.Post(
		env.serverURL+"/api/auth",
		"application/json",
		bytes.NewReader(bodyBytes),
	)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	respParsed := AuthResponse{}
	err = json.NewDecoder(resp.Body).Decode(&respParsed)
	require.NoError(t, err)
}

func TestAuth(t *testing.T) {
	err := env.resetDB()
	require.NoError(t, err)

	testUsername, testUserPass := "testuser", "testpass"

	bodyBytes, err := json.Marshal(AuthRequest{
		Username: testUsername,
		Password: testUserPass,
	})
	require.NoError(t, err)

	resp, err := http.Post(
		env.serverURL+"/api/auth",
		"application/json",
		bytes.NewReader(bodyBytes),
	)
	require.NoError(t, err)
	resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	testcases := []struct {
		name           string
		authBody       AuthRequest
		wantAuthStatus int
	}{
		{
			name:           "Success",
			authBody:       AuthRequest{Username: testUsername, Password: testUserPass},
			wantAuthStatus: http.StatusOK,
		},
		{
			name:           "InvalidPassword",
			authBody:       AuthRequest{Username: testUsername, Password: "invalid"},
			wantAuthStatus: http.StatusUnauthorized,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			bodyBytes, err := json.Marshal(tc.authBody)
			assert.NoError(t, err)
			resp, err := http.Post(
				env.serverURL+"/api/auth",
				"application/json",
				bytes.NewReader(bodyBytes),
			)
			assert.NoError(t, err)
			defer resp.Body.Close()
			assert.Equal(t, tc.wantAuthStatus, resp.StatusCode)
		})
	}
}
