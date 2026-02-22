package authserver

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestAuthServer_FormPage(t *testing.T) {
	resultChan := make(chan Credentials, 1)
	handler := NewHandler(resultChan)

	req := httptest.NewRequest("GET", "/auth", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("GET /auth status = %d, want 200", w.Code)
	}

	body := w.Body.String()
	if !strings.Contains(body, "client_id") {
		t.Errorf("form page missing client_id field")
	}
	if !strings.Contains(body, "api_key") {
		t.Errorf("form page missing api_key field")
	}
}

func TestAuthServer_Submit(t *testing.T) {
	resultChan := make(chan Credentials, 1)
	handler := NewHandler(resultChan)

	form := url.Values{}
	form.Set("client_id", "test_client_id")
	form.Set("api_key", "test_api_key")

	req := httptest.NewRequest("POST", "/auth", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("POST /auth status = %d, want 200", w.Code)
	}

	// Check that credentials were sent to channel
	select {
	case creds := <-resultChan:
		if creds.ClientID != "test_client_id" {
			t.Errorf("ClientID = %v, want test_client_id", creds.ClientID)
		}
		if creds.APIKey != "test_api_key" {
			t.Errorf("APIKey = %v, want test_api_key", creds.APIKey)
		}
	default:
		t.Error("credentials not received on channel")
	}
}

func TestAuthServer_SubmitValidation(t *testing.T) {
	resultChan := make(chan Credentials, 1)
	handler := NewHandler(resultChan)

	// Empty form
	form := url.Values{}
	req := httptest.NewRequest("POST", "/auth", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("POST /auth with empty form status = %d, want 400", w.Code)
	}
}

func TestAuthServer_NotFound(t *testing.T) {
	resultChan := make(chan Credentials, 1)
	handler := NewHandler(resultChan)

	tests := []struct {
		name string
		path string
	}{
		{"root path", "/"},
		{"other path", "/other"},
		{"nested path", "/auth/extra"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.path, nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			if w.Code != http.StatusNotFound {
				t.Errorf("GET %s status = %d, want 404", tt.path, w.Code)
			}
		})
	}
}

func TestAuthServer_MethodNotAllowed(t *testing.T) {
	resultChan := make(chan Credentials, 1)
	handler := NewHandler(resultChan)

	methods := []string{
		http.MethodPut,
		http.MethodDelete,
		http.MethodPatch,
		http.MethodHead,
		http.MethodOptions,
	}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/auth", nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			if w.Code != http.StatusMethodNotAllowed {
				t.Errorf("%s /auth status = %d, want 405", method, w.Code)
			}
		})
	}
}

func TestAuthServer_PartialCredentials(t *testing.T) {
	tests := []struct {
		name         string
		clientID     string
		apiKey       string
		wantStatus   int
		wantClientID string
		wantAPIKey   string
	}{
		{
			name:         "only client_id",
			clientID:     "test_client",
			apiKey:       "",
			wantStatus:   http.StatusOK,
			wantClientID: "test_client",
			wantAPIKey:   "",
		},
		{
			name:         "only api_key",
			clientID:     "",
			apiKey:       "test_api",
			wantStatus:   http.StatusOK,
			wantClientID: "",
			wantAPIKey:   "test_api",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resultChan := make(chan Credentials, 1)
			handler := NewHandler(resultChan)

			form := url.Values{}
			if tt.clientID != "" {
				form.Set("client_id", tt.clientID)
			}
			if tt.apiKey != "" {
				form.Set("api_key", tt.apiKey)
			}

			req := httptest.NewRequest("POST", "/auth", strings.NewReader(form.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("POST /auth status = %d, want %d", w.Code, tt.wantStatus)
			}

			// Verify credentials were sent to channel with partial values
			select {
			case creds := <-resultChan:
				if creds.ClientID != tt.wantClientID {
					t.Errorf("ClientID = %v, want %v", creds.ClientID, tt.wantClientID)
				}
				if creds.APIKey != tt.wantAPIKey {
					t.Errorf("APIKey = %v, want %v", creds.APIKey, tt.wantAPIKey)
				}
			default:
				t.Error("credentials should be sent for partial input")
			}
		})
	}
}

