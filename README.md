# go!Telemetry
Monolithic daemon for telemetry and power consumption monitoring designed for Linux servers.

## Installation

```bash
git clone <repo-url>
cd gotelemetry
make install
```
Or if you want to force a path for the config file:

```bash
make build CONFIG_PATH="<any/path/that/you/want>"
```

this will be hardcoded.

### config file

Whether or not you hardcoded a path you can use a **env variable**:

```env
export GOTELEMETRY_CONFIG_PATH="$HOME/.config/gotelemetry/config.yml
```

Do not use `~`

## Usage

```bash
# Normal use
gotelemetry

# options
gotelemetry -h
```
