package main

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/base64"
)

// signingKey is the HMAC-SHA512 key required by the Tori.fi API gateway.
// Decoded from the torium Python library's ROT13-obfuscated value.
var signingKey = []byte("3b535f36-79be-424b-a6fd-116c6e69f137")

// gwKey computes the finn-gw-key header for a request.
//
// Message format: {METHOD};{path}{?query};{finn-gw-service};{body bytes}
func gwKey(method, path, service string, body []byte, query string) string {
	if path == "/" {
		path = ""
	}
	queryPart := ""
	if query != "" {
		queryPart = "?" + query
	}
	prefix := method + ";" + path + queryPart + ";" + service + ";"
	msg := append([]byte(prefix), body...)

	mac := hmac.New(sha512.New, signingKey)
	mac.Write(msg)
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}
