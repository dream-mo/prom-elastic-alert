FROM golang:1.18
ENV GOPROXY=https://goproxy.cn,direct
WORKDIR /app/
ENTRYPOINT ["go", "run", "/app/main.go"]
