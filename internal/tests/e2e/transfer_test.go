package e2e

import (
	"bytes"
	"encoding/json"
	"math"
	"net/http"
	"sync"
	"testing"

	"github.com/sawthorn/merch-wallet/internal/info"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type TransferRequest struct {
	ToUser string `json:"toUser"`
	Amount int    `json:"amount"`
}

func TestTransfer_Success(t *testing.T) {
	err := env.resetDB()
	require.NoError(t, err)
	sender := "sender"
	senderToken := makeUserAuthReq(t, env.serverURL, sender, "testpass")
	receiver := "receiver"
	_ = makeUserAuthReq(t, env.serverURL, receiver, "recpass")

	url := env.serverURL + "/api/sendCoin"

	resp := makeJSONRequest(
		t, http.DefaultClient,
		http.MethodPost,
		url,
		TransferRequest{
			ToUser: receiver,
			Amount: 1,
		},
		senderToken,
	)
	resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestTransfer_Circular(t *testing.T) {
	err := env.resetDB()
	require.NoError(t, err)
	balanceByDefault := 1000
	sender := "sender"
	senderToken := makeUserAuthReq(t, env.serverURL, sender, "testpass")
	receiver := "receiver"
	receiverToken := makeUserAuthReq(t, env.serverURL, receiver, "recpass")

	start := make(chan struct{})

	requests := []struct {
		token string
		body  TransferRequest
	}{
		{
			token: senderToken,
			body: TransferRequest{
				ToUser: receiver,
				Amount: 1,
			},
		},
		{
			token: receiverToken,
			body: TransferRequest{
				ToUser: sender,
				Amount: 1,
			},
		},
	}
	var wg sync.WaitGroup
	wg.Add(len(requests))
	for _, r := range requests {
		go func() {
			defer wg.Done()

			<-start

			bodyBytes, err := json.Marshal(r.body)
			if err != nil {
				t.Error(err)
				return
			}
			req, err := http.NewRequest(
				http.MethodPost,
				env.serverURL+"/api/sendCoin",
				bytes.NewReader(bodyBytes))
			if err != nil {
				t.Error(err)
				return
			}

			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+r.token)

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Error(err)
				return
			}
			resp.Body.Close()
			assert.Equal(t, http.StatusOK, resp.StatusCode)
		}()
	}

	close(start)
	wg.Wait()

	for _, r := range requests {
		req, err := http.NewRequest(
			http.MethodGet,
			env.serverURL+"/api/info",
			nil,
		)
		if err != nil {
			t.Error(err)
			return
		}
		req.Header.Set("Authorization", "Bearer "+r.token)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Error(err)
			return
		}
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		infoResp := info.InfoResponse{}
		err = json.NewDecoder(resp.Body).Decode(&infoResp)
		if err != nil {
			t.Error(err)
			return
		}
		assert.Equal(t, balanceByDefault, infoResp.Balance)
	}
}

func TestTransfer_BadAmount(t *testing.T) {
	err := env.resetDB()
	require.NoError(t, err)
	sender := "sender"
	senderToken := makeUserAuthReq(t, env.serverURL, sender, "testpass")
	receiver := "receiver"
	_ = makeUserAuthReq(t, env.serverURL, receiver, "recpass")

	url := env.serverURL + "/api/sendCoin"

	testcases := []struct {
		name   string
		amount int
	}{
		{
			name:   "InsufficientBalance",
			amount: math.MaxInt,
		},
		{
			name:   "NegativeAmount",
			amount: -1,
		},
	}
	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			resp := makeJSONRequest(
				t, http.DefaultClient,
				http.MethodPost,
				url,
				TransferRequest{
					ToUser: receiver,
					Amount: tc.amount,
				},
				senderToken,
			)
			resp.Body.Close()
			assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		})
	}
}

func TestTransfer_ToSelf(t *testing.T) {
	err := env.resetDB()
	require.NoError(t, err)
	sender := "sender"
	senderToken := makeUserAuthReq(t, env.serverURL, sender, "testpass")
	receiver := "receiver"
	_ = makeUserAuthReq(t, env.serverURL, receiver, "recpass")

	url := env.serverURL + "/api/sendCoin"

	resp := makeJSONRequest(
		t, http.DefaultClient,
		http.MethodPost,
		url,
		TransferRequest{
			ToUser: sender,
			Amount: 1,
		},
		senderToken,
	)
	resp.Body.Close()
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func makeUserAuthReq(
	t *testing.T,
	url string,
	username string,
	password string,
) (token string) {
	t.Helper()
	bodyBytes, err := json.Marshal(AuthRequest{
		Username: username,
		Password: password,
	})
	require.NoError(t, err)

	resp, err := http.Post(
		url+"/api/auth",
		"application/json",
		bytes.NewReader(bodyBytes),
	)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	respParsed := AuthResponse{}
	err = json.NewDecoder(resp.Body).Decode(&respParsed)
	require.NoError(t, err)
	return respParsed.Token
}

func makeJSONRequest(
	t *testing.T,
	client *http.Client,
	method string,
	url string,
	body any,
	token string,
) *http.Response {
	t.Helper()

	bodyBytes, err := json.Marshal(body)
	require.NoError(t, err)
	req, err := http.NewRequest(method, url, bytes.NewReader(bodyBytes))
	require.NoError(t, err)

	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := client.Do(req)
	require.NoError(t, err)
	return resp
}
