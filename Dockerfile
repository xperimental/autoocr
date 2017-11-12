FROM ubuntu:17.10
LABEL maintainer="Robert Jacob <xperimental@solidproject.de>"

ENV DEBIAN_FRONTEND noninteractive

RUN apt-get update \
 && apt-get install -y pdfsandwich tesseract-ocr-all \
 && apt-get clean \
 && rm -r /var/lib/apt/lists/* \
 && rm /var/cache/apt/*
