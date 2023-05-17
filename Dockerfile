FROM --platform=$BUILDPLATFORM golang:1.18 as build

ARG GOPROXY
ARG TARGETARCH

WORKDIR /app

ENV GO111MODULE=on
#ENV GOPROXY=https://goproxy.cn,direct

COPY . .

RUN make elastic-alert-linux

FROM scratch

ARG USER_UID=10001
USER ${USER_UID}

COPY --from=build /app/elastic-alert  /bin/elastic-alert
COPY --from=build /app/config.yaml    /etc/config.yaml
COPY --from=build /usr/share/zoneinfo /usr/share/zoneinfo

ENTRYPOINT ["/bin/elastic-alert"]
CMD ["--config", "/etc/config.yaml"]
EXPOSE 9003
