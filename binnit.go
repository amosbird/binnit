/*
 *  This program is free software: you can redistribute it and/or
 *  modify it under the terms of the GNU Affero General Public License as
 *  published by the Free Software Foundation, either version 3 of the
 *  License, or (at your option) any later version.
 *
 *  This program is distributed in the hope that it will be useful,
 *  but WITHOUT ANY WARRANTY; without even the implied warranty of
 *  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
 *  General Public License for more details.
 *
 *  You should have received a copy of the GNU Affero General Public
 *  License along with this program.  If not, see
 *  <http://www.gnu.org/licenses/>.
 *
 *  (c) Vincenzo "KatolaZ" Nicosia 2017 -- <katolaz@freaknet.org>
 *
 *
 *  This file is part of "binnit", a minimal no-fuss pastebin-like
 *  server written in golang
 *
 */

package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/crypto/bcrypt"

	auth "github.com/abbot/go-http-auth"
	"github.com/amosbird/binnit/paste"
)

var userpass = flag.String("g", "", "Generate user password")

var p_conf = Config{
	server_name: "oracle.wentropy.com",
	bind_addr:   "0.0.0.0",
	bind_port:   "443",
	paste_dir:   "./pastes",
	templ_dir:   "./tmpl",
	max_size:    10000000,
	log_file:    "./binnit.log",
}

func min(a, b int) int {
	if a > b {
		return b
	} else {
		return a
	}
}

func handle_get_paste(w http.ResponseWriter, r *auth.AuthenticatedRequest) {
	var paste_name, orig_name string
	orig_name = filepath.Clean(r.URL.Path)
	paste_name = p_conf.paste_dir + "/" + orig_name
	orig_IP := r.RemoteAddr
	log.Printf("Received GET from %s for  '%s'\n", orig_IP, orig_name)
	if strings.HasSuffix(paste_name, ".html") {
		paste_name = strings.TrimSuffix(paste_name, ".html")
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
	}
	if orig_name == "/" {
		w.Write([]byte("<!DOCTYPE html><html><head><title>PasteBinServer</title></head><body><h1>Amos Bird's Pastebin Server</h1></body></html>"))
	} else {
		// otherwise, if the requested paste exists, we serve it...
		content, err := paste.Retrieve(paste_name)
		if err == nil {
			w.Write(content)
			return
		} else {
			// otherwise, we give say we didn't find it
			fmt.Fprintf(w, "%s\n", err)
			return
		}
	}
}

func handle_put_paste(w http.ResponseWriter, r *auth.AuthenticatedRequest) {
	err1 := r.ParseForm()
	err2 := r.ParseMultipartForm(int64(2 * p_conf.max_size))
	if err1 != nil && err2 != nil {
		w.Write([]byte("<!DOCTYPE html><html><head><title>PasteBin Server</title></head><body><h1>Amos Bird's Pastebin Server</h1></body></html>"))
	} else {
		req_body := r.PostForm
		orig_IP := r.RemoteAddr
		log.Printf("Received new POST from %s\n", orig_IP)
		content := req_body.Get("paste")
		ID, err := paste.Store(content, p_conf.paste_dir)
		log.Printf("   ID: %s; err: %s\n", ID, err)
		if err == nil {
			hostname := p_conf.server_name
			if show := req_body.Get("show"); show != "1" {
				fmt.Fprintf(w, "https://%s/%s\n", hostname, ID)
				return
			} else {
				fmt.Fprintf(w, "<html><body>Link: <a href='https://%s/%s'>https://%s/%s</a></body></html>",
					hostname, ID, hostname, ID)
				return
			}
		} else {
			fmt.Fprintf(w, "%s\n", err)
		}
	}
}

func req_handler(w http.ResponseWriter, r *auth.AuthenticatedRequest) {
	switch r.Method {
	case "GET":
		handle_get_paste(w, r)
	case "POST":
		handle_put_paste(w, r)
	}
}

func secret(user, realm string) string {
	if user == "amos" {
		return "$2a$10$rLY7MbjM.nz4y55yCIMIHuHf1Cse7PLs.r9yAN9HwGlGkYT4ZF6tS"
	}
	return ""
}

func Wrap(a *auth.BasicAuth, wrapped auth.AuthenticatedHandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			if username := a.CheckAuth(r); username == "" {
				a.RequireAuth(w, r)
			} else {
				ar := &auth.AuthenticatedRequest{Request: *r, Username: username}
				wrapped(w, ar)
			}
		} else {
			ar := &auth.AuthenticatedRequest{Request: *r, Username: "amos"}
			wrapped(w, ar)
		}
	}
}

func getCertificate(info *tls.ClientHelloInfo) (*tls.Certificate, error) {
	caFiles, err := tls.LoadX509KeyPair("server.crt", "server.key")
	if err != nil {
		return nil, err
	}

	return &caFiles, nil
}

func main() {
	flag.Parse()
	if *userpass != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(*userpass), bcrypt.DefaultCost)
		if err == nil {
			fmt.Println(string(hashedPassword))
		}
		os.Exit(0)
	}
	f, err := os.OpenFile(p_conf.log_file, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening log_file: %s. Exiting\n", p_conf.log_file)
		os.Exit(1)
	}
	defer f.Close()
	log.SetOutput(io.Writer(f))
	log.SetPrefix("[binnit]: ")
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)
	log.Println("Binnit version 0.1 -- Starting ")
	log.Printf("  + Serving pastes on: %s\n", p_conf.server_name)
	log.Printf("  + listening on: %s:%s\n", p_conf.bind_addr, p_conf.bind_port)
	log.Printf("  + paste_dir: %s\n", p_conf.paste_dir)
	log.Printf("  + templ_dir: %s\n", p_conf.templ_dir)
	log.Printf("  + max_size: %d\n", p_conf.max_size)
	authenticator := auth.NewBasicAuthenticator(p_conf.server_name, secret)
	http.HandleFunc("/", Wrap(authenticator, req_handler))

	s := &http.Server{
		Addr: "0.0.0.0:443",
		TLSConfig: &tls.Config{
			GetCertificate: getCertificate,
		},
	}

	log.Fatal(s.ListenAndServeTLS("", ""))
}
