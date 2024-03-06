# gpbackup Exporter

[![Actions Status](https://github.com/woblerr/gpbackup_exporter/workflows/build/badge.svg)](https://github.com/woblerr/gpbackup_exporter/actions)
[![Coverage Status](https://coveralls.io/repos/github/woblerr/gpbackup_exporter/badge.svg?branch=master)](https://coveralls.io/github/woblerr/gpbackup_exporter?branch=master)
[![Go Report Card](https://goreportcard.com/badge/github.com/woblerr/gpbackup_exporter)](https://goreportcard.com/report/github.com/woblerr/gpbackup_exporter)

Prometheus exporter for collecting metrics from [gpbackup](https://github.com/greenplum-db/gpbackup) history file `gpbackup_history.yaml`.

By default, the metrics are collected for all databases and backups in history file. You need to run exporter or Docker image on the same host where is `gpbackup_history.yaml` file located (Greenplum Master host).

## Collected metrics

The metrics provided by exporter.

* `gpbackup_backup_status` - backup status.

    Labels: backup_type, database_name, object_filtering, plugin, timestamp.

    Values description:
     - `0` - success,
     - `1` - failure.

* `gpbackup_backup_deletion_status` - backup deletion status.

    Labels: backup_type, database_name, date_deleted, object_filtering, plugin, timestamp.

    Values description:

    - `0` - backup still exists,
    - `1` - backup was successfully deleted,
    - `2` - the deletion is in progress,
    - `3` - last delete attempt failed to delete backup from plugin storage,
    - `4` - last delete attempt failed to delete backup from local storage.

* `gpbackup_backup_info` - backup info.

    Labels: backup_dir, backup_ver, backup_type, compression_type, database_name, database_ver, object_filtering, plugin, plugin_ver, timestamp, with_statistic.

    Values description:
     - `1` - info about backup is exist.

* `gpbackup_backup_duration_seconds` - backup duration in seconds.

    Labels: backup_type, database_name, object_filtering, plugin, timestamp.

* `gpbackup_backup_since_last_completion_seconds` - Seconds since the last completed backup.

    Labels: backup_type, database_name.


* `gpbackup_exporter_info` - information about gpbackup exporter.

    Labels: version.

## Getting Started
### Building and running

```bash
git clone https://github.com/woblerr/gpbackup_exporter.git
cd gpbackup_exporter
make build
./gpbackup_exporter <flags>
```

Available configuration flags:

```bash
./gpbackup_exporter --help
usage: gpbackup_exporter [<flags>]

Flags:
  -h, --[no-]help                Show context-sensitive help (also try --help-long and --help-man).
      --web.endpoint="/metrics"  Endpoint used for metrics.
      --web.listen-address=:19854 ...  
                                 Addresses on which to expose metrics and web interface. Repeatable for multiple addresses.
      --web.config.file=""       [EXPERIMENTAL] Path to configuration file that can enable TLS or authentication.
      --collect.interval=600     Collecting metrics interval in seconds.
      --collect.depth=0          Metrics depth collection in days. Metrics for backup older than this interval will not be collected.
                                 0 - disable.
      --gpbackup.history-file=""  
                                 Path to gpbackup_history.yaml.
      --gpbackup.db-include="" ...  
                                 Specific db for collecting metrics. Can be specified several times.
      --gpbackup.db-exclude="" ...  
                                 Specific db to exclude from collecting metrics. Can be specified several times.
      --gpbackup.backup-type=""  Specific backup type for collecting metrics. One of: [full, incremental, data-only, metadata-only].
      --log.level=info           Only log messages with the given severity or above. One of: [debug, info, warn, error]
      --log.format=logfmt        Output format of log messages. One of: [logfmt, json]
```

#### Additional description of flags.

It's necessary to specify the `gpbackup_history.yaml` file location via `--gpbackup.history-file` flag.

Custom database for collecting metrics can be specified via `--gpbackup.db-include` flag. You can specify several databases.<br>
For example, `--gpbackup.db-include=demo1 --gpbackup.db-include=demo2`.<br>
For this case, metrics will be collected only for `demo1` and `demo2` databases.

Custom database to exclude from collecting metrics can be specified via `--gpbackup.db-exclude` flag. You can specify several databases.<br>
For example, `--gpbackup.db-exclude=demo1 --gpbackup.db-exclude=demo2`.<br>
For this case, metrics **will not be collected** for `demo1` and `demo2` databases.<br>
If the same database is specified for include and exclude flags, then metrics for this database will not be collected. 
The flag `-gpbackup.db-exclude` has a higher priority.<br>
For example, `--gpbackup.db-include=demo1 -gpbackup.db-exclude=demo1`.<br>
For this case, metrics **will not be collected** for `demo1` database.

Custom `backup type` for collecting metrics can be specified via `--gpbackup.backup-type` flag. Valid values: `full`, `incremental`, `data-only`, `metadata-only`.<br>
For example, `--gpbackup.backup-type=full`.<br>
For this case, metrics will be collected only for `full` backups.<br>

Custom metrics depth collection in days can be specified via `--collect.depth` flag. Since gpbackup doesn't have regular options for removing info about outdated backups from history file, it is possible to limit the depth of collection metrics.<br>
For example, `--collect.depth=14`.<br> 
For this case, metrics will be collected for backups not older then 14 days from current time.<br>
Value `0` - disable this functionality.

When `--log.level=debug` is specified - information of values and labels for metrics is printing to the log.

The flag `--web.config.file` allows to specify the path to the configuration for TLS and/or basic authentication.<br>
The description of TLS configuration and basic authentication can be found at [exporter-toolkit/web](https://github.com/prometheus/exporter-toolkit/blob/v0.9.1//docs/web-configuration.md).

### Building and running docker

Environment variables supported by this image:
* `EXPORTER_ENDPOINT` - metrics endpoint, default `/metrics`;
* `EXPORTER_PORT` - port for prometheus metrics to listen on, default `19854`;
* `EXPORTER_CONFIG` - path to the configuration file for TLS and/or basic authentication, default `""`;
* `COLLECT_INTERVAL` - collecting metrics interval in seconds, default `600`;
* `COLLECT_DEPTH` - metrics depth collection in days, default `0`;
* `HISTORY_FILE` - path to gpbackup history file, default `""`;
* `DB_INCLUDE` - specific database for collecting metrics, default `""`;
* `DB_EXCLUDE` - specific database to exclude from collecting metrics, default `""`;
* `BACKUP_TYPE` - specific backup type for collecting metrics, default `""`.

#### Pull

Change `tag` to the release number.

* Docker Hub:

```bash
docker pull woblerr/gpbackup_exporter:tag
```

* GitHub Registry:

```bash
docker pull ghcr.io/woblerr/gpbackup_exporter:tag
```

#### Build

```bash
make docker
```

#### Run

Simple run:

```bash
docker run -d --restart=always \
    --name gpbackup_exporter \
    -e HISTORY_FILE=/data/gpbackup_history.yaml \
    -p 19854:19854 \
    -v /data/master/gpseg-1/gpbackup_history.yaml:/data/gpbackup_history.yaml:ro \
    gpbackup_exporter
```

For specific database:

```bash
docker run -d --restart=always \
    --name gpbackup_exporter \
    -e HISTORY_FILE=/data/gpbackup_history.yaml \
    -e DB_INCLUDE=demo1 \
    -p 19854:19854 \
    -v /data/master/gpseg-1/gpbackup_history.yaml:/data/gpbackup_history.yaml:ro \
    gpbackup_exporter
```

If you want to specify several databases for collecting metrics, 
you can run containers on different ports:

```bash
docker run -d --restart=always \
    --name gpbackup_exporter \
    -e HISTORY_FILE=/data/gpbackup_history.yaml \
    -e DB_INCLUDE=demo1 \
    -p 19854:19854 \
    -v /data/master/gpseg-1/gpbackup_history.yaml:/data/gpbackup_history.yaml:ro \
    gpbackup_exporter

docker run -d --restart=always \
    --name gpbackup_exporter \
    -e HISTORY_FILE=/data/gpbackup_history.yaml \
    -e DB_INCLUDE=demo2 \
    -p 19855:19854 \
    -v /data/master/gpseg-1/gpbackup_history.yaml:/data/gpbackup_history.yaml:ro \
    gpbackup_exporter
```

To exclude specific database:

```bash
docker run -d --restart=always \
    --name gpbackup_exporter \
    -e HISTORY_FILE=/data/gpbackup_history.yaml \
    -e DB_EXCLUDE=demo1 \
    -p 19854:19854 \
    -v /data/master/gpseg-1/gpbackup_history.yaml:/data/gpbackup_history.yaml:ro \
    gpbackup_exporter
```

For specific backup type not older than 14 days:

```bash
docker run -d --restart=always \
    --name gpbackup_exporter \
    -e HISTORY_FILE=/data/gpbackup_history.yaml \
    -e BACKUP_TYPE=full \
    -e COLLECT_DEPTH=14 \
    -p 19854:19854 \
    -v /data/master/gpseg-1/gpbackup_history.yaml:/data/gpbackup_history.yaml:ro \
    gpbackup_exporter
```

### Running as systemd service

* Register `gpbackup_exporter` (already builded, if not - exec `make build` before) as a systemd service:

```bash
make prepare-service HISTORY_FILE="/path/to/gpbackup_history.yaml"
```

Validate prepared file `gpbackup_exporter.service` and run:

```bash
sudo make install-service
```

* View service logs:

```bash
journalctl -u gpbackup_exporter.service
```

* Delete systemd service:

```bash
sudo make remove-service
```

---
Manual register systemd service:

```bash
cp gpbackup_exporter.service.template gpbackup_exporter.service
```

In file `gpbackup_exporter.service.template` replace `/usr/bin/gpbackup_exporter` to full path to `gpbackup_exporter` and `/data/master/gpseg-1/gpbackup_history.yaml` to full path to `gpbackup_history.yaml`.

```bash
sudo cp gpbackup_exporter.service /etc/systemd/system/gpbackup_exporter.service
sudo systemctl daemon-reload
sudo systemctl enable gpbackup_exporter.service
sudo systemctl restart gpbackup_exporter.service
systemctl -l status gpbackup_exporter.service
```

### RPM/DEB packages

You can use the already prepared rpm/deb package to install the exporter. Only the gpbackup_exporter binary and the service file are installed by package.

For example:
```bash
rpm -ql gpbackup_exporter

/etc/systemd/system/gpbackup_exporter.service
/usr/bin/gpbackup_exporter
```

After installation RPM/DEB package, you need to set correct path to `gpbackup_history.yaml` in `/etc/systemd/system/gpbackup_exporter.service`.


### Running tests

Run the unit tests:

```bash
make test
```

Run the end-to-end tests:

```bash
make test-e2e
```
