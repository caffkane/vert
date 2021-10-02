# syntax=docker/dockerfile:1

FROM debian

RUN apt-get update && \
    apt-get install -y \
    wget
RUN mkdir /tmp/deepspech/lib && mkdir /tmp/deepspech/include
RUN wget \
    https://github.com/mozilla/DeepSpeech/releases/download/v0.9.0/native_client.amd64.cpu.linux.tar.xz \
    && tar -xf native_client.amd64.cpu.linux.tar.xz -C /tmp/deepspeech/lib \
RUN wget \
    https://github.com/mozilla/DeepSpeech/raw/v0.9.0/native_client/deepspeech.h \
    && cp deepspeech.h /tmp/deepspeech/include

ENV CGO_LDFLAGS="-L/tmp/deepspeech/lib/"
ENV CGO_CXXFLAGS="-I/tmp/deepspeech/include/"
ENV LD_LIBRARY_PATH=/tmp/deepspeech/lib/:$LD_LIBRARY_PATH

RUN go get -u github.com/asticode/go-astideepspeech/...



FROM golang:1.17

WORKDIR /app

COPY go.mod ./
#COPY go.sum ./
RUN go mod download

COPY *.go ./

RUN go build -o /app

EXPOSE 8080

CMD [ "/app" ]
