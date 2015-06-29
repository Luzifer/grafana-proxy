# Luzifer / grafana-proxy

[![License: Apache v2.0](https://badge.luzifer.io/v1/badge?color=5d79b5&title=license&text=Apache+v2.0)](http://www.apache.org/licenses/LICENSE-2.0)
[![Binary Download](https://badge.luzifer.io/v1/badge?title=Download&text=on GoBuilder)](https://gobuilder.me/github.com/Luzifer/grafana-proxy)

This project emerged from the wish to display my grafana dashboard on a screen attached to a Raspberry-PI without having to login myself. Also the chromium browser running on that PI is running in incognito mode. (Even if it wasn't the login with stored credentials would required manual interaction.) So I was in need for something to display the dasboards without the snapshot function of grafana (I change that dashboard quite often and also I don't trust that public snapshot service) without having to do anything to interact with that PI.

## How does it work

1. You start the `grafana-proxy` with an username, password and the base-url of your grafana instance
2. You call the browser to open `http://localhost:8081/dashboard_url`
3. You can watch your dashboard without manually logging in

```bash
# ./grafana-proxy
Usage of ./grafana-proxy:
      --baseurl="": BaseURL (excluding last /) of Grafana
  -p, --pass="": Password for Grafana login
  -u, --user="": Username for Grafana login

# ./grafana-proxy -u [...] -p [...] --baseurl=https://grafana
2015/06/29 23:45:19 GET /? 200 2421
2015/06/29 23:45:19 GET /css/grafana.dark.min.5aa0b879.css? 200 185096
2015/06/29 23:45:19 GET /app/app.6e379bdb.js? 200 874636
```
