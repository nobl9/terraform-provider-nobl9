package frameworkprovider

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/nobl9/nobl9-go/sdk"
	replayV1 "github.com/nobl9/nobl9-go/sdk/endpoints/replay/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSDKClientReplay(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name    string
		status  int
		body    string
		wantErr string
	}{
		{name: "success", status: http.StatusOK},
		{
			name: "known availability reason", status: http.StatusConflict,
			body:    " integration_does_not_support_replay\n",
			wantErr: "The Data Source does not support Replay yet",
		},
		{
			name: "unknown availability reason", status: http.StatusBadRequest,
			body:    "additional restriction",
			wantErr: "bad response (status: 400): additional restriction",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodPost, r.Method)
				assert.Equal(t, "/api/timetravel", r.URL.Path)
				assert.Equal(t, "example", r.Header.Get(sdk.HeaderProject))
				body, err := io.ReadAll(r.Body)
				if assert.NoError(t, err) {
					assert.JSONEq(t, `{
						"project":"example",
						"slo":"application-score",
						"duration":{"unit":"Minute","value":60}
					}`, string(body))
				}
				w.WriteHeader(tc.status)
				_, err = io.WriteString(w, tc.body)
				assert.NoError(t, err)
			}))
			t.Cleanup(server.Close)
			apiURL, err := url.Parse(server.URL + "/api")
			require.NoError(t, err)
			client, err := sdk.NewClient(&sdk.Config{
				URL: apiURL, DisableOkta: true, Organization: "example", Timeout: time.Second,
			})
			require.NoError(t, err)
			err = (sdkClient{client: client}).Replay(t.Context(), replayV1.RunRequest{
				Project: "example", SLO: "application-score",
				Duration: replayV1.Duration{Unit: replayV1.DurationUnitMinute, Value: 60},
			})
			if tc.wantErr != "" {
				require.EqualError(t, err, tc.wantErr)
				return
			}
			require.NoError(t, err)
		})
	}
}
