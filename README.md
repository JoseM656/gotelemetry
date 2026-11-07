# go!telemetry

Robust monolithic daemon for telemetry and power consumption monitoring designed for Linux servers.

```json
{"hostname":"MACHINA-GEN-X","version":"v1","timestamp":"2026-11-07T20:36:04Z","local_timestamp":"2026-11-07T17:36:04-03:00","component":"cpu","interval":5000,"pretty_interval":"5s","data":{"User":1254318,"Nice":348,"System":376742,"Idle":34050078,"Iowait":40766,"IRQ":103512,"SoftIRQ":51245,"Steal":0}}
```

## Installation

```bash
git clone <repo-url>
cd gotelemetry
make install
```

If you use NixOS or home-manager you can:

```bash
# For only run
nix run github:JoseM656/gotelemetry

# To install
nix profile install github:JoseM656/gotelemetry

```

## Config

An easy format of configuration in yml:

```yml
Output:
    local_timestamp:
        enable: false
    # optional: add an extra field to the JSON containing the interval in seconds or minutes, not milliseconds.
    pretty_interval_output:
        enable: true 
collectors:
    cpu:
        enabled: true
        interval: 5s
    ram:
        enabled: true
        interval: 5s
    swap:
        enabled: true
        interval: 5s
    storage:
        enabled: false
        interval: 120s
    gpu:
        enabled: false
        interval: 30s

```

> This will be updated with more options as development progresses.

### Options

- `local_timestamp` add an extra field to the JSON (`"local_timestamp":"2026-11-07T17:36:04-03:00"`) containing the local time, default is in UTC.
- `pretty_interval_output` add an extra field to the JSON (`"pretty_interval":"5s"`) containing the interval in seconds or minutes, not milliseconds.

You can regen the config file using `gotelemetry -regen`

### Config path

For this monolith you can change the `config.yml` path, use a **env variable**:

```env
export GOTELEMETRY_CONFIG_PATH="$HOME/.config/gotelemetry/config.yml"
# Supports "~/"
export GOTELEMETRY_CONFIG_PATH="~/.config/gotelemetry/config.yml"
```

This is useful if you dont want to use sudo, because the default path is `/etc/gotelemetry/config.yml`. Also can hardcode a path using `make install CONFIG_PATH="/etc/gotelemetry/config.yml"`. But you will have to recompile for change it.

## Usage

```bash
# Normal use
gotelemetry

# ctrl + c for exit.

# options
gotelemetry -h
```
