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
