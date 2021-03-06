FROM golang:1.14.3-alpine AS builder

WORKDIR /build

COPY go.mod go.sum /build/

RUN go mod download
RUN go mod verify

COPY . /build/

ENV CGO_ENABLED=0

RUN go build -tags netgo -ldflags "-w" .

FROM ubuntu:20.04
LABEL maintainer="Robert Jacob <xperimental@solidproject.de>"

ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update \
 && apt-get install -y pdfsandwich tesseract-ocr-all \
 && apt-get clean \
 && rm -r /var/lib/apt/lists/* \
 && rm -r /var/cache/apt/*

COPY _contrib/policy.xml /etc/ImageMagick-6/policy.xml

RUN mkdir -p /data/input /data/output
WORKDIR /data/output

VOLUME [ "/data/input", "/data/output" ]

COPY --from=builder /build/autoocr /bin/autoocr

ENTRYPOINT [ "/bin/autoocr" ]
CMD [ ]
