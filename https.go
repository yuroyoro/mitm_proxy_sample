package main

import (
	"bufio"
	"io"
	"net"
	"net/http"

	"crypto/tls"
)

func (proxy *MiTMProxy) relayHTTPSRequest(w http.ResponseWriter, r *http.Request) {
	proxy.info("relayHTTPSRequest : %s %s", r.Method, r.URL.String())

	dest, err := net.Dial("tcp", r.Host)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	conn := hijackConnect(w)
	conn.Write([]byte("HTTP/1.0 200 OK\r\n\r\n"))

	proxy.info("relayHTTPSRequest : start relaying tcp packets %s %s", r.Method, r.URL.String())
	go transfer(dest, conn)
	go transfer(conn, dest)
}

func transfer(dest io.WriteCloser, source io.ReadCloser) {
	defer dest.Close()
	defer source.Close()
	io.Copy(dest, source)
}

func (proxy *MiTMProxy) mitmRequest(w http.ResponseWriter, r *http.Request) {
	conn := hijackConnect(w)
	conn.Write([]byte("HTTP/1.0 200 OK\r\n\r\n"))

	// launch goroutine to transporting request with mitm sniffing
	go proxy.transportHTTPSRequest(w, r, conn)
}

func hijackConnect(w http.ResponseWriter) net.Conn {
	hj, ok := w.(http.Hijacker)
	if !ok {
		panic("httpserver does not support hijacking")
	}

	conn, _, err := hj.Hijack()
	if err != nil {
		panic("Cannot hijack connection " + err.Error())
	}

	return conn
}

func (proxy *MiTMProxy) transportHTTPSRequest(w http.ResponseWriter, r *http.Request, conn net.Conn) {
	proxy.info("transportHTTPSRequest : %s %s", r.Method, r.URL.String())

	host := r.Host
	tlsConfig, err := proxy.generateTLSConfig(host)
	if err != nil {
		if _, err := conn.Write([]byte("HTTP/1.0 500 Internal Server Error\r\n\r\n")); err != nil {
			proxy.error("Failed to write response : %v", err)
		}
		conn.Close()
	}

	tlsConn := tls.Server(conn, tlsConfig)
	if err := tlsConn.Handshake(); err != nil {
		proxy.error("Cannot handshake client %v %v", r.Host, err)
		return
	}
	defer tlsConn.Close()

	proxy.info("transportHTTPSRequest : established tls connection")

	tlsIn := bufio.NewReader(tlsConn)
	for !isEOF(tlsIn) {
		req, err := http.ReadRequest(tlsIn)
		if err != nil {
			if err == io.EOF {
				proxy.error("EOF detected when read request from client: %v %v", r.Host, err)
			} else {
				proxy.error("Cannot read request from client: %v %v", r.Host, err)
			}
			return
		}

		proxy.info("transportHTTPSRequest : read request : %s %s", req.Method, req.URL.String())

		req.URL.Scheme = "https"
		req.URL.Host = r.Host
		req.RequestURI = req.URL.String()
		req.RemoteAddr = r.RemoteAddr

		dumpRequest(req)
		removeProxyHeaders(req)

		// transport request to target host
		resp, err := proxy.transport.RoundTrip(req)
		if err != nil {
			proxy.error("error read response %v %v", r.URL.Host, err.Error())
			if resp == nil {
				http.Error(w, err.Error(), 500)
				return
			}
		}

		proxy.info("transportHTTPSRequest : transport request: %s", resp.Status)

		dumpResponse(resp)

		// copy response to client
		resp.Write(tlsConn)
	}

	proxy.info("transportHTTPSRequest : finished ")
}

func isEOF(r *bufio.Reader) bool {
	_, err := r.Peek(1)
	if err == io.EOF {
		return true
	}
	return false
}

func (proxy *MiTMProxy) generateTLSConfig(host string) (*tls.Config, error) {
	config := tls.Config{InsecureSkipVerify: true}

	host, _ = proxy.splitHostPort(host)
	proxy.warn("generate tls config for : %s", host)
	cert, err := proxy.findOrCreateCert(host)
	if err != nil {
		proxy.warn("failed to find cert : %s : %v", host, err)
		return nil, err
	}

	config.Certificates = append(config.Certificates, *cert)
	return &config, nil
}

func (proxy *MiTMProxy) splitHostPort(s string) (string, string) {
	host, port, err := net.SplitHostPort(s)
	if err != nil {
		proxy.warn("failed to split host and port : %s : %v", s, err)
		port = ""
	}
	return host, port
}
