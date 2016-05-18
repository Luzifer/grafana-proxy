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
      --listen="127.0.0.1:8081": IP/Port to listen on
  -p, --pass="": Password for Grafana login
      --token="": (optional) require a ?token=xyz parameter to show the dashboard
  -u, --user="luzifer": Username for Grafana login

# ./grafana-proxy -u [...] -p [...] --baseurl=https://grafana
2015/06/29 23:45:19 GET /? 200 2421
2015/06/29 23:45:19 GET /css/grafana.dark.min.5aa0b879.css? 200 185096
2015/06/29 23:45:19 GET /app/app.6e379bdb.js? 200 874636
```

### Starting from docker

Starting with version `v0.3.0` the proxy also supports being started using docker. To use it you can just use this command line:

```bash
# docker run --rm -ti -e USER=[...] -e PASS=[...] -e BASE_URL=[...] -p 3000:3000 quay.io/luzifer/grafana-proxy
2016/05/18 11:45:35 GET /dashboard/db/host-dashboard 200 7971
2016/05/18 11:45:35 GET /api/dashboards/db/host-dashboard? 200 18202
```

### Using the `token` parameter

If you want to run the `grafana-proxy` on a public accessible host but do not want everyone to be able to see your dashboard you can add some pseudo security using a shared token:

```bash
# docker run --rm -ti -e USER=myuser -e PASS=mypass -e BASE_URL=http://mygrafana.com -e TOKEN=mysharedsecret -p 3000:3000 quay.io/luzifer/grafana-proxy
2016/05/18 11:45:35 GET /dashboard/db/host-dashboard 200 7971
2016/05/18 11:45:35 GET /api/dashboards/db/host-dashboard? 200 18202
```

Your users now need to use `http://<ip>:3000/mydashboard?token=mysharedsecret` to access the dashboard.
