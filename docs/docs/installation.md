---
sidebar_position: 1
title: Installation
description: How to install Sindr on your system
keywords: [sindr, installation, install, setup]
---

# Installation

Get Sindr up and running on your system with one of the following methods.

## Using the install script

The easiest way to install Sindr is using our install script:

```bash
curl -sSL https://raw.githubusercontent.com/your-repo/sindr/main/install.sh | bash
```

This will download the latest release and install it to your system.

## Building from source

If you prefer to build from source or need the latest development version:

```bash
git clone https://github.com/your-repo/sindr.git
cd sindr
go build -o sindr cmd/main.go
```

Make sure you have Go 1.19+ installed on your system.

## Package Managers

### Homebrew (macOS/Linux)

```bash
brew install sindr
```

### Manual Installation

1. Download the latest release for your platform from [GitHub Releases](https://github.com/your-repo/sindr/releases)
2. Extract the archive
3. Move the `sindr` binary to a directory in your `PATH` (e.g., `/usr/local/bin`)

## Verification

After installation, verify it works:

```bash
sindr --version
```

You should see version information printed to the terminal.

## Next Steps

Now that Sindr is installed, check out the [Getting Started guide](./getting-started.md) to create your first Sindr project.