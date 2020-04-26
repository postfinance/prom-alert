package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPost(t *testing.T) {
	var reqBody []byte

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		d, err := ioutil.ReadAll(req.Body)
		require.NoError(t, err)
		reqBody = d
		rw.Write([]byte(`OK`))
	}))

	defer server.Close()

	c := client{
		url: server.URL,
		Client: &http.Client{
			Timeout: timeout,
		},
	}

	expectedAlert := alert{
		Status: statusFiring,
		Labels: labels{
			"team": "linux",
		},
		Annotations: annotations{
			Summary: "test",
		},
	}

	err := c.post(expectedAlert)
	require.NoError(t, err)

	alerts := []alert{}
	err = json.Unmarshal(reqBody, &alerts)
	require.NoError(t, err)
	require.Len(t, alerts, 1)
	assert.Equal(t, expectedAlert, alerts[0])
}

func TestLabelsSet(t *testing.T) {
	tests := []struct {
		labelsStr      string
		expectedLabels labels
		wantErr        bool
	}{
		{
			labelsStr:      "",
			expectedLabels: labels{},
			wantErr:        false,
		},
		{
			labelsStr: "team=linux",
			expectedLabels: labels{
				"team": "linux",
			},
			wantErr: false,
		},
		{
			labelsStr: "team=linux,severity=critical",
			expectedLabels: labels{
				"team":     "linux",
				"severity": "critical",
			},
			wantErr: false,
		},
		{
			labelsStr:      "team=linux,severity",
			expectedLabels: labels{},
			wantErr:        true,
		},
	}

	// nolint: scopelint
	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			l := labels{}
			err := l.Set(tt.labelsStr)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.expectedLabels, l)
		})
	}
}
