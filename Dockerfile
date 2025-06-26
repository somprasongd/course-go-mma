# Builder stage
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY . .
RUN cd src/app && go mod download

# ปิด cgo (CGO_ENABLED=0)
ENV CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

# ตั้งค่า default สำหรับ VERSION
ARG VERSION=latest
ENV IMAGE_VERSION=${VERSION}
RUN echo "Build version: $IMAGE_VERSION"
RUN cd src/app && \
	go build -ldflags \
	# strip debugging information ออกจาก binary ทำให้ไฟล์เล็กลง
	"-s -w \
	-X 'go-mma/build.Version=${IMAGE_VERSION}' \
	-X 'go-mma/build.Time=$(date +"%Y-%m-%dT%H:%M:%S%z")'" \
	-o /app/app cmd/api/main.go

# Final stage
FROM alpine:3.22
RUN apk add --no-cache ca-certificates tzdata && \
		# สร้าง non-root user
    addgroup -S appgroup && adduser -S appuser -G appgroup 
	
WORKDIR /app
COPY --from=builder /app/app .
USER appuser
EXPOSE 8090
ENV TZ=Asia/Bangkok

# ใช้ ENTRYPOINT เพื่อล็อกว่าต้องรันไฟล์ไหน
ENTRYPOINT ["./app"]