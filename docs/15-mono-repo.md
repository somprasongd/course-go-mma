## รวมโค้ดทั้งหมดไว้ใน Mono-Repository อย่างเป็นระบบ

เพื่อให้การพัฒนาเป็นระบบมากขึ้น และลดความซับซ้อนเมื่อโปรเจกต์เติบโต เราแนะนำให้แยกแต่ละโมดูลหลัก เช่น `customer`, `order`, และ `notification` ออกเป็น **โปรเจกต์ย่อย (submodule)** โดยแต่ละโปรเจกต์จะมี `go.mod` ของตัวเอง

แต่ยังคงเก็บทุกอย่างไว้ใน Git repository เดียวกัน เรียกว่า **Mono-Repository**

**โดยจะแบ่งโค้ดออกเป็น 3 ส่วน หลักๆ คือ**

- **app:** สำหรับโหลดโมดูล และรันโปรแกรม
- **modules:** สำหรับสร้างโมดูลต่างๆ
- **shared:** สำหรับโค้ดที่ใช้งานร่วมกัน

### โครงสร้างใหม่

```bash
.
├── docker-compose.dev.yml
├── docker-compose.yml
├── go-mma.code-workspace
├── Makefile
├── migrations
│   ├── 20250529103238_create_customer.down.sql
│   ├── 20250529103238_create_customer.up.sql
│   ├── 20250529103715_create_order.down.sql
│   └── 20250529103715_create_order.up.sql
└── src
    ├── app
    │   ├── application
    │   │   ├── application.go
    │   │   ├── http.go
    │   │   └── middleware.go
    │   │   │   ├── request_logger.go
    │   │   │   └── response_error.go
    │   ├── cmd
    │   │   └── api
    │   │       └── main.go
    │   ├── config
    │   │   └── config.go
    │   ├── go.mod
    │   ├── go.sum
    │   └── util
    │       └── env
    │           └── env.go
    ├── modules
    │   ├── customers
    │   │   ├── dtos
    │   │   │   ├── customer_request.go
    │   │   │   ├── customer_response.go
    │   │   │   └── customer.go
    │   │   ├── handler
    │   │   │   └── customer.go
    │   │   ├── internal
    │   │   │   ├── model
    │   │   │   │   └── customer.go
    │   │   │   └── repository
    │   │   │       └── customer.go
    │   │   ├── module.go
    │   │   ├── service
    │   │   │   └── customer.go
    │   │   ├── test
    │   │   │   └── customer.http
    │   │   ├── go.mod
    │   │   └── go.sum
    │   ├── notifications
    │   │   ├── module.go
    │   │   ├── service
    │   │   │   └── notification.go
    │   │   ├── go.mod
    │   │   └── go.sum
    │   └── orders
    │   │   ├── dtos
    │   │   │   ├── order_request.go
    │   │   │   └── order_response.go
    │   │   ├── handler
    │   │   │   └── order.go
    │   │   ├── internal
    │   │   │   ├── model
    │   │   │   │   └── order.go
    │   │   │   └── repository
    │   │   │       └── order.go
    │   │   ├── module.go
    │   │   ├── service
    │   │   │   └── order.go
    │   │   ├── test
    │   │   │   └── order.http
    │   │   ├── go.mod
    │   │   └── go.sum
    └── shared
        └──common
            ├── errs
            │   ├── errs.go
            │   ├── helper.go
            │   └── types.go
            ├── logger
            │   └── logger.go
            ├── module
            │   └── module.go
            ├── registry
            │   ├── helper.go
            │   └── service_registry.go
            ├── storage
            │   └── db
            │       ├── db.go
            │       └── transactor
            │           ├── nested_transactions_none.go
            │           ├── nested_transactions_savepoints.go
            │           ├── transactor.go
            │           └── types.go
            ├── go.mod
            └── go.sum
```

**ข้อดีของการจัดแบบนี้**

- **แยกขอบเขตชัดเจน**: แต่ละโมดูลพัฒนาและทดสอบได้แบบอิสระ
- **ควบคุมเวอร์ชันได้ง่าย**: ใช้ `go.mod` จัดการ dependency ภายในแต่ละโมดูล
- **รวมศูนย์การจัดการ**: ยังใช้ Git ร่วมกันใน repo เดียว ไม่ต้องแยกหลาย repo
- **พร้อมสำหรับการแยกเป็น microservice ในอนาคต**: โครงสร้างรองรับการแยก deploy ได้หากจำเป็น

### สร้างโปรเจกต์ใหม่

- สร้าง Folder ใหม่ ดังนี้

    ```bash
    # อยู่ที่ root project
    mkdir -p src/app
    mkdir -p src/modules/customer
    mkdir -p src/modules/notification
    mkdir -p src/modules/order
    mkdir -p src/shared/common
    ```

