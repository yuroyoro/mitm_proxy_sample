package main

import (
	// "flag"
	"fmt"
	// "log"
	"net/http"
	"net/http/httputil"

	"github.com/elazarl/goproxy"
)

// func main() {
// verbose := flag.Bool("v", false, "should every proxy request be logged to stdout")
// addr := flag.String("addr", ":4080", "proxy listen address")
// flag.Parse()
// proxy := goproxy.NewProxyHttpServer()
// proxy.OnRequest().DoFunc(onRequestProxyHandler)
// proxy.OnResponse().DoFunc(onResponseProxyHandler)
// proxy.OnRequest().HandleConnect(goproxy.AlwaysMitm)

// proxy.Verbose = *verbose
// log.Println("Starting Proxy listend: 4080")
// log.Fatal(http.ListenAndServe(*addr, proxy))
// }

// func orPanic(err error) {
// if err != nil {
// panic(err)
// }
// }

func onRequestProxyHandler(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
	fmt.Println("---------------------------------------------------------------------")
	fmt.Printf("Request Url : %s\n", req.URL)
	dump, _ := httputil.DumpRequestOut(req, true)
	fmt.Printf("Request: %s\n", dump)
	fmt.Println("---------------------------------------------------------------------")

	return req, nil
}

func onResponseProxyHandler(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
	dumpResp, _ := httputil.DumpResponse(resp, true)
	fmt.Println("---------------------------------------------------------------------")
	fmt.Printf("Response: %s\n", dumpResp)
	fmt.Println("---------------------------------------------------------------------")

	return resp
}
