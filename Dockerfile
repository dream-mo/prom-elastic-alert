FROM --platform=$BUILDPLATFORM golang:1.18 as build

ARG GOPROXY
ARG GOSUMDB
ARG GOPRIVATE
ARG TARGETARCH

WORKDIR /app

ENV GO111MODULE=on \
    GOPROXY=https://goproxy.cn,direct

COPY . .

RUN make elastic-alert-linux

FROM docker.m.daocloud.io/alpine:3.15

RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories

RUN apk add --no-cache ca-certificates tzdata

COPY --from=build /app/elastic-alert  /bin/elastic-alert
COPY --from=build /app/config.yaml    /bin/config.yaml

EXPOSE 8000 9000

ENTRYPOINT ["/bin/elastic-alert"]
