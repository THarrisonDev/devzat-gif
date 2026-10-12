# devzat-gif

A [Devzat](https://github.com/quackduck/devzat) plugin (Go) that posts GIFs with
a `gif` command. Uses Giphy by default, or Tenor; or paste an image URL directly.

```
gif <search terms>          Post a random matching GIF
gif <https://image-url>     Post an image URL directly (no API key needed)
```

Because Devzat renders images (it decodes the first frame only — GIFs show as a
still), the message includes the animated original's page URL too.

## Build

```bash
go build -o devzat-gif .
```

## Configure

| Env | Default | Purpose |
|---|---|---|
| `DEVZAT_ADDRESS` | — | Plugin-API address, e.g. `localhost:5556` (arg overrides) |
| `DEVZAT_TOKEN` | — | Plugin API token (required) |
| `GIF_PROVIDER` | `giphy` | `giphy` or `tenor` |
| `GIF_API_KEY` | Giphy beta key | Provider API key; set your own for reliability |
| `GIF_RATING` | `pg-13` | Giphy content rating: `g`, `pg`, `pg-13`, `r` |

The Giphy default uses Giphy's long-standing public beta key, so it works
out of the box for a small private server. For anything serious, grab a free key
at <https://developers.giphy.com/dashboard/> (or a Tenor key from Google) and set
`GIF_API_KEY`.

## Run

```bash
DEVZAT_TOKEN=... ./devzat-gif localhost:5556
```

Dry run (no server):

```bash
./devzat-gif --print cat vibing
```

## Install (systemd)

```bash
sudo DEVZAT_TOKEN=dvz@... ./install.sh
```

Reuses the same plugin-API token as `devzat-github`. See that project's README
for enabling the RPC integration (`rpc.port: 5556`) and minting a token.

## Attribution

Search results are served by Giphy or Tenor, and the bot labels each post with
`Powered by GIPHY` / `Powered by Tenor` accordingly.

## Operate

```bash
systemctl status devzat-gif
journalctl -u devzat-gif -f
systemctl restart devzat-gif
```
