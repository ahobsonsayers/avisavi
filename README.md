# avisavi

<p align="center">
  <img src="assets/logo.png" alt="avisavi logo">
</p>

A CLI for searching Avios reward flights from the terminal.

Search reward flight availability, browse routes across cabin classes, and check your balance.

It uses the same API as the official Android app, built via decompiling and reverse engineering its code. Decompilation scripts and learnings can be found in the [`avios-decompile`](https://github.com/ahobsonsayers/avios-decompile) repo.

## Features

- **Availability** - seat counts per cabin (Economy, Premium, Business, First) for a route.
- **Routes** - list every reward destination from an origin with Economy/Business price ranges.
- **Search** - scan all destinations for seats on your exact outbound and return dates.
- **Balance** - check your Avios balance, including household accounts.

### Why not avios-cli

There is a project called [avios-cli](https://github.com/alexechoi/avios-cli) that does a very similar thing to this project.

It's a good project, but this CLI has several advantages, mainly:

- This CLI does not require Python - it is written in Go and compiled to a single static binary, meaning you can simply download it and run it anywhere 🏃
- This CLI uses the API used by the Avios app - `avios-cli` uses Playwright, which is slow and inefficient as well as very fragile. This CLI relies on the official API used by the Avios app, so it isn't hit by any of those drawbacks 🏃

## Install

Download a prebuilt binary from the [latest release](https://github.com/ahobsonsayers/avisavi/releases/latest) - binaries are available for macOS, Linux, and Windows.

## Usage

> [!WARNING]
> See the [disclaimer](#%EF%B8%8F-disclaimer) before going further

Log in first - the browser flow saves your session to `~/.config/avisavi/auth.json` (macOS/Linux) or `%AppData%\avisavi\auth.json` (Windows):

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

## Build

Build from source:

```sh
go build ./cmd/avisavi
```

### Configuration

Prebuilt binaries from the releases page work out of the box - no configuration needed.

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
> The Avios client ID is not provided here due to liability concerns, but it's a fixed, unchanging value. You can extract it from the network tab of the official Avios site or app.

## ⚠️ Disclaimer

This tool is unofficial and not affiliated with, endorsed by, or supported by Avios.

Using it may violate Avios' terms of service - your account could be suspended or banned, and any points in it could be lost.

I take no responsibility for any consequences of using this software. Please use a dedicated burner account, not your primary one.