- สร้าง app โปรเจกต์

    ```bash
    # อยู่ที่ root project
    cd src/app
    go mod init go-mma
    ```

- สร้าง customer โปรเจกต์

    ```bash
    cd ../..
    # อยู่ที่ root project
    cd src/modules/customer
    go mod init go-mma/modules/customer
    ```

- สร้าง notification โปรเจกต์

    ```bash
    cd ../../..
    # อยู่ที่ root project
    cd src/modules/notification
    go mod init go-mma/modules/notification
    ```

- สร้าง order โปรเจกต์

    ```bash
    cd ../../..
    # อยู่ที่ root project
    cd src/modules/order
    go mod init go-mma/modules/order
    ```

- สร้าง common โปรเจค

    ```bash
    cd ../../..
    # อยู่ที่ root project
    cd src/shared/common
    go mod init go-mma/shared/common
    ```

### ทำ L**ocal module replacement**

เมื่อพัฒนาแบบ Monorepo เพื่อให้แต่ละโมดูลสามารถอ้างถึงกันได้โดยตรงจากไฟล์ในเครื่อง โดยไม่ต้อง publish ไปที่ remote repo ใน Go ทำได้โดย การใช้คำสั่ง `replace` ใน `go.mod`

- โปรเจกต์ notification มีการใช้งาน common

    > แก้ไขไฟล์ `notification/go.mod`
    >

    ```bash
    module go-mma/modules/notification
    
    go 1.24.1
    
    replace go-mma/shared/common v0.0.0 => ../../shared/common
    ```

- โปรเจกต์ customer มีการใช้งาน common, notification

    > แก้ไขไฟล์ `customer/go.mod`
    >

    ```bash
    module go-mma/modules/customer
    
    go 1.24.1
    
    replace go-mma/shared/common v0.0.0 => ../../shared/common
    
    replace go-mma/modules/notification v0.0.0 => ../../modules/notification
    ```

- โปรเจกต์ order มีการใช้งาน common, notification, customer

    > แก้ไขไฟล์ `order/go.mod`
    >

    ```bash
    module go-mma/modules/order
    
    go 1.24.1
    
    replace go-mma/shared/common v0.0.0 => ../../shared/common
    
    replace go-mma/modules/notification v0.0.0 => ../../modules/notification
    
    replace go-mma/modules/customer v0.0.0 => ../../modules/customer
    ```

- โปรเจกต์ app มีการใช้งาน common, notification, customer, order

    แก้ไขไฟล์ `app/go.mod`

    ```bash
    module go-mma
    
    go 1.24.1
    
    replace go-mma/shared/common v0.0.0 => ../shared/common
    
    replace go-mma/modules/notification v0.0.0 => ../modules/notification
    
    replace go-mma/modules/customer v0.0.0 => ../modules/customer
    
    replace go-mma/modules/order v0.0.0 => ../modules/order
    ```

### สร้าง VS Code Workspace

สำหรับการทำ Mono-Repo ใน VS Code ต้องเปิดแบบ Workspace ถึงจะสามารถทำงานได้ถูกต้อง

- สร้างไฟล์ `go-mma.code-workspace`

    ```bash
    {
     "folders": [
      {
       "path": "."
      },
      {
       "path": "src/app"
      },
      {
       "path": "src/modules/customer"
      },
      {
       "path": "src/modules/order"
      },
      {
       "path": "src/modules/notification"
      },
      {
       "path": "src/shared/common"
      }
     ],
     "settings": {}
    }
    ```

- เลือกที่เมนู File เลือก Open Workspace from file…
- เลือกที่ไฟล์ `go-mma.code-workspace`
- กด Open
- ใน Explorer จะแสดง แบบนี้

    ```bash
    go-mma
    app
    customer
    order
    notification
    common
    ```

### โปรเจกต์ common

- ให้ทำการย้ายโค้ดใน `util` ทั้งหมด ยกเว้น `env` มาไว้ในโปรเจกต์ `common`

    ```bash
    common
    ├── go.mod
    ├── errs
    │   ├── errs.go
    │   ├── helpers.go
    │   └── types.go
    ├── idgen
    │   └── idgen.go
    ├── logger
    │   └── logger.go
    ├── module
    │   └── module.go
    ├── registry
    │   ├── helper.go
    │   └── service_registry.go
    └── storage
        └── sqldb
            ├── sqldb.go
            └── transactor
                ├── nested_transactions_none.go
                ├── nested_transactions_savepoints.go
                ├── transactor.go
                └── types.go
    ```

- ติดตั้ง dependencies ด้วย `go mod tidy`

### โปรเจกต์ notification

- ให้ทำการย้ายโค้ดใน `modules/notification` ทั้งหมด  มาไว้ในโปรเจกต์ `notification`

    ```bash
    notification
    ├── go.mod
    ├── go.sum
    ├── module.go
    └── service
        └── notification.go
    ```

