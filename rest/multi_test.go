package rest

import (
	"bytes"
	"github.com/image357/password"
	"github.com/image357/password/log"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
	"unicode/utf8"
)

func ExampleStartMultiService() {
	// Start rest service on localhost:8080 without any access control.
	err := StartMultiService(":8080", "/prefix", "123", func(string, string, string, string) bool { return true })
	if err != nil {
		// handle error
	}
}

func TestStartMultiPasswordService(t *testing.T) {
	type args struct {
		bindAddress string
		prefix      string
		key         string
		callback    TestAccessFunc
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"start stop", args{":8080", "/prefix/", "123", FullAccessCallback}, false},
		{"error", args{":8080", "/another", "123", FullAccessCallback}, true},
	}
	// init
	password.SetStorePath("tests/workdir/StartMultiService")

	// tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := StartMultiService(tt.args.bindAddress, tt.args.prefix, tt.args.key, tt.args.callback); (err != nil) != tt.wantErr {
				t.Errorf("StartMultiService() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

	// cleanup
	err := os.RemoveAll(password.GetStorePath())
	if err != nil {
		t.Error(err)
	}

	time.Sleep(time.Second)
	err = StopService(1000)
	if err != nil {
		t.Error(err)
	}

	time.Sleep(time.Second)
	err = StopService(1000)
	if err == nil {
		t.Errorf("StopService should have failed")
	}
}

