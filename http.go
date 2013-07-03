package oauth

import (
	"bufio"
	"crypto/tls"
	"io"
	"net"
	"net/http"
	"strings"
	"time"
)

// get Taken from the golang source modifed to allow headers to be passed and no redirection allowed
func get(url_ string, headers map[string]string, timeout int64) (r *http.Response, err error) {

	req, err := http.NewRequest("GET", url_, nil)
	if err != nil {
		return
	}
	for k, v := range headers {
		req.Header.Add(k, v)
	}

	r, err = send(req, timeout)
	if err != nil {
		return
	}
	return
}

// post taken from Golang modified to allow Headers to be pased
func post(url_ string, headers map[string]string, body io.Reader, timeout int64, size int) (r *http.Response, err error) {
	req, err := http.NewRequest("POST", url_, nopCloser{body})
	req.ContentLength = int64(size)
	if err != nil {
		return
	}
	req.ProtoMajor = 1
	req.ProtoMinor = 1
	req.Close = true
	for k, v := range headers {
		req.Header.Add(k, v)
	}

	return send(req, timeout)
}

// Copyright (c) 2009 The Go Authors. All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are
// met:
//
//    * Redistributions of source code must retain the above copyright
// notice, this list of conditions and the following disclaimer.
//    * Redistributions in binary form must reproduce the above
// copyright notice, this list of conditions and the following disclaimer
// in the documentation and/or other materials provided with the
// distribution.
//    * Neither the name of Google Inc. nor the names of its
// contributors may be used to endorse or promote products derived from
// this software without specific prior written permission.
//

// From the http package - modified to allow Headers to be sent to the Post method
type nopCloser struct {
	io.Reader
}

func (nopCloser) Close() error { return nil }

type TimeoutReaderCloser struct {
	R io.Reader
	C net.Conn
	T int64
}

func (r *TimeoutReaderCloser) Read(b []byte) (int, error) {
	size, err := r.R.Read(b)
	if err != nil {
		return size, err
	}
	t := time.Now().Add(time.Duration(r.T))
	r.C.SetReadDeadline(t)
	return size, err
}

func (r *TimeoutReaderCloser) Close() error {
	return r.C.Close()
}

func send(req *http.Request, timeout int64) (resp *http.Response, err error) {
	if req.URL.Scheme != "http" && req.URL.Scheme != "https" {
		return nil, nil
	}

	addr := req.URL.Host
	if !hasPort(addr) {
		addr += ":" + req.URL.Scheme
	}
	/*info := req.URL.Userinfo
	  if len(info) > 0 {
	      enc := base64.URLEncoding
	      encoded := make([]byte, enc.EncodedLen(len(info)))
	      enc.Encode(encoded, []byte(info))
	      if req.Header == nil {
	          req.Header = make(map[string]string)
	      }
	      req.Header["Authorization"] = "Basic " + string(encoded)
	  }
	*/
	var conn net.Conn
	if req.URL.Scheme == "http" {
		if timeout > 0 {
			conn, err = net.DialTimeout("tcp", addr, time.Duration(timeout*1e9))
		} else {
			conn, err = net.Dial("tcp", addr)
		}
	} else { // https
		conn, err = tls.Dial("tcp", addr, nil)
	}

	if err != nil {
		return nil, err
	}
	fmt.Println("0")

	if timeout > 0 {
		conn.SetReadDeadline(time.Now().Add(time.Duration(timeout)))
	}
	fmt.Println("1")
	err = req.Write(conn)
	fmt.Println("2")

	if err != nil {
		conn.Close()
		return nil, err
	}
	fmt.Println("3")
	reader := bufio.NewReader(conn)
	fmt.Println("4")
	resp, err = http.ReadResponse(reader, req)
	fmt.Println("5")

	if err != nil {
		conn.Close()
		return nil, err
	}

	resp.Body = &TimeoutReaderCloser{R: resp.Body, C: conn, T: timeout}

	return
}

func hasPort(s string) bool { return strings.LastIndex(s, ":") > strings.LastIndex(s, "]") }
