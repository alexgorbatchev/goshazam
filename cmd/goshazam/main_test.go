package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strings"
	"testing"

	"github.com/alexgorbatchev/goshazam"
	"github.com/alexgorbatchev/goshazam/pkg/client"
)

func TestCLICommands(t *testing.T) {
	cmd := newRootCommand()
	if cmd == nil {
		t.Fatalf("expected non-nil root command")
	}

	// Verify command registration
	commands := make(map[string]bool)
	for _, c := range cmd.Commands() {
		commands[c.Name()] = true
	}

	for _, expected := range []string{"recognize", "signature", "related", "upgrade"} {
		if !commands[expected] {
			t.Errorf("expected command %q to be registered", expected)
		}
	}

	// Verify version template output
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"--version"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("executing --version failed: %v", err)
	}

	out := buf.String()
	if out != version+"\n" {
		t.Errorf("expected version %q, got %q", version+"\n", out)
	}
}

func TestCLISignatureAndRecognize(t *testing.T) {
	oggPath := filepath.Join("..", "..", "ShazamIO", "examples", "data", "Gloria.ogg")

	// Test signature command
	var sigBuf bytes.Buffer
	cmdSig := newRootCommand()
	cmdSig.SetOut(&sigBuf)
	cmdSig.SetArgs([]string{"signature", oggPath})
	if err := cmdSig.Execute(); err != nil {
		t.Fatalf("signature command failed: %v", err)
	}
	if !strings.HasPrefix(sigBuf.String(), "data:audio/vnd.shazam.sig;base64,") {
		t.Errorf("expected signature output prefix, got %s", sigBuf.String()[:30])
	}

	// Test signature with --json
	var sigJSONBuf bytes.Buffer
	cmdSigJSON := newRootCommand()
	cmdSigJSON.SetOut(&sigJSONBuf)
	cmdSigJSON.SetArgs([]string{"signature", "--json", oggPath})
	if err := cmdSigJSON.Execute(); err != nil {
		t.Fatalf("signature --json failed: %v", err)
	}
	if !strings.Contains(sigJSONBuf.String(), `"uri":`) {
		t.Errorf("expected JSON with uri field, got %s", sigJSONBuf.String())
	}

	// Test recognize with non-existent file
	var recErrBuf bytes.Buffer
	cmdRecErr := newRootCommand()
	cmdRecErr.SetOut(&recErrBuf)
	cmdRecErr.SetArgs([]string{"recognize", "non_existent.mp3"})
	if err := cmdRecErr.Execute(); err == nil {
		t.Errorf("expected error on non-existent file")
	}

	// Test signature with non-existent file
	var sigErrBuf bytes.Buffer
	cmdSigErr := newRootCommand()
	cmdSigErr.SetOut(&sigErrBuf)
	cmdSigErr.SetArgs([]string{"signature", "non_existent.mp3"})
	if err := cmdSigErr.Execute(); err == nil {
		t.Errorf("expected error on non-existent signature file")
	}

	// Test related with invalid ID
	cmdRel := newRootCommand()
	cmdRel.SetArgs([]string{"related", "not_a_number"})
	if err := cmdRel.Execute(); err == nil {
		t.Errorf("expected error for non-numeric track ID")
	}

	// Test upgrade command
	var upgBuf bytes.Buffer
	cmdUpg := newRootCommand()
	cmdUpg.SetOut(&upgBuf)
	cmdUpg.SetArgs([]string{"upgrade"})
	if err := cmdUpg.Execute(); err == nil {
		if !strings.Contains(upgBuf.String(), "already up to date") && !strings.Contains(upgBuf.String(), "Checking for newer") {
			t.Errorf("expected upgrade output, got %s", upgBuf.String())
		}
	}
}

