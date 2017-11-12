FROM golang:1.9.2-alpine AS builder

ENV CGO_ENABLED=0

COPY . /go/src/github.com/xperimental/autoocr/
WORKDIR /go/src/github.com/xperimental/autoocr/
RUN go install .

FROM ubuntu:17.10
LABEL maintainer="Robert Jacob <xperimental@solidproject.de>"

ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update \
 && apt-get install -y pdfsandwich tesseract-ocr-all \
 && apt-get clean \
 && rm -r /var/lib/apt/lists/* \
 && rm -r /var/cache/apt/*

RUN mkdir -p /data/input /data/output
WORKDIR /data/output

VOLUME [ "/data/input", "/data/output" ]

COPY --from=builder /go/bin/autoocr /bin/autoocr

ENTRYPOINT [ "/bin/autoocr" ]
CMD [ ]
