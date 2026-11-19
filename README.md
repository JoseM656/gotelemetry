# go!telemetry

Robust monolithic daemon for telemetry and power consumption monitoring designed for Linux servers. Export information about various components using json in stdout, log files, and more.

```json
{"hostname":"MACHINA-GEN-X","version":"v1","timestamp":"2026-11-07T20:36:04Z","local_timestamp":"2026-11-07T17:36:04-03:00","component":"cpu","interval":5000,"pretty_interval":"5s","data":{"User":1254318,"Nice":348,"System":376742,"Idle":34050078,"Iowait":40766,"IRQ":103512,"SoftIRQ":51245,"Steal":0}}
```

## Installation

For installation you can download the releases or use make:

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
general:
    local_timestamp: 
        enabled: false
    pretty_interval_output: 
        enabled: true
output:
    logger:
        enabled: true
        path: "~/Documentos/logs" # put here where do you want .log files
    stdout:
        enabled: false
collectors:
    cpu:
        enabled: true
        interval: 500ms
    ram:
        enabled: true
        interval: 500ms
    swap:
        enabled: true
        interval: 500ms
    storage:
        enabled: false
        interval: 1h
    gpu:
        enabled: false
        interval: 5s
```

> Use units like ms, s or h.

### Options

- `local_timestamp` add an extra field to the JSON (`"local_timestamp":"2026-11-07T17:36:04-03:00"`) containing the local time, default is in UTC.
- `pretty_interval_output` add an extra field to the JSON (`"pretty_interval":"5s"`) containing the interval in seconds or minutes, not milliseconds.

You can regen the config file using `gotelemetry regen path/that/you/want`

### Config path

For this monolith you can change the `config.yml` path, use a **env variable**:

```env
export GOTELEMETRY_CONFIG_PATH="$HOME/.config/gotelemetry/config.yml"
# Supports "~/"
export GOTELEMETRY_CONFIG_PATH="~/.config/gotelemetry/config.yml"
```

This is useful if you dont want to use sudo, because the default path is `/etc/gotelemetry/config.yml`.

### CLI

On the CLI you can regenerate or select a config file using:

```bash
gotelemtery regen --path ~/.config/gotelemetry/config.yml

# And for running that configuration
gotelemetry --path ~/.config/gotelemetry/config.yml
```

In that way you can have multiple configurations files.

## Usage

```bash
# Normal use
gotelemetry

# ctrl + c for exit.

# options
gotelemetry help
```

---

This project is still ongoing, so further changes may occur. Thanks for your interest! :D
