FROM golang:alpine as builder
LABEL maintainer="<brooksyang@outlook.com>"

# Install system dependencies including 'make'
RUN apk update && apk add --no-cache bash make protobuf-dev

WORKDIR /workspace

ADD go.mod .
ADD go.sum .

RUN go mod download

COPY . .

RUN make ca & make protoc && make build

FROM alpine:latest

ARG app_name=bingo
WORKDIR /opt/${app_name}

COPY --from=builder /workspace/_output/cert /opt/${app_name}/cert
COPY --from=builder /workspace/_output/bin/${app_name}-apiserver /opt/${app_name}/bin/

# configs
COPY build/docker /opt/${app_name}/

EXPOSE 8080

ENTRYPOINT ["/opt/bingo/bin/bingo-apiserver"]
