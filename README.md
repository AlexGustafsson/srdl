# srdl

srdl is a spiritual sibling to [ytdl](https://github.com/ytdl-org/youtube-dl) as
well as [ytdl-sub](https://github.com/jmbannon/ytdl-sub), but for Sveriges Radio
(SR). That is, srdl allows you to easily archive programs from SR.

## Features

- Focused on performance
  - Zero runtime dependencies like ffmpeg
  - Virtually zero RAM or CPU usage
- Metadata is written into downloaded audio files, compatible with Jellyfin and
  others
- Downloads include cover, backdrop and episode images
- Throttling configuration for fair bandwidth

## Getting started (srdl)

### Running srdl on host

```shell
srdl program "https://sverigesradio.se/textochmusikmedericschuldt"
```

```json
{
  "id": 4914,
  "name": "Text och musik med Eric Schüldt",
  "description": "En timme med den vackraste musiken ackompanjerad av poesi, filosofi och personliga reflektioner.",
  "programcategory": {
    "id": 5,
    "name": "Musik"
  },
  "broadcastinfo": "Söndag 11.00",
  "email": "textochmusik@sverigesradio.se",
  "phone": "",
  "programurl": "https://sverigesradio.se/default.aspx?programid=4914",
  "programslug": "textochmusikmedericschuldt",
  "programimage": "https://static-cdn.sr.se/images/4914/dd5ffd1e-5548-4f2e-87ea-0ab681a23855.jpg?preset=api-default-square",
  "programimagetemplate": "https://static-cdn.sr.se/images/4914/dd5ffd1e-5548-4f2e-87ea-0ab681a23855.jpg",
  "programimagewide": "https://static-cdn.sr.se/images/4914/74ebbeb2-9948-499b-9bc9-94cffd2d456a.jpg?preset=api-default-rectangle",
  "programimagetemplatewide": "https://static-cdn.sr.se/images/4914/74ebbeb2-9948-499b-9bc9-94cffd2d456a.jpg",
  "socialimage": "https://static-cdn.sr.se/images/4914/dd5ffd1e-5548-4f2e-87ea-0ab681a23855.jpg?preset=api-default-square",
  "socialimagetemplate": "https://static-cdn.sr.se/images/4914/dd5ffd1e-5548-4f2e-87ea-0ab681a23855.jpg",
  "socialmediaplatforms": [
    {
      "platform": "Facebook",
      "platformurl": "https://facebook.com/sverigesradioP2"
    },
    {
      "platform": "Twitter",
      "platformurl": "https://twitter.com/sverigesradioP2/"
    },
    {
      "platform": "Instagram",
      "platformurl": "https://instagram.com/sverigesradio_p2/"
    }
  ],
  "channel": {
    "id": 163,
    "name": "P2"
  },
  "archived": false,
  "hasondemand": true,
  "haspod": false,
  "responsibleeditor": "Pia Kalischer"
}
```

### Running srdl using docker

```shell
docker run --rm \
  --volume "$PWD/output:/output" \
  --entrypoint srdl \
  ghcr.com/alexgustafsson/srdl:latest \
    program \
    "https://sverigesradio.se/textochmusikmedericschuldt"
```

```shell
docker run --rm \
  --volume "$PWD/output:/output" \
  --entrypoint srdl \
  ghcr.com/alexgustafsson/srdl:latest \
    download \
    --output /output/out.m4a \
    --episode-id 1234
```

## Getting started (srdl-sub)

### Running srdl-sub on host

```shell
srdl-sub \
  --config config/config.yaml \
  --subscriptions config/subscriptions.yaml
```

### Running srdl-sub using docker

```shell
docker run --rm \
  --volume "$PWD/output:/output" \
  --volume "$PWD/examples:/config" \
  ghcr.com/alexgustafsson/srdl:latest \
    --config config/config.yaml \
    --subscriptions config/subscriptions.yaml
```

### Output

```text
output
└── Erik Schüldt
    └── Text och musik
        ├── Bland blåskummande blommor.m4a
        ├── Bland blåskummande blommor.png
        ├── Detta är skönheten.jpg
        ├── Detta är skönheten.m4a
        ├── En romantikers bekännelse.jpg
        ├── En romantikers bekännelse.m4a
        ├── Ensamheten.m4a
        ├── Ensamheten.png
        ├── backdrop.jpg
        └── cover.jpg
```

```text
ffprobe version 7.1 Copyright (c) 2007-2024 the FFmpeg developers
  built with Apple clang version 16.0.0 (clang-1600.0.26.3)
  configuration: --prefix=/opt/homebrew/Cellar/ffmpeg/7.1 --enable-shared --enable-pthreads --enable-version3 --cc=clang --host-cflags= --host-ldflags='-Wl,-ld_classic' --enable-ffplay --enable-gnutls --enable-gpl --enable-libaom --enable-libaribb24 --enable-libbluray --enable-libdav1d --enable-libharfbuzz --enable-libjxl --enable-libmp3lame --enable-libopus --enable-librav1e --enable-librist --enable-librubberband --enable-libsnappy --enable-libsrt --enable-libssh --enable-libsvtav1 --enable-libtesseract --enable-libtheora --enable-libvidstab --enable-libvmaf --enable-libvorbis --enable-libvpx --enable-libwebp --enable-libx264 --enable-libx265 --enable-libxml2 --enable-libxvid --enable-lzma --enable-libfontconfig --enable-libfreetype --enable-frei0r --enable-libass --enable-libopencore-amrnb --enable-libopencore-amrwb --enable-libopenjpeg --enable-libspeex --enable-libsoxr --enable-libzmq --enable-libzimg --disable-libjack --disable-indev=jack --enable-videotoolbox --enable-audiotoolbox --enable-neon
  libavutil      59. 39.100 / 59. 39.100
  libavcodec     61. 19.100 / 61. 19.100
  libavformat    61.  7.100 / 61.  7.100
  libavdevice    61.  3.100 / 61.  3.100
  libavfilter    10.  4.100 / 10.  4.100
  libswscale      8.  3.100 /  8.  3.100
  libswresample   5.  3.100 /  5.  3.100
  libpostproc    58.  3.100 / 58.  3.100
[aac @ 0x11df06150] Prediction is not allowed in AAC-LC.
Input #0, mov,mp4,m4a,3gp,3g2,mj2, from 'output/Erik Schüldt/Text och musik/Detta är skönheten.m4a':
  Metadata:
    major_brand     : M4A
    minor_version   : 512
    compatible_brands: isomiso2
    title           : Detta är skönheten
    album           : Text och musik med Eric Schüldt
    description     : Vi söker efter den stora skönheten. Från medeltiden till en omvälvande tolkning av Mozarts operor. Det är facklan som ska lysa i bergen där luften är välsignelse, på tundran där himlen är melankoli. Det är claritas – klarhet och ljus med gud
    date            : 2024-10-13T09:00:00Z
  Duration: 00:59:00.00, start: 0.000000, bitrate: 96 kb/s
  Stream #0:0[0x1](und): Audio: aac (LC) (mp4a / 0x6134706D), 48000 Hz, stereo, fltp, 96 kb/s (default)
      Metadata:
        handler_name    : SoundHandler
        vendor_id       : [0][0][0][0]
```

## Building

Either build using go, or docker.

```shell
go build -o srdl cmd/srdl/*.go
go build -o srdl-sub cmd/srdl-sub/*.go
```

```shell
# Build for running the container
docker build -t ghcr.com/alexgustafsson/srdl:latest .

# Build inside the container, for running on host
DOCKER_BUILDKIT=1 docker build --target=export . --output .
```
