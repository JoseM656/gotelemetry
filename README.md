# go!telemetry

Robust monolithic daemon for telemetry and power consumption monitoring designed for Linux servers.

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

You can regen the config file using `gotelemetry -regen`

### Config path

For this monolith you can change the `config.yml` path, use a **env variable**:

```env
export GOTELEMETRY_CONFIG_PATH="$HOME/.config/gotelemetry/config.yml"
# Supports "~/"
export GOTELEMETRY_CONFIG_PATH="~/.config/gotelemetry/config.yml"
```
This is useful if you dont want to use sudo, because the default path is `/etc/gotelemetry/config.yml`. Also can 
hardcode a path using `make install CONFIG_PATH="/etc/gotelemetry/config.yml"`. But you will have to recompile for change it.

## Usage

```bash
# Normal use
gotelemetry

# options
gotelemetry -h
```
