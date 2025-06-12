module go-mma/modules/order

go 1.24.4

replace go-mma/shared/common v0.0.0 => ../../shared/common

replace go-mma/modules/notification v0.0.0 => ../../modules/notification

replace go-mma/modules/customer v0.0.0 => ../../modules/customer

replace go-mma/shared/contract/customercontract v0.0.0 => ../../shared/contract/customercontract

require (
	github.com/gofiber/fiber/v3 v3.0.0-beta.4
	go-mma/modules/notification v0.0.0
	go-mma/shared/common v0.0.0
	go-mma/shared/contract/customercontract v0.0.0
)

require (
	github.com/andybalholm/brotli v1.1.1 // indirect
	github.com/fxamacker/cbor/v2 v2.7.0 // indirect
	github.com/gofiber/schema v1.2.0 // indirect
	github.com/gofiber/utils/v2 v2.0.0-beta.7 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/jmoiron/sqlx v1.4.0 // indirect
	github.com/klauspost/compress v1.17.11 // indirect
	github.com/lib/pq v1.10.9 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/philhofer/fwd v1.1.3-0.20240916144458-20a13a1f6b7c // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/tinylib/msgp v1.2.5 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasthttp v1.58.0 // indirect
	github.com/valyala/tcplisten v1.0.0 // indirect
	github.com/x448/float16 v0.8.4 // indirect
	go.elastic.co/ecszap v1.0.3 // indirect
	go.uber.org/multierr v1.10.0 // indirect
	go.uber.org/zap v1.27.0 // indirect
	golang.org/x/crypto v0.31.0 // indirect
	golang.org/x/net v0.31.0 // indirect
	golang.org/x/sys v0.28.0 // indirect
	golang.org/x/text v0.21.0 // indirect
)
