package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"time"

	"github.com/Luzifer/rconfig"
	"github.com/cenkalti/backoff"
	log "github.com/sirupsen/logrus"
)

var (
	cfg = struct {
		User      string `flag:"user,u" default:"" env:"USER" description:"Username for Grafana login"`
		Pass      string `flag:"pass,p" default:"" env:"PASS" description:"Password for Grafana login"`
		BaseURL   string `flag:"baseurl" default:"" env:"BASEURL" description:"BaseURL (excluding last /) of Grafana"`
		Listen    string `flag:"listen" default:"127.0.0.1:8081" description:"IP/Port to listen on"`
		Token     string `flag:"token" default:"" env:"TOKEN" description:"(optional) require a ?token=xyz parameter to show the dashboard"`
		LogFormat string `flag:"log-format" default:"text" env:"LOG_FORMAT" description:"Output format for logs (text/json)"`
	}{}
	cookieJar *cookiejar.Jar
	client    *http.Client
	base      *url.URL
)

func init() {
	if err := rconfig.Parse(&cfg); err != nil {
		log.Fatalf("Unable to parse commandline options: %s", err)
	}

	switch cfg.LogFormat {
	case "text":
		log.SetFormatter(&log.TextFormatter{})
	case "json":
		log.SetFormatter(&log.JSONFormatter{})
	default:
		log.Fatalf("Unknown log format: %s", cfg.LogFormat)
	}

	log.SetLevel(log.InfoLevel)

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
			log.WithError(err).WithFields(log.Fields{
				"user": cfg.User,
			}).Error("Login failed")
			return err
		}
		defer resp.Body.Close()
		return nil
	}, backoff.NewExponentialBackOff())
}

type proxy struct{}

func (p proxy) ServeHTTP(res http.ResponseWriter, r *http.Request) {
	requestLog := log.WithFields(log.Fields{
		"http_user_agent": r.Header.Get("User-Agent"),
		"host":            r.Host,
		"remote_addr":     r.Header.Get("X-Forwarded-For"),
		"request":         r.URL.Path,
		"request_full":    r.URL.String(),
		"request_method":  r.Method,
	})

	bo := backoff.NewExponentialBackOff()
	bo.MaxElapsedTime = 5 * time.Second
	if err := backoff.Retry(func() error {
		r.URL.Host = base.Host
		r.URL.Scheme = base.Scheme
		r.RequestURI = ""
		r.Host = base.Host

		suppliedToken := r.URL.Query().Get("token")
		if authCookie, err := r.Cookie("grafana-proxy-auth"); err == nil {
			suppliedToken = authCookie.Value
		}

		if cfg.Token != "" && suppliedToken != cfg.Token {
			requestLog.Error("Token parameter is wrong")
			http.Error(res, "Please add the `?token=xyz` parameter with correct token", http.StatusForbidden)
			return nil
		}

		resp, err := client.Do(r)
		if err != nil {
			requestLog.WithError(err).Error("Request failed")
			return err
		}

		defer resp.Body.Close()

		res.Header().Del("Content-Type")
		for k, v := range resp.Header {
			for _, v1 := range v {
				res.Header().Set(k, v1)
			}
		}

		if r.URL.Query().Get("token") != "" {
			http.SetCookie(res, &http.Cookie{
				Name:   "grafana-proxy-auth",
				Value:  r.URL.Query().Get("token"),
				MaxAge: 31536000, // 1 Year
				Path:   "/",
			})
		}

		if resp.StatusCode == 401 {
			errmsg, _ := ioutil.ReadAll(resp.Body)
			requestLog.WithFields(log.Fields{
				"error":  string(errmsg),
				"header": r.Header,
			}).Info("Unauthorized, trying to login")
			loadLogin()
			return fmt.Errorf("Need to relogin")
		}

		res.WriteHeader(resp.StatusCode)
		written, _ := io.Copy(res, resp.Body)

		requestLog.WithFields(log.Fields{
			"status":     resp.StatusCode,
			"bytes_sent": written,
		}).Info("Request completed")
		return nil
	}, bo); err != nil {
		requestLog.WithError(err).WithFields(log.Fields{
			"status": http.StatusInternalServerError,
		}).Error("Backend request failed")
		http.Error(res, fmt.Sprintf("Woot?\n%s", err), http.StatusInternalServerError)
	}
}

func main() {
	loadLogin()

	var err error
	base, err = url.Parse(cfg.BaseURL)
	if err != nil {
		log.WithError(err).WithField("base_url", base).Fatalf("BaseURL is not parsesable")
	}

	log.Fatal(http.ListenAndServe(cfg.Listen, proxy{}))
}
