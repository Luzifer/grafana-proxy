package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"time"

	"github.com/Luzifer/rconfig"
	"github.com/cenkalti/backoff"
)

var (
	cfg = struct {
		User    string `flag:"user,u" default:"" description:"Username for Grafana login"`
		Pass    string `flag:"pass,p" default:"" description:"Password for Grafana login"`
		BaseURL string `flag:"baseurl" default:"" description:"BaseURL (excluding last /) of Grafana"`
		Listen  string `flag:"listen" default:"127.0.0.1:8081" description:"IP/Port to listen on"`
		Token   string `flag:"token" default:"" description:"(optional) require a ?token=xyz parameter to show the dashboard"`
	}{}
	cookieJar *cookiejar.Jar
	client    *http.Client
	base      *url.URL
)

func init() {
	rconfig.Parse(&cfg)

	if cfg.User == "" || cfg.Pass == "" || cfg.BaseURL == "" {
		rconfig.Usage()
		os.Exit(1)
	}

	cookieJar, _ = cookiejar.New(nil)
	client = &http.Client{
		Jar: cookieJar,
	}
}

func loadLogin() {
	backoff.Retry(func() error {
		resp, err := client.PostForm(fmt.Sprintf("%s/login", cfg.BaseURL), url.Values{
			"user":     {cfg.User},
			"password": {cfg.Pass},
		})
		if err != nil {
			log.Printf("[ERR][loadLogin] %s", err)
			return err
		}
		defer resp.Body.Close()
		return nil
	}, backoff.NewExponentialBackOff())
}

type proxy struct{}

func (p proxy) ServeHTTP(res http.ResponseWriter, r *http.Request) {
	bo := backoff.NewExponentialBackOff()
	bo.MaxElapsedTime = 5 * time.Second
	if err := backoff.Retry(func() error {
		r.URL.Host = base.Host
		r.URL.Scheme = base.Scheme
		r.RequestURI = ""
		r.Host = base.Host

		if cfg.Token != "" && r.URL.Query().Get("token") != cfg.Token {
			http.Error(res, "Please add the `?token=xyz` parameter with correct token", http.StatusForbidden)
			return nil
		}

		resp, err := client.Do(r)
		if err != nil {
			return err
		}

		defer resp.Body.Close()

		res.Header().Del("Content-Type")
		for k, v := range resp.Header {
			for _, v1 := range v {
				res.Header().Add(k, v1)
			}
		}

		if resp.StatusCode == 401 {
			loadLogin()
			return fmt.Errorf("Need to relogin")
		}

		res.WriteHeader(resp.StatusCode)
		written, _ := io.Copy(res, resp.Body)

		log.Printf("%s %s?%s %d %d\n", r.Method, r.URL.Path, r.URL.RawQuery, resp.StatusCode, written)
		return nil
	}, bo); err != nil {
		http.Error(res, fmt.Sprintf("Woot?\n%s", err), http.StatusInternalServerError)
	}
}

func main() {
	loadLogin()

	var err error
	base, err = url.Parse(cfg.BaseURL)
	if err != nil {
		fmt.Printf("Please provide a parseable baseurl: %s\n", err)
	}

	log.Fatal(http.ListenAndServe(cfg.Listen, proxy{}))
}
