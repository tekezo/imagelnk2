# ImageLnk2

Extract the URL of images from a given URL.

## Preparation

Run the following commands on ubuntu 22.04.

```shell
wget https://dl.google.com/linux/direct/google-chrome-stable_current_amd64.deb
apt install ./google-chrome-stable_current_amd64.deb

sudo apt install libvips-dev
```

## Setup

```shell
cp imagelnk2.toml.example imagelnk2.toml
```

## Run

```shell
make
./imagelnk2
```

## Usage

```shell
curl 'http://localhost:8930/get?url=https://github.com/tekezo/'
```

```json
{
    "title": "tekezo - Overview",
    "imageURLs": ["https://avatars.githubusercontent.com/u/659178?v=4?s=400"],
    "extraURLs": [],
    "imageCacheURLs": [
        "https://example.com/original/80/649590e0-80f52c30f8a4263.jpeg"
    ]
}
```

### CLI

Build the command-line client:

```shell
make build
```

Pass a URL to fetch and print the result as JSON:

```shell
./imagelnk2-cli 'https://x.com/ImageLnk/status/1296296267526500353'
```

The CLI reads `imagelnk2.toml` from the current directory and launches Chrome
directly, so the HTTP server does not need to be running.
