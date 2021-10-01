# InfluxDB IOx Datasource Graafana plugin

**Note**: this is currently an experimental proof-of-concept plugin.

## What is it?
This is a Grafana datasource backend plugin for querying an InfluxDB IOx server via its SQL frontend over Arrow Flight.
Currently the plugin provides the ability to manage a connection to a database on IOx, via its gRPC API, and also to execute SQL queries against that database.

## How do I use it?
The plugin is not currently signed so you will need to explicitly give Grafana permission to use it.
To do that you need to either set the following in your `grafana.ini`

```ini
[plugins]
;enable_alpha = false
;app_tls_skip_verify_insecure = false
# Enter a comma-separated list of plugin identifiers to identify plugins that are allowed to be loaded even if they lack a valid signature.
allow_loading_unsigned_plugins = influxdata-influxdb-iox-grafana
;marketplace_url = https://grafana.com/grafana/plugins/
```

or you will need to set the relevant environment variable:

```
GF_PLUGINS_ALLOW_LOADING_UNSIGNED_PLUGINS=influxdata-influxdb-iox-grafana
```

Oh, you will need the plugin of course! Grab the archive from the releases page on the repo.
Unpack it an put it somewhere. 
You will then need to tell Grafana where to look for the plugin, which you can do either by modifying the Grafana config in `grafana.ini`:

```
#################################### Paths ####################################
[paths]
# Path to where grafana can store temp files, sessions, and the sqlite3 db (if that is used)
;data = /var/lib/grafana

# Temporary files in `data` directory older than given duration will be removed
;temp_data_lifetime = 24h

# Directory where grafana can store logs
;logs = /var/log/grafana

# Directory where grafana will automatically scan and look for plugins
plugins = /Users/me/grafana-dev/plugins/
```

or by using an env var:

```
GF_PATHS_PLUGINS=/Users/me/grafana-dev/plugins/
```

If that all works out then when you restart Grafana it should find it and you should be able to see it as an unsigned datasource.

## Contributing

A data source backend plugin consists of both frontend and backend components.

### Frontend

1. Install dependencies

   ```bash
   yarn install
   ```

2. Build plugin in development mode or run in watch mode

   ```bash
   yarn dev
   ```

   or

   ```bash
   yarn watch
   ```

3. Build plugin in production mode

   ```bash
   yarn build
   ```

### Backend

1. Update [Grafana plugin SDK for Go](https://grafana.com/docs/grafana/latest/developers/plugins/backend/grafana-plugin-sdk-for-go/) dependency to the latest minor version:

   ```bash
   go get -u github.com/grafana/grafana-plugin-sdk-go
   go mod tidy
   ```

2. Build backend plugin binaries for Linux, Windows and Darwin:

   ```bash
   mage -v
   ```

3. List all available Mage targets for additional commands:

   ```bash
   mage -l
   ```

## Learn more

- [Build a data source backend plugin tutorial](https://grafana.com/tutorials/build-a-data-source-backend-plugin)
- [Grafana documentation](https://grafana.com/docs/)
- [Grafana Tutorials](https://grafana.com/tutorials/) - Grafana Tutorials are step-by-step guides that help you make the most of Grafana
- [Grafana UI Library](https://developers.grafana.com/ui) - UI components to help you build interfaces using Grafana Design System
- [Grafana plugin SDK for Go](https://grafana.com/docs/grafana/latest/developers/plugins/backend/grafana-plugin-sdk-for-go/)
