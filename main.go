package main

import (
	"flag"
	"fmt"
	"log"

	"net/http"
	"net/http/httputil"

	"crypto/tls"
)

// MiTMProxy : proxy instance
type MiTMProxy struct {
	mitm      bool
	transport *http.Transport
	signingCertificate
}

func main() {
	mitm := flag.Bool("mitm", false, "enable mitm sniffing on https")
	addr := flag.String("addr", ":4080", "proxy listen address")
	certfile := flag.String("cert-pem", "", "ca cert file")
	keyfile := flag.String("key-pem", "", "ca key file")

	flag.Parse()
	proxy := newProxy(*mitm, *certfile, *keyfile)

	proxy.info("Starting Proxy listend: %s : mitm %v", *addr, *mitm)
	log.Fatal(http.ListenAndServe(*addr, proxy))
}

func (proxy *MiTMProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodConnect {
		if proxy.mitm {
			proxy.mitmRequest(w, r)
		} else {
			proxy.relayHTTPSRequest(w, r)
		}
		return
	}

	proxy.transportHTTPRequest(w, r)
}

func (proxy *MiTMProxy) info(s string, v ...interface{}) {
	logging("Info", s, v...)
}

func (proxy *MiTMProxy) warn(s string, v ...interface{}) {
	logging("Warn", s, v...)
}
func (proxy *MiTMProxy) error(s string, v ...interface{}) {
	logging("Info", s, v...)
}

func logging(level, s string, v ...interface{}) {
	msg := fmt.Sprintf(s, v...)
	log.Printf("[%s] %s\n", level, msg)
}

func dumpRequest(req *http.Request) {
	fmt.Println("---------------------------------------------------------------------")
	fmt.Printf("-> Request : %s %s\n", req.Method, req.URL)
	dump, _ := httputil.DumpRequestOut(req, true)
	fmt.Println(string(dump))
	fmt.Println("---------------------------------------------------------------------")
}

func dumpResponse(resp *http.Response) {
	dumpResp, _ := httputil.DumpResponse(resp, true)
	fmt.Println("---------------------------------------------------------------------")
	fmt.Printf("<- Response: %s %s\n", resp.Request.Method, resp.Request.URL)
	fmt.Println(string(dumpResp))
	fmt.Println("---------------------------------------------------------------------")
}

func newProxy(mitm bool, certfile, keyfile string) *MiTMProxy {
	proxy := &MiTMProxy{
		mitm:      mitm,
		transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, Proxy: http.ProxyFromEnvironment},
	}

	proxy.setupCert(certfile, keyfile)
	return proxy
}
