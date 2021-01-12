# go-av

Simple, idiomatic, opinionated wrapper for libav (aka ffmpeg).

## Why?

[go-ffmpeg](https://github.com/ssttevee/go-ffmpeg) is good enough for many cases, but is not sufficient when lower level control is required. 

## Goals

- Idiomatic Go/hide C specific mechanisms
- Expose low level libav APIs
- Keep clear library boundaries

## Non-Goals

- Teaching how to use libav

## Usage/Environment Setup

[`pkg-config`](https://linux.die.net/man/1/pkg-config) is used for linking, so the ffmpeg libraries and `.pc` files must be installed from your system's package manager or from source or the `PKG_CONFIG_PATH` environment variable must be set.

## Additional Notes

- The surface area of the libav API pretty large, so no effort was spent on being complete. Pull requests are more than welcome to improve API coverage.
- Tested on linux with glibc and musl using [ffmpeg 4.3.x](https://git.ffmpeg.org/gitweb/ffmpeg.git/log/refs/heads/release/4.3).