func TestCLIRecognizeMock(t *testing.T) {
	mockResponse := `{
		"matches": [{"id": "1", "offset": 10.0}],
		"track": {
			"key": "53982678",
			"title": "I Will Survive",
			"subtitle": "Gloria Gaynor",
			"images": {"coverart": "https://example.com/cover.jpg"},
			"hub": {
				"options": [{"actions": [{"uri": "https://music.apple.com/song/1"}]}],
				"providers": [{"actions": [{"uri": "https://spotify.com/track/1"}]}]
			},
			"sections": [{"type": "VIDEO", "youtubeurl": "https://youtube.com/1"}]
		}
	}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(mockResponse))
	}))
	defer server.Close()

	origFactory := clientFactory
	clientFactory = func(language, country, proxy string) *goshazam.Shazam {
		c := client.NewHTTPClient(client.WithHTTPClient(server.Client()))
		return goshazam.New(goshazam.WithHTTPClient(c), goshazam.WithDiscoveryURL(server.URL))
	}
	defer func() { clientFactory = origFactory }()

	oggPath := filepath.Join("..", "..", "ShazamIO", "examples", "data", "Gloria.ogg")

	// Test recognize text output
	var recBuf bytes.Buffer
	cmd := newRootCommand()
	cmd.SetOut(&recBuf)
	cmd.SetArgs([]string{"recognize", oggPath})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("recognize failed: %v", err)
	}
	out := recBuf.String()
	if !strings.Contains(out, "I Will Survive") || !strings.Contains(out, "Gloria Gaynor") {
		t.Errorf("expected track info in output, got %s", out)
	}

	// Test recognize --json output
	var jsonBuf bytes.Buffer
	cmdJSON := newRootCommand()
	cmdJSON.SetOut(&jsonBuf)
	cmdJSON.SetArgs([]string{"recognize", "--json", oggPath})
	if err := cmdJSON.Execute(); err != nil {
		t.Fatalf("recognize --json failed: %v", err)
	}
	if !strings.Contains(jsonBuf.String(), `"53982678"`) {
		t.Errorf("expected key in JSON output, got %s", jsonBuf.String())
	}

	// Test recognize with no matches
	serverEmpty := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"matches": []}`))
	}))
	defer serverEmpty.Close()

	clientFactory = func(language, country, proxy string) *goshazam.Shazam {
		c := client.NewHTTPClient(client.WithHTTPClient(serverEmpty.Client()))
		return goshazam.New(goshazam.WithHTTPClient(c), goshazam.WithDiscoveryURL(serverEmpty.URL))
	}

	var emptyRecBuf bytes.Buffer
	cmdEmpty := newRootCommand()
	cmdEmpty.SetOut(&emptyRecBuf)
	cmdEmpty.SetArgs([]string{"recognize", oggPath})
	if err := cmdEmpty.Execute(); err != nil {
		t.Fatalf("recognize on empty matches failed: %v", err)
	}
	if !strings.Contains(emptyRecBuf.String(), "No matches found.") {
		t.Errorf("expected 'No matches found.', got %s", emptyRecBuf.String())
	}
}

type cliRewriteTransport struct {
	target    *url.URL
	transport http.RoundTripper
}

func (t *cliRewriteTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.URL.Scheme = t.target.Scheme
	req.URL.Host = t.target.Host
	return t.transport.RoundTrip(req)
}

func TestCLIRelatedMock(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if strings.Contains(r.URL.Path, "999") {
			_, _ = w.Write([]byte(`{"tracks": []}`))
		} else {
			_, _ = w.Write([]byte(`{"tracks": [{"key": "123", "title": "Track", "subtitle": "Artist"}]}`))
		}
	}))
	defer server.Close()

	serverURL, _ := url.Parse(server.URL)
	origFactory := clientFactory
	clientFactory = func(language, country, proxy string) *goshazam.Shazam {
		httpClient := &http.Client{
			Transport: &cliRewriteTransport{
				target:    serverURL,
				transport: server.Client().Transport,
			},
		}
		c := client.NewHTTPClient(client.WithHTTPClient(httpClient))
		return goshazam.New(goshazam.WithHTTPClient(c))
	}
	defer func() { clientFactory = origFactory }()

	// Test related --json
	var buf bytes.Buffer
	cmd := newRootCommand()
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{"related", "--json", "123"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("related --json failed: %v", err)
	}

	// Test related text format
	var bufText bytes.Buffer
	cmdText := newRootCommand()
	cmdText.SetOut(&bufText)
	cmdText.SetArgs([]string{"related", "123"})
	if err := cmdText.Execute(); err != nil {
		t.Fatalf("related text format failed: %v", err)
	}

	// Test related with no tracks
	var emptyRelBuf bytes.Buffer
	cmdEmptyRel := newRootCommand()
	cmdEmptyRel.SetOut(&emptyRelBuf)
	cmdEmptyRel.SetArgs([]string{"related", "999"})
	if err := cmdEmptyRel.Execute(); err != nil {
		t.Fatalf("related on empty failed: %v", err)
	}
	if !strings.Contains(emptyRelBuf.String(), "No related tracks found.") {
		t.Errorf("expected 'No related tracks found.', got %s", emptyRelBuf.String())
	}
}
