package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/fastly/compute-sdk-go/fsthttp"
)

// Temporarily here but all this should come from an sdk-go import

type StanzaGuard struct {
	// required
	Backend     string
	HubURL      string
	Name        string
	Environment string

	// optional
	Feature       string
	PriorityBoost int32
	DefaultWeight float32
}

type DynamicBackends struct {
	Name string
	URL  string
	SSL  bool
}

type FastlyRequest struct {
	Backend string
	URL     string
	Method  string
	Body    io.Reader
}

type GetTokenResponse struct {
	Granted bool   `json:"granted"`
	Token   string `json:"token"`
	Reason  string `json:"reason"`
}

func Guard(sg StanzaGuard, fr FastlyRequest) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		key := r.Header.Get("X-Stanza-Key")
		if key == "" {
			w.WriteHeader(fsthttp.StatusProxyAuthRequired)
			w.Write([]byte("407 Proxy Authentication Required"))
			return
		}

		// Make a Stanza Quota Request
		body := []byte(fmt.Sprintf("{\"selector\":{\"environment\":\"%s\",\"guardName\":\"%s\",\"featureName\":\"%s\"}}", sg.Environment, sg.Name, sg.Feature))
		req, err := fsthttp.NewRequest("POST", fmt.Sprintf("%s/v1/quota/token", sg.HubURL), bytes.NewBuffer(body))
		if err != nil {
			w.WriteHeader(fsthttp.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		req.Header.Set("X-Stanza-Key", key)
		resp, err := req.Send(r.Context(), sg.Backend)
		if err != nil {
			w.WriteHeader(fsthttp.StatusBadGateway)
			w.Write([]byte(err.Error()))
			return
		}
		if resp.StatusCode != http.StatusOK {
			w.WriteHeader(fsthttp.StatusFailedDependency)
			io.Copy(w, resp.Body)
			return
		}

		// Decode JSON response
		var getTokenResp GetTokenResponse
		err = json.NewDecoder(resp.Body).Decode(&getTokenResp)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
		}
		if !getTokenResp.Granted {
			w.WriteHeader(fsthttp.StatusTooManyRequests)
			w.Write([]byte("429 Too Many Requests"))
			return
		}

		// Success, proxy request
		r.Header.Set("X-Stanza-Token", getTokenResp.Token)
		fastlyRequest(fr, w, r)
		return
	}
}

func fastlyRequest(fr FastlyRequest, w http.ResponseWriter, r *http.Request) {
	req, err := fsthttp.NewRequest(fr.Method, fr.URL, fr.Body)
	if err != nil {
		w.WriteHeader(fsthttp.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	resp, err := req.Send(r.Context(), fr.Backend)
	if err != nil {
		w.WriteHeader(fsthttp.StatusBadGateway)
		w.Write([]byte(err.Error()))
		return
	}
	for k, v := range resp.Header {
		w.Header()[k] = v
	}
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

func RegisterDynamicBackends(backends []DynamicBackends) {
	for _, backend := range backends {
		opts := fsthttp.NewBackendOptions()
		opts.HostOverride(backend.URL)
		opts.ConnectTimeout(time.Duration(1) * time.Second)
		opts.FirstByteTimeout(time.Duration(15) * time.Second)
		opts.BetweenBytesTimeout(time.Duration(10) * time.Second)
		opts.UseSSL(backend.SSL)
		fsthttp.RegisterDynamicBackend(backend.Name, backend.URL, opts)
	}
}
