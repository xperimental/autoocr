# autoocr

[![Docker Build Status](https://img.shields.io/docker/build/xperimental/autoocr.svg?style=flat-square)](https://store.docker.com/community/images/xperimental/autoocr)

`autoocr` is a small tool which watches a directory for PDF files and then uses [`pdfsandwich`](http://www.tobias-elze.de/pdfsandwich/) to do OCR on them. After processing the files are moved to an output location.

## Usage

You can build the tool yourself if you have Go installed. Generally using the provided [docker image](https://hub.docker.com/r/xperimental/autoocr/) is much easier though.

The docker image has two pre-defined volumes:

- `/data/input`
- `/data/output`

The input directory is watched by the process for new files and the resulting files are written to the output directory. Running the image should generally be as easy as:

```bash
docker run --name autoocr \
  -v /path/to/input:/data/input \
  -v /path/to/output:/data/output \
  xperimental/autoocr:latest
```

The `autoocr` executable has a few options that can also be passed to the container:

```plain
Usage of autoocr:
      --delay duration        Processing delay after receiving watch events. (default 5s)
  -i, --input string          Directory to use for input. (default "input")
      --keep-original         Keep backup of original file. (default true)
      --languages string      OCR Languages to use. (default "deu+eng")
      --log-format string     Logging format to use. (default "plain")
      --log-level string      Logging level to show. (default "info")
  -o, --output string         Directory to use for output. (default "output")
      --pdf-sandwich string   Path to pdfsandwich utility. (default "pdfsandwich")
```

## Acknowledgements

This software would not be possible without the existence of `pdfsandwich`, `tesseract`, `ImageMagick` and others. Thanks for providing those!
