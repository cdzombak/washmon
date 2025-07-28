# washmon

Monitor washing machine power usage from InfluxDB and send notifications when a load is done.

## Usage

```text
washmon -config /path/to/config.json [-version]
```

### Options

- `-config`: Path to the configuration JSON file. Required.
- `-version`: Print version and exit.

### Configuration

Configuration is provided by a JSON file, which contains the following fields:

#### InfluxDB Configuration

- `influx_server`: InfluxDB server URL.
- `influx_org`: InfluxDB organization (optional).
- `influx_user`, `influx_password`: InfluxDB credentials (optional if using token).
- `influx_token`: InfluxDB token (optional if using user/password).
- `influx_health_check_disabled`: If set to `true`, skip checking the Influx server's health before querying.
- `influx_timeout_s`: InfluxDB query timeout in seconds (default: 10).

#### Power Monitoring Configuration

- `power_mean_running_threshold`: Power threshold in watts to determine if the machine is running (default: 5.0).
- `prior_window_power_mean_query`: InfluxDB query to get the previous power window mean.
- `current_window_power_mean_query`: InfluxDB query to get the current power window mean.

#### Notification Configuration

- `ntfy_server`: ntfy server URL (e.g., "https://ntfy.sh").
- `ntfy_token`: ntfy access token (optional).
- `ntfy_topic`: ntfy topic to publish notifications to.
- `ntfy_timeout_s`: ntfy request timeout in seconds (default: 10).
- `ntfy_tags`: Comma-separated list of ntfy tags (optional).
- `ntfy_priority`: ntfy message priority, 1-5 (default: 3).
- `notify_every_minutes`: How often to send reminder notifications in minutes (default: 30).

#### API Configuration

- `api_port`: Port for the web API (default: 8080).
- `api_root`: Base URL for the web API (default: "http://localhost:8080").
- `state_file`: Path to persist application state (optional).

A sample config file is included in this repository to help you get started: [`config.example.json`](https://github.com/cdzombak/washmon/blob/main/config.example.json).

### How It Works

`washmon` continuously monitors washing machine power usage by querying InfluxDB for power measurements. It compares the current power window mean with a previous window to detect state transitions:

1. **Clear**: Machine is idle, no notifications needed
2. **Running**: Machine is actively running a cycle
3. **Done**: Machine has finished and needs to be emptied

When the machine transitions to "Done", `washmon` sends periodic notifications via ntfy with action buttons to:
- **✅ I emptied it**: Acknowledges the machine has been emptied (transitions to Clear)
- **💤 Mute 3h**: Temporarily mutes notifications for 3 hours

The web API provides endpoints for these actions, allowing integration with home automation systems or manual acknowledgment via web requests.

## Installation

### macOS via Homebrew

```shell
brew install cdzombak/oss/washmon
```

### Debian via apt repository

Install my Debian repository if you haven't already:

```shell
sudo apt-get install ca-certificates curl gnupg
sudo install -m 0755 -d /etc/apt/keyrings
curl -fsSL https://dist.cdzombak.net/deb.key | sudo gpg --dearmor -o /etc/apt/keyrings/dist-cdzombak-net.gpg
sudo chmod 0644 /etc/apt/keyrings/dist-cdzombak-net.gpg
echo -e "deb [signed-by=/etc/apt/keyrings/dist-cdzombak-net.gpg] https://dist.cdzombak.net/deb/oss any oss\n" | sudo tee -a /etc/apt/sources.list.d/dist-cdzombak-net.list > /dev/null
sudo apt-get update
```

Then install `washmon` via `apt-get`:

```shell
sudo apt-get install washmon
```

### Manual installation from build artifacts

Pre-built binaries for Linux and macOS on various architectures are downloadable from each [GitHub Release](https://github.com/cdzombak/washmon/releases). Debian packages for each release are available as well.

### Build and install locally

```shell
git clone https://github.com/cdzombak/washmon.git
cd washmon
make build

cp out/washmon $INSTALL_DIR
```

## Docker images

Docker images are available for a variety of Linux architectures from [Docker Hub](https://hub.docker.com/r/cdzombak/washmon) and [GHCR](https://github.com/cdzombak/washmon/pkgs/container/washmon). Images are based on the `scratch` image and are as small as possible.

Run them via, for example:

```shell
docker run --rm -v ./my/config.json:/config.json:ro cdzombak/washmon:1
docker run --rm -v ./my/config.json:/config.json:ro ghcr.io/cdzombak/washmon:1
```

The default Docker command is `["-config", "/config.json"]`, so you can mount your config file at that path.

## Example Usage

This runs on my home server as a systemd service, monitoring my washing machine's power usage and sending notifications to my phone when loads are complete.

## About

- Issues: [github.com/cdzombak/washmon/issues](https://github.com/cdzombak/washmon/issues)
- Author: [Chris Dzombak](https://www.dzombak.com)
  - [GitHub: @cdzombak](https://www.github.com/cdzombak)

## License

MIT; see `LICENSE` in this repository.
