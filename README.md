# Cinema Booking API

Backend web đặt vé rạp chiếu phim viết bằng Go, Echo v5, GORM v2 và PostgreSQL 15. Code tổ chức theo luồng `handler → service → repository`, có Swagger UI tại `/swaggers`.

## Yêu cầu

- Go 1.26+
- Docker và Docker Compose (PostgreSQL local)

## Cấu trúc thư mục

```text
cmd/
├── app/                 # HTTP entry point
└── migrate/             # CLI chạy migration up/down
config/                  # Load .env / biến môi trường
api/swagger/             # Swagger spec sinh tự động — không sửa tay
migrations/              # gormigrate, mỗi migration là raw SQL trong file Go (chưa có migration nào)
internal/
├── dto/                 # Request/response DTO (json + validate + example tags)
├── handlers/            # Echo handlers, routes, HTTP error handler
├── repositories/        # Truy vấn GORM
├── services/            # Business logic
└── utils/               # Error response thống nhất, validator
pkg/
├── db/                  # Kết nối GORM/pgx, pool
└── logger/              # Khởi tạo slog
```

## Chạy local

```bash
cp .env.example .env                       # mặc định khớp docker-compose
docker compose up -d cinema-postgres       # PostgreSQL 15 tại localhost:5434
go run ./cmd/migrate -direction=up         # chạy migration (danh sách hiện rỗng → báo "No migration defined", sẽ có từ S3)
go run ./cmd/app                           # API tại http://localhost:8080
```

Kiểm tra: `curl localhost:8080/api/health` → `{"status":"ok"}`; Swagger UI: http://localhost:8080/docs.

## Lệnh thường dùng

| Mục đích | Lệnh |
| --- | --- |
| Chạy API | `go run ./cmd/app` |
| Migration up / down | `go run ./cmd/migrate -direction=up` / `-direction=down` |
| Build toàn bộ | `go build ./...` |
| Kiểm tra tĩnh | `go vet ./...` |
| Format | `gofmt -w <files>` |
| Sinh lại Swagger | `go run github.com/swaggo/swag/cmd/swag@v1.16.6 init --parseInternal -g cmd/app/main.go -o api/swagger --packageName swagger` |
| PostgreSQL up / down | `docker compose up -d cinema-postgres` / `docker compose down` |

## Biến môi trường

| Biến | Mặc định | Mô tả |
| --- | --- | --- |
| `HTTP_ADDR` | `:8080` | Địa chỉ HTTP |
| `DATABASE_URL` | bắt buộc | PostgreSQL URL |
| `DB_MAX_OPEN_CONNS` / `DB_MAX_IDLE_CONNS` | `25` / `5` | Pool |
| `DB_LOG_SQL` | `false` | `true` log mọi câu SQL; mặc định chỉ log lỗi và query > 200ms |
| `LOG_LEVEL` | `info` | `debug` `info` `warn` `error` |
| `LOG_FORMAT` | `text` | `text` (dev, log SQL có màu) hoặc `json` (prod, không màu) |
| `POSTGRES_*` | `cinema` / `5434` | Chỉ dùng bởi Docker Compose |

## Quy ước

### Lỗi trả về

Mọi lỗi đều là JSON cùng dạng, do `internal/handlers/error_handler.go` render:

```json
{"errorCode": 404, "errorMessage": "resource not found"}
{"errorCode": 400, "errorMessage": "email is required", "errors": [{"field": "email", "message": "email is required"}]}
```

Trong handler: bind + validate bằng `utils.BindAndValidate(c, &req)`; lỗi từ service đi qua `utils.ServiceError(err)`, hàm này map `gorm.ErrRecordNotFound` → 404, còn lại → 500 (chi tiết chỉ ra log, không ra client). Map mã lỗi Postgres (23505, 23P01...) sẽ thêm khi có endpoint ghi dữ liệu. Lỗi có trạng thái sẵn (`utils.APIError(status, msg)`) đi qua nguyên vẹn.

### Swagger

- Thông tin chung nằm trên `main()` trong `cmd/app/main.go`.
- Mỗi handler public có block annotation ngay trên hàm: `@Summary`, `@Tags`, `@Accept`/`@Produce`, `@Param`, `@Success`, `@Failure` (dùng `utils.APIErrorResponse` / `utils.ValidationError`), `@Router /path [method]` (đường dẫn tương đối với `@BasePath /api`); endpoint cần đăng nhập thêm `@Security BearerAuth`.
- DTO response có tag `example:"..."`. Sau khi sửa annotation chạy lệnh sinh Swagger ở trên và commit `api/swagger/`.

### Migration

- File `migrations/YYYYMMDDNNNN_mo_ta.go`, `ID` trùng tiền tố; thêm vào danh sách trong `migrations.go` theo thứ tự.
- Mỗi `tx.Exec` chỉ một câu SQL (pgx extended protocol không nhận nhiều câu). Luôn viết `Rollback`.