- แก้ไข path ของการ `import` ดังนี้
  - `go-mma/util/logger` → `go-mma/shared/common/logger`
- ติดตั้ง dependencies ด้วย `go mod tidy`

### โปรเจกต์ customer

- ให้ทำการย้ายโค้ดใน `modules/customer` ทั้งหมด  มาไว้ในโปรเจกต์ `customer`

    ```bash
    customer
    ├── dto
    │   ├── customer_request.go
    │   ├── customer_response.go
    │   └── customer.go
    ├── go.mod
    ├── go.sum
    ├── handler
    │   └── customer.go
    ├── internal
    │   ├── model
    │   │   └── customer.go
    │   └── repository
    │       └── customer.go
    ├── module.go
    ├── service
    │   └── customer.go
    └── test
        └── customers.http
    ```

- แก้ไข path ของการ `import` ดังนี้
  - `go-mma/util/errs` → `go-mma/shared/common/errs`
  - `go-mma/util/logger` → `go-mma/shared/common/logger`
  - `go-mma/util/module` → `go-mma/shared/common/module`
  - `go-mma/util/registry` → `go-mma/shared/common/registry`
  - `go-mma/util/storage/sqldb/transactor` → `go-mma/shared/common/storage/sqldb/transactor`
- ติดตั้ง dependencies ด้วย `go mod tidy`

### โปรเจกต์ order

- ให้ทำการย้ายโค้ดใน `modules/order` ทั้งหมด  มาไว้ในโปรเจกต์ `order`

    ```bash
    order
    ├── dto
    │   ├── customer_request.go
    │   ├── customer_response.go
    │   └── customer.go
    ├── go.mod
    ├── go.sum
    ├── handler
    │   └── customer.go
    ├── internal
    │   ├── model
    │   │   └── customer.go
    │   └── repository
    │       └── customer.go
    ├── module.go
    ├── service
    │   └── customer.go
    └── test
        └── customers.http
    ```

- แก้ไข path ของการ `import` ดังนี้
  - `go-mma/util/errs` → `go-mma/shared/common/errs`
  - `go-mma/util/logger` → `go-mma/shared/common/logger`
  - `go-mma/util/module` → `go-mma/shared/common/module`
  - `go-mma/util/registry` → `go-mma/shared/common/registry`
  - `go-mma/util/storage/sqldb/transactor` → `go-mma/shared/common/storage/sqldb/transactor`
- ติดตั้ง dependencies ด้วย `go mod tidy`

### โปรเจกต์ app

- ให้ทำการย้ายโค้ดใน `application`, `cmd`, `config` และ `util`   มาไว้ในโปรเจกต์ `app`

    ```bash
    app
    ├── application
    │   ├── application.go
    │   ├── http.go
    │   └── middleware
    │       ├── request_logger.go
    │       └── response_error.go
    ├── cmd
    │   └── api
    │       └── main.go
    ├── config
    │   └── config.go
    ├── go.mod
    ├── go.sum
    └── util
        └── env
            └── env.go
    ```

- แก้ไข path ของการ `import` ดังนี้
  - `go-mma/util/errs` → `go-mma/shared/common/errs`
  - `go-mma/util/logger` → `go-mma/shared/common/logger`
  - `go-mma/util/module` → `go-mma/shared/common/module`
  - `go-mma/util/registry` → `go-mma/shared/common/registry`
  - `go-mma/util/storage/sqldb` → `go-mma/shared/common/storage/sqldb`
  - `go-mma/util/storage/sqldb/transactor` → `go-mma/shared/common/storage/sqldb/transactor`
- ติดตั้ง dependencies ด้วย `go mod tidy`

### รันโปรแกรม

- แก้ไฟล์ `Makefile` เพื่อแก้ path ในการรัน

    > แก้ไขไฟล์ `main.go`
    >

    ```bash
    .PHONY: run
    run:
     cd src/app && \
     go run cmd/api/main.go
    ```

- ทดลองรันโปรแกรม

    ```bash
    make run
    ```

### Build โปรแกรม

- แก้ไฟล์ `Makefile` เพื่อแก้ path ในการ build

    > แก้ไขไฟล์ `main.go`

    ```bash
    .PHONY: build
    build:
        cd src/app && \
        go build -ldflags \
        "-s -w \
        -X 'go-mma/build.Version=${BUILD_VERSION}' \
        -X 'go-mma/build.Time=${BUILD_TIME}'" \
        -o ../../app cmd/api/main.go
    ```

- ทดลองรัน build

    ```bash
    BUILD_VERSION=0.0.2 make build
    ```

### Build โปรแกรมเป็น Docker image

- แก้ไฟล์ `Dockerfile` เพื่อแก้ path ในการ build

    > แก้ไขไฟล์ `Dockerfile`

    ```docker
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
    ```

- ทดลองรัน build

    ```bash
    BUILD_VERSION=0.0.2 make image
    ```

