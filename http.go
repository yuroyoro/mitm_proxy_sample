package main

import (
	"io"

	"net/http"
)

func (proxy *MiTMProxy) transportHTTPRequest(w http.ResponseWriter, r *http.Request) {
	dumpRequest(r)

	// remove proxy headers(i.e Connection...)
	removeProxyHeaders(r)

	// transport request to target host
	resp, err := proxy.transport.RoundTrip(r)
	if err != nil {
		proxy.error("error read response %v %v", r.URL.Host, err.Error())
		if resp == nil {
			http.Error(w, err.Error(), 500)
			return
		}
	}

	dumpResponse(resp)

	// copy response to client
	proxy.writeResponse(w, resp)
}

func removeProxyHeaders(r *http.Request) {
	r.RequestURI = ""
	r.Header.Del("Accept-Encoding")
	r.Header.Del("Proxy-Connection")
	r.Header.Del("Proxy-Authenticate")
	r.Header.Del("Proxy-Authorization")
	r.Header.Del("Connection")
}

func (proxy *MiTMProxy) writeResponse(w http.ResponseWriter, resp *http.Response) {
	// copy headers
	dest := w.Header()
	for k, vs := range resp.Header {
		for _, v := range vs {
			dest.Add(k, v)
		}
	}
	w.WriteHeader(resp.StatusCode)

	_, err := io.Copy(w, resp.Body)
	if err != nil {
		proxy.warn("Can't read response body %v", err)
	}

	if err := resp.Body.Close(); err != nil {
		proxy.warn("Can't close response body %v", err)
	}
}
