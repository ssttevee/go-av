# go-av

Simple, idiomatic, opinionated wrapper for libav (aka ffmpeg).

## Why?

[go-ffmpeg](https://github.com/ssttevee/go-ffmpeg) is good enough for many cases, but is not sufficient when lower level control is required. 

## Features

- Go-native data structures
- Go-native logging (with `*log.Logger`)
- Opinionated but light library for common libav use cases
- Direct low level libav function access (via avcodec, avformat, etc. sub packages)
- Option for static and dynamic binding

## Goals

- Idiomatic Go/hide C specific mechanisms
- Expose low level libav APIs
- Keep clear library boundaries

## Non-Goals

- Teaching how to use libav
- Bundling and building libav

## Usage/Environment Setup

[`pkg-config`](https://linux.die.net/man/1/pkg-config) is used for linking, so the ffmpeg libraries and `.pc` files must be installed from your system's package manager or from source or the `PKG_CONFIG_PATH` environment variable must be set.

## Gotchas

[Cgo is very particular with pointers.](https://golang.org/pkg/cmd/cgo/#hdr-Passing_pointers) The Go runtime will actively prevent Go pointers from reaching C code by panicking if tried. However, some libav functions (like `*_free_context`) require a pointer to a pointer for output.

Fortunately, discovered only by experimentation, stack pointers are able to pass through this boundary with no problem. This means that pointers to local variables are safe to use for these functions. Whether this is intended behaviour is beyond me.

## Additional Notes

- The surface area of the libav API pretty large, so no effort was spent on being complete. Pull requests are more than welcome to improve API coverage.
- Tested on linux with glibc and musl using [ffmpeg 4.3.x](https://git.ffmpeg.org/gitweb/ffmpeg.git/log/refs/heads/release/4.3).
