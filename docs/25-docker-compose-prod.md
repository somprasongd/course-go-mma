## Docker Compose สำหรับ Production

ในบทนี้จะเป็นตัวอย่างการสร้าง `docker-compose.prod.yml` สำหรับ production ซึ่งมี

- Nginx: สำหรับเป็น reverse proxy
- Migrate: สำหรับการทำ database migration ก่อนรันโปรแกรม
- App: มีการ build app จาก Dockerfile

### สร้าง Nginx Config

> สร้างไฟล์ `config/nginx.conf` ที่ root project
>

```text
worker_processes auto;

events {
  worker_connections 1024;
}

http {
  # เปิด request_id module (nginx มีในตัว)
  # ไม่ต้องเพิ่ม module พิเศษ

  server {
    listen 80;

    location / {
      # ถ้า header X-Request-ID มีอยู่แล้ว ให้ใช้ของเดิม
      # ถ้าไม่มี nginx จะ generate $request_id อัตโนมัติ
      proxy_set_header X-Request-ID $request_id;

      proxy_pass http://backend;
    }
  }

  upstream backend {
    server app:8090;
  }
}
```

### สร้าง Docker Compose

> สร้างไฟล์ `docker-compose.prod.yml` ที่ root project
>

```yaml
networks:
  # network สำหรับ reverse proxy
  frontend:
  # network สำหรับ services ภายใน
  backend:

services:
  db:
    environment:
      POSTGRES_DB: go-mma-db
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres
    networks:
      - backend
    # ตรวจสอบว่าพร้อมทำงานรึยัง
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres"]
      interval: 1s
      timeout: 5s
      retries: 10

  migrate:
    image: migrate/migrate:v4.18.3
    volumes:
      - ./migrations:/migrations
    networks:
      - backend
    depends_on:
      # ต้องรันหลังจากที่ db พร้อมทำงานแล้ว
      db:
        condition: service_healthy
    # สั่ง migrate up
    command: -verbose -path=/migrations/ -database "postgres://postgres:postgres@db:5432/go-mma-db?sslmode=disable" up

  app:
    build: .
    image: go-mma-api
    container_name: go-mma-api
    environment:
      - GRACEFUL_TIMEOUT=5s
      - DB_DSN=postgres://postgres:postgres@db:5432/go-mma-db?sslmode=disable
      - GATEWAY_HOST=localhost
      - GATEWAY_BASEURL=/api/v1
    networks:
      - frontend
      - backend
    depends_on:
      db:
        # ต้องรันหลังจากที่ db พร้อมทำงานแล้ว
        condition: service_healthy 
      migrate:
        # ต้องรันหลังจากที่ migrate ทำงานสำเร็จแล้ว
        condition: service_completed_successfully

  proxy:
    image: nginx:1.29.0-alpine
    container_name: nginx-proxy
    ports:
      - "80:80"
    volumes:
      # เรียกใช้ config/nginx.conf
      - ./config/nginx.conf:/etc/nginx/nginx.conf:ro
    networks:
      - frontend
    depends_on:
      app:
        # ต้องรันหลังจากที่ app เริ่มทำงานแล้ว
        condition: service_started
```

### Run Production

สามารถรันได้จากคำสั่ง `docker compose -f docker-compose.yml -f docker-compose.prod.yml up -d` แต่เราจะนำคำสั่งนี้มาใส่ใน `Makefile` แทน

> แก้ไข `Makefile` เพิ่ม
>

```makefile
.PHONY: produp
produp:
 docker compose -f docker-compose.yml -f docker-compose.prod.yml up -d

.PHONY: proddown
proddown:
 docker compose -f docker-compose.yml -f docker-compose.prod.yml down
```

รันคำสั่ง: `make produp`

เรียบร้อยพร้อมทำงาน ถ้าจะหยุดการทำงานก็รันคำสั่ง: `make proddown`

---

