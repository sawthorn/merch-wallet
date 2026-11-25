//go:build e2e

package e2e

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/sawthorn/merch-wallet/internal/info"
	"github.com/stretchr/testify/require"
)

func TestPurchase(t *testing.T) {
	err := env.resetDB()
	require.NoError(t, err)
	userToken := makeUserAuthReq(t, env.serverURL, "testuser", "testpass")

	itemName := "cup"
	url := env.serverURL + "/api/buy/" + itemName

	req, err := http.NewRequest(http.MethodGet, url, nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+userToken)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	req, err = http.NewRequest(
		http.MethodGet,
		env.serverURL+"/api/info",
		nil,
	)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+userToken)

	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	infoResp := info.InfoResponse{}
	err = json.NewDecoder(resp.Body).Decode(&infoResp)
	require.NoError(t, err)
	want := []info.InventoryItemResponse{
		{
			Name:     itemName,
			Quantity: 1,
		},
	}
	require.Equal(t, want, infoResp.Inventory)
}

func TestPurchase_ItemNotFound(t *testing.T) {
	err := env.resetDB()
	require.NoError(t, err)
	userToken := makeUserAuthReq(t, env.serverURL, "testuser", "testpass")

	itemName := "doesnotexist"
	url := env.serverURL + "/api/buy/" + itemName

	req, err := http.NewRequest(http.MethodGet, url, nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+userToken)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	resp.Body.Close()
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}