func TestAuthServer_SuccessPageContent(t *testing.T) {
	resultChan := make(chan Credentials, 1)
	handler := NewHandler(resultChan)

	form := url.Values{}
	form.Set("client_id", "test_client_id")
	form.Set("api_key", "test_api_key")

	req := httptest.NewRequest("POST", "/auth", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("POST /auth status = %d, want 200", w.Code)
	}

	body := w.Body.String()
	expectedTexts := []string{
		"You're Connected",
		"close this window",
		"Return to your terminal",
	}

	for _, text := range expectedTexts {
		if !strings.Contains(body, text) {
			t.Errorf("success page missing expected text: %q", text)
		}
	}

	// Verify content type
	contentType := w.Header().Get("Content-Type")
	if contentType != "text/html" {
		t.Errorf("Content-Type = %q, want text/html", contentType)
	}
}

func TestServer_URL(t *testing.T) {
	server, err := NewServer()
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}
	defer server.Shutdown()

	url := server.URL()

	// Check URL format
	if !strings.HasPrefix(url, "http://127.0.0.1:") {
		t.Errorf("URL = %q, want prefix http://127.0.0.1:", url)
	}

	if !strings.HasSuffix(url, "/auth") {
		t.Errorf("URL = %q, want suffix /auth", url)
	}

	// Verify URL contains a valid port
	if !strings.Contains(url, "127.0.0.1:") {
		t.Errorf("URL = %q, should contain port", url)
	}
}

func TestServer_NewServer(t *testing.T) {
	server, err := NewServer()
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}
	defer server.Shutdown()

	// Verify server components are initialized
	if server.listener == nil {
		t.Error("server.listener is nil")
	}

	if server.server == nil {
		t.Error("server.server is nil")
	}

	if server.resultChan == nil {
		t.Error("server.resultChan is nil")
	}

	// Verify listener is on localhost
	addr := server.listener.Addr().String()
	if !strings.HasPrefix(addr, "127.0.0.1:") {
		t.Errorf("listener address = %q, want prefix 127.0.0.1:", addr)
	}
}

func TestServer_StartAndShutdown(t *testing.T) {
	server, err := NewServer()
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}

	// Start the server
	server.Start()

	// Give server time to start
	time.Sleep(50 * time.Millisecond)

	// Verify server is listening by making a request
	resp, err := http.Get(server.URL())
	if err != nil {
		t.Fatalf("GET %s error = %v", server.URL(), err)
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("GET %s status = %d, want 200", server.URL(), resp.StatusCode)
	}

	// Shutdown the server
	err = server.Shutdown()
	if err != nil {
		t.Errorf("Shutdown() error = %v", err)
	}

	// Verify server is no longer accessible
	time.Sleep(50 * time.Millisecond)
	_, err = http.Get(server.URL())
	if err == nil {
		t.Error("expected error after shutdown, got nil")
	}
}

func TestServer_WaitForCredentials_Timeout(t *testing.T) {
	server, err := NewServer()
	if err != nil {
		t.Fatalf("NewServer() error = %v", err)
	}
	defer server.Shutdown()

	// Wait with a short timeout and no credentials submitted
	timeout := 100 * time.Millisecond
	start := time.Now()

	creds, err := server.WaitForCredentials(timeout)

	elapsed := time.Since(start)

	// Verify timeout occurred
	if err == nil {
		t.Error("WaitForCredentials() expected timeout error, got nil")
	}

	if creds != nil {
		t.Errorf("WaitForCredentials() returned credentials = %v, want nil", creds)
	}

	// Verify error message
	expectedMsg := "timeout waiting for credentials"
	if err != nil && !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("error = %q, want to contain %q", err.Error(), expectedMsg)
	}

	// Verify timeout duration is reasonable (allow some margin)
	if elapsed < timeout || elapsed > timeout+50*time.Millisecond {
		t.Errorf("elapsed time = %v, want approximately %v", elapsed, timeout)
	}
}
