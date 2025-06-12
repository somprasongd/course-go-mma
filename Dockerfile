FROM golang:1.24-alpine AS base
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .

FROM base AS builder
ENV GOARCH=amd64

# ตั้งค่า default สำหรับ VERSION
ARG VERSION=latest
ENV IMAGE_VERSION=${VERSION}
RUN echo "Build version: $IMAGE_VERSION"
RUN cd src/app && \
	go build -ldflags \
	"-X 'go-mma/build.Version=${IMAGE_VERSION}' \
	-X 'go-mma/build.Time=$(date +"%Y-%m-%dT%H:%M:%S%z")'" \
	-o ../../app cmd/api/main.go

FROM alpine:latest
WORKDIR /root/
EXPOSE 8090
ENV TZ=Asia/Bangkok
RUN apk --no-cache add ca-certificates tzdata

COPY --from=builder /app/app .

CMD ["./app"]