func TestMultiRestCalls(t *testing.T) {
	tests := []struct {
		name   string
		method string
		url    string
		access bool
		json   string
		want   string
		status int
	}{
		// general
		{
			"Wrong method", http.MethodPost, "http://localhost:8080/prefix/overwrite", true,
			`{"accessToken": "abc", "id": "someId", "password": "123"}`, `404 page not found`, http.StatusNotFound,
		},
		{
			"Wrong resource", http.MethodPut, "http://localhost:8080/unknown", true,
			`{"accessToken": "abc", "id": "someId", "password": "123"}`, `404 page not found`, http.StatusNotFound,
		},

		// Overwrite
		{
			"Overwrite success", http.MethodPut, "http://localhost:8080/prefix/overwrite", true,
			`{"accessToken": "abc", "id": "someId", "password": "123"}`, `{}`, http.StatusOK,
		},
		{
			"Overwrite access denied", http.MethodPut, "http://localhost:8080/prefix/overwrite", false,
			`{"accessToken": "abc", "id": "someId", "password": "456"}`, `{}`, http.StatusForbidden,
		},
		{
			"Overwrite bad data", http.MethodPut, "http://localhost:8080/prefix/overwrite", true,
			`{"accessToken": abc, "id": "someId", "password": "789"}`, `{}`, http.StatusBadRequest,
		},
		{
			"Overwrite missing data", http.MethodPut, "http://localhost:8080/prefix/overwrite", true,
			`{"id": "someId", "password": "789"}`, `{}`, http.StatusBadRequest,
		},

		// Get
		{
			"Get success", http.MethodGet, "http://localhost:8080/prefix/get", true,
			`{"accessToken": "abc", "id": "someId"}`, `{"password":"123"}`, http.StatusOK,
		},
		{
			"Get invalid id", http.MethodGet, "http://localhost:8080/prefix/get", true,
			`{"accessToken": "abc", "id": "anotherId"}`, `{}`, http.StatusInternalServerError,
		},
		{
			"Get access denied", http.MethodGet, "http://localhost:8080/prefix/get", false,
			`{"accessToken": "abc", "id": "someId"}`, `{}`, http.StatusForbidden,
		},
		{
			"Get bad data", http.MethodGet, "http://localhost:8080/prefix/get", true,
			`{"accessToken": abc, "id": "someId"}`, `{}`, http.StatusBadRequest,
		},
		{
			"Get missing data", http.MethodGet, "http://localhost:8080/prefix/get", true,
			`{"id": "someId"}`, `{}`, http.StatusBadRequest,
		},

		// Check
		{
			"Check success", http.MethodGet, "http://localhost:8080/prefix/check", true,
			`{"accessToken": "abc", "id": "someId", "password": "123"}`, `{"result":true}`, http.StatusOK,
		},
		{
			"Check invalid password", http.MethodGet, "http://localhost:8080/prefix/check", true,
			`{"accessToken": "abc", "id": "someId", "password": "456"}`, `{"result":false}`, http.StatusOK,
		},
		{
			"Check invalid id", http.MethodGet, "http://localhost:8080/prefix/check", true,
			`{"accessToken": "abc", "id": "anotherId", "password": "456"}`, `{}`, http.StatusInternalServerError,
		},
		{
			"Check access denied", http.MethodGet, "http://localhost:8080/prefix/check", false,
			`{"accessToken": "abc", "id": "someId", "password": "123"}`, `{}`, http.StatusForbidden,
		},
		{
			"Check bad data", http.MethodGet, "http://localhost:8080/prefix/check", true,
			`{"accessToken": abc, "id": "someId", "password": "123"}`, `{}`, http.StatusBadRequest,
		},
		{
			"Check missing data", http.MethodGet, "http://localhost:8080/prefix/check", true,
			`{"id": "someId", "password": "123"}`, `{}`, http.StatusBadRequest,
		},

		// Set
		{
			"Set success", http.MethodPut, "http://localhost:8080/prefix/set", true,
			`{"accessToken": "abc", "id": "someId", "oldPassword": "123", "newPassword": "456"}`, `{}`, http.StatusOK,
		},
		{
			"Set invalid password", http.MethodPut, "http://localhost:8080/prefix/set", true,
			`{"accessToken": "abc", "id": "someId", "oldPassword": "123", "newPassword": "456"}`, `{}`, http.StatusInternalServerError,
		},
		{
			"Set access denied", http.MethodPut, "http://localhost:8080/prefix/set", false,
			`{"accessToken": "abc", "id": "someId", "oldPassword": "123", "newPassword": "456"}`, `{}`, http.StatusForbidden,
		},
		{
			"Set bad data", http.MethodPut, "http://localhost:8080/prefix/set", true,
			`{"accessToken": abc, "id": "someId", "oldPassword": "456", "newPassword": "789"}`, `{}`, http.StatusBadRequest,
		},
		{
			"Set missing data", http.MethodPut, "http://localhost:8080/prefix/set", true,
			`{"id": "someId", "oldPassword": "456", "newPassword": "789"}`, `{}`, http.StatusBadRequest,
		},

		// Unset
		{
			"Unset success", http.MethodDelete, "http://localhost:8080/prefix/unset", true,
			`{"accessToken": "abc", "id": "someId", "password": "456"}`, `{}`, http.StatusOK,
		},
		{
			"Unset invalid id", http.MethodDelete, "http://localhost:8080/prefix/unset", true,
			`{"accessToken": "abc", "id": "someId", "password": "456"}`, `{}`, http.StatusInternalServerError,
		},
		{
			"Overwrite create", http.MethodPut, "http://localhost:8080/prefix/overwrite", true,
			`{"accessToken": "abc", "id": "someId", "password": "123"}`, `{}`, http.StatusOK,
		},
		{
			"Unset invalid password", http.MethodDelete, "http://localhost:8080/prefix/unset", true,
			`{"accessToken": "abc", "id": "someId", "password": "456"}`, `{}`, http.StatusInternalServerError,
		},
		{
			"Unset access denied", http.MethodDelete, "http://localhost:8080/prefix/unset", false,
			`{"accessToken": "abc", "id": "someId", "password": "wrong"}`, `{}`, http.StatusForbidden,
		},
		{
			"Unset bad data", http.MethodDelete, "http://localhost:8080/prefix/unset", true,
			`{"accessToken": abc, "id": "someId", "password": "456"}`, `{}`, http.StatusBadRequest,
		},
		{
			"Unset missing data", http.MethodDelete, "http://localhost:8080/prefix/unset", true,
			`{"id": "someId", "password": "456"}`, `{}`, http.StatusBadRequest,
		},

		// List
		{
			"Overwrite create a", http.MethodPut, "http://localhost:8080/prefix/overwrite", true,
			`{"accessToken": "abc", "id": "a", "password": "123"}`, `{}`, http.StatusOK,
		},
		{
			"Overwrite create b/foo", http.MethodPut, "http://localhost:8080/prefix/overwrite", true,
			`{"accessToken": "abc", "id": "b/foo", "password": "123"}`, `{}`, http.StatusOK,
		},
		{
			"Overwrite create c/bar", http.MethodPut, "http://localhost:8080/prefix/overwrite", true,
			`{"accessToken": "abc", "id": "c\\bar", "password": "123"}`, `{}`, http.StatusOK,
		},
		{
			"List success", http.MethodGet, "http://localhost:8080/prefix/list", true,
			`{"accessToken": "abc"}`, `{"ids":["a","b/foo","c/bar","someid"]}`, http.StatusOK,
		},
		{
			"List access denied", http.MethodGet, "http://localhost:8080/prefix/list", false,
			`{"accessToken": "abc"}`, `{}`, http.StatusForbidden,
		},
		{
			"List bad data", http.MethodGet, "http://localhost:8080/prefix/list", true,
			`{"accessToken": abc}`, `{}`, http.StatusBadRequest,
		},
		{
			"List missing data", http.MethodGet, "http://localhost:8080/prefix/list", true,
			`{}`, `{}`, http.StatusBadRequest,
		},

		// Delete
		{
			"Delete success", http.MethodDelete, "http://localhost:8080/prefix/delete", true,
			`{"accessToken": "abc", "id": "someId"}`, `{}`, http.StatusOK,
		},
		{
			"Delete invalid id", http.MethodDelete, "http://localhost:8080/prefix/delete", true,
			`{"accessToken": "abc", "id": "someId"}`, `{}`, http.StatusInternalServerError,
		},
		{
			"Delete access denied", http.MethodDelete, "http://localhost:8080/prefix/delete", false,
			`{"accessToken": "abc", "id": "someId"}`, `{}`, http.StatusForbidden,
		},
		{
			"Delete bad data", http.MethodDelete, "http://localhost:8080/prefix/delete", true,
			`{"accessToken": abc, "id": "someId"}`, `{}`, http.StatusBadRequest,
		},
		{
			"Delete missing data", http.MethodDelete, "http://localhost:8080/prefix/delete", true,
			`{"id": "someId"}`, `{}`, http.StatusBadRequest,
		},

		// Clean
		{
			"List before Clean", http.MethodGet, "http://localhost:8080/prefix/list", true,
			`{"accessToken": "abc"}`, `{"ids":["a","b/foo","c/bar"]}`, http.StatusOK,
		},
		{
			"Clean success", http.MethodDelete, "http://localhost:8080/prefix/clean", true,
			`{"accessToken": "abc"}`, `{}`, http.StatusOK,
		},
		{
			"List after Clean", http.MethodGet, "http://localhost:8080/prefix/list", true,
			`{"accessToken": "abc"}`, `{"ids":[]}`, http.StatusOK,
		},
		{
			"Clean access denied", http.MethodDelete, "http://localhost:8080/prefix/clean", false,
			`{"accessToken": "abc"}`, `{}`, http.StatusForbidden,
		},
		{
			"Clean bad data", http.MethodDelete, "http://localhost:8080/prefix/clean", true,
			`{"accessToken": abc}`, `{}`, http.StatusBadRequest,
		},
		{
			"Clean missing data", http.MethodDelete, "http://localhost:8080/prefix/clean", true,
			`{}`, `{}`, http.StatusBadRequest,
		},
	}
	// init
	oldLevel := log.Level(slog.LevelDebug)
	password.SetStorePath("tests/workdir/MultiRestCalls")
	err := StartMultiService(":8080", "/prefix", "123", DebugAccessCallback)
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(time.Second)

	// tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			debugAccessSuccess = tt.access

			req, err := http.NewRequest(tt.method, tt.url, bytes.NewBuffer([]byte(tt.json)))
			if err != nil {
				t.Error(err)
			}
			req.Header.Set("Content-Type", "application/json")

			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				t.Error(err)
			}

			if resp.StatusCode != tt.status {
				t.Errorf("StatusCode error = %v, wantErr %v", resp.StatusCode, tt.status)
			}

			b, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Error(err)
			}
			if !utf8.Valid(b) {
				t.Errorf("invalid utf8 character in response body")
			}

			result := string(b)
			if result != tt.want {
				if !(password.HashPassword && strings.HasSuffix(tt.url, "/get")) {
					t.Errorf("result = %v, want %v", result, tt.want)
				}
			}

			err = resp.Body.Close()
			if err != nil {
				t.Error(err)
			}
		})
	}

	// cleanup
	err = os.RemoveAll(password.GetStorePath())
	if err != nil {
		t.Error(err)
	}

	time.Sleep(time.Second)
	err = StopService(1000)
	if err != nil {
		t.Error(err)
	}

	log.Level(oldLevel)
}
