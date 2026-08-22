# avisavi

A CLI for searching Avios reward flights from the terminal.

Search reward flight availability, check your balance, and browse routes across
cabin classes — without opening the Avios website. It talks to the same backend
the official app uses, so prices and availability match what you'd see in the browser.

## Features

- **Balance** — check your Avios balance, including household accounts.
- **Routes** — list every reward destination from an origin with Economy/Business price ranges.
- **Availability** — seat counts per cabin (Economy, Premium, Business, First) for a route.
- **Search** — scan all destinations for seats on your exact outbound and return dates.

## Install

Build from source:

```sh
go build ./cmd/avisavi
```

Prebuilt binaries for macOS (arm64/amd64), Linux, and Windows are available from the GitHub releases.

## Usage

Log in first — the browser flow saves your session to `~/.config/avisavi/auth.json`:

```sh
avisavi login
```

Then query your account and flights:

```sh
avisavi balance
avisavi routes --origin LON
avisavi availability --origin LON --destination NYC
avisavi search --origin LON --outbound 2026-09-09 --return 2026-09-13
```

Add `--json` to any command for raw JSON output.

## Configuration

Prebuilt binaries from the releases page work out of the box — no configuration needed.

When running from source (`go run ./cmd/avisavi`), you must provide an Auth0 client ID, as there is nothing embedded yet:

```sh
export AVIOS_AUTH_CLIENT_ID=client-id
```

The client ID is resolved in this order:

1. `--client-id` flag (any run)
2. `AVIOS_AUTH_CLIENT_ID` env var (runtime)
3. Value embedded into the binary at build time (releases/CI only)

So a built binary works with no env var if the client ID was embedded via `-ldflags` when it was compiled:

```sh
go build -ldflags "-X github.com/ahobsonsayers/avisavi/pkg/gavios/auth.defaultClientID=$AVIOS_AUTH_CLIENT_ID" ./cmd/avisavi
```

> [!NOTE]
> The Avios client ID is not provided here due to liability concerns, but it's a fixed,
> unchanging value. You can extract it from the network tab of the official Avios site
> or app.

## ⚠️ Disclaimer

This tool is **unofficial** and not affiliated with, endorsed by, or supported by
Avios or its parent company. It interacts with Avios' internal APIs, which are not
public and can change without notice.

Using this tool may violate Avios' terms of service. **By using it you accept that
your account could be suspended or banned, and that any Avios points held in that
account could be lost — however unlikely.** I am not responsible for any loss of
access, points, or account standing that results from using this software.

Because login is required and misuse carries account risk, **use a dedicated burner
account** rather than your primary one, and never share your credentials.
