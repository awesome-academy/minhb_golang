# Cinema Booking API

Backend web đặt vé rạp chiếu phim viết bằng Go, Echo v5, GORM v2 và PostgreSQL 15; Redis 7 chỉ giữ session admin SSR. Code tổ chức theo luồng `handler → service → repository`, có Swagger UI tại `/swaggers`.

## Yêu cầu

- Go 1.26+
- Docker và Docker Compose (PostgreSQL và Redis local)

## Cấu trúc thư mục

```text
cmd/
├── app/                 # HTTP entry point
└── migrate/             # CLI chạy migration up/down
config/                  # Load .env / biến môi trường
api/swagger/             # Swagger spec sinh tự động — không sửa tay
migrations/              # gormigrate, mỗi migration là raw SQL trong file Go (7 migration: extension, enum, 12 bảng)
internal/
├── dto/                 # Request/response DTO (json + validate + example tags)
├── errors/              # Sentinel error nghiệp vụ dùng chung (ErrInvalidCredentials, ErrSessionNotFound), import alias apperrors
├── middleware/          # AdminCSRF, AdminNoStore, RequireAdminSession, helper cookie AdminSessionCookie
├── handlers/            # routes.go duy nhất đăng ký mọi route, health handler, HTTP error handler
│   ├── admin/           # Handler admin SSR (package admin)
│   └── user/            # Handler REST API cho user (package user, thêm từ U1)
├── models/              # GORM model
├── repositories/        # Truy vấn GORM
├── services/            # Business logic
└── utils/               # Error response thống nhất, validator
pkg/
├── db/                  # Kết nối GORM/pgx, pool
├── logger/              # Khởi tạo slog
└── redis/               # Kết nối go-redis (session admin), ping khi khởi động
web/                     # Admin SSR: template html/template + static CSS, embed vào binary
├── templates/admin/     # layout.html + một file mỗi trang (login.html, dashboard.html, error.html...), cho phép thư mục con
└── static/              # bootstrap.min.css (5.3.3, vendored) + admin.css
```

## Chạy local

```bash
cp .env.example .env                                # mặc định khớp docker-compose
docker compose up -d cinema-postgres cinema-redis   # PostgreSQL 15 tại :5434, Redis 7 tại :6380
go run ./cmd/migrate -direction=up                  # tạo extension, enum và 12 bảng theo thiết kế
go run ./cmd/app                                    # API tại http://localhost:8080
```

Kiểm tra: `curl localhost:8080/api/health` → `{"status":"ok"}`; Swagger UI: http://localhost:8080/swaggers; trang admin: http://localhost:8080/admin/login.

### Tạo admin dev

Chưa có seed (S4) nên tạo tay một admin. Hash bcrypt sinh bằng `htpasswd` (không có thì dùng image `httpd`):

```bash
HASH=$(docker run --rm httpd:2-alpine htpasswd -bnBC 10 "" 'Admin@12345' | tr -d ':\n' | sed 's/^\$2y/\$2a/')
docker compose exec -T cinema-postgres psql -U cinema -d cinema <<SQL
INSERT INTO users (email, password_hash, full_name, role)
VALUES ('admin@cinema.local', '$HASH', 'Cinema Admin', 'admin')
ON CONFLICT (lower(email)) DO UPDATE SET password_hash = EXCLUDED.password_hash, role = 'admin', deleted_at = NULL;
SQL
```

Đăng nhập tại `/admin/login` với `admin@cinema.local` / `Admin@12345` (chỉ dùng local; prod sinh hash riêng).

## Lệnh thường dùng

| Mục đích | Lệnh |
| --- | --- |
| Chạy API | `go run ./cmd/app` |
| Migration up / down | `go run ./cmd/migrate -direction=up` / `-direction=down` |
| Build toàn bộ | `go build ./...` |
| Kiểm tra tĩnh | `go vet ./...` |
| Format | `gofmt -w <files>` |
| Sinh lại Swagger | `go run github.com/swaggo/swag/cmd/swag@v1.16.6 init --parseInternal -g cmd/app/main.go -o api/swagger --packageName swagger` |
| PostgreSQL + Redis up / down | `docker compose up -d cinema-postgres cinema-redis` / `docker compose down` |
| Xem session admin trong Redis | `docker compose exec -T cinema-redis redis-cli --scan --pattern 'admin_session:*'` |

## Biến môi trường

| Biến | Mặc định | Mô tả |
| --- | --- | --- |
| `HTTP_ADDR` | `:8080` | Địa chỉ HTTP |
| `DATABASE_URL` | bắt buộc | PostgreSQL URL |
| `DB_MAX_OPEN_CONNS` / `DB_MAX_IDLE_CONNS` | `25` / `5` | Pool |
| `DB_LOG_SQL` | `false` | `true` log mọi câu SQL; mặc định chỉ log lỗi và query > 200ms |
| `LOG_LEVEL` | `info` | `debug` `info` `warn` `error` |
| `LOG_FORMAT` | `text` | `text` (dev, log SQL có màu) hoặc `json` (prod, không màu) |
| `REDIS_URL` | bắt buộc | Redis URL (`redis://[:pass@]host:port/db`), chỉ giữ session admin; thiếu hoặc ping lỗi → app dừng ngay |
| `ADMIN_SESSION_TTL` | `8h` | Go duration > 0; là TTL key Redis và `Max-Age` cookie `admin_session`, không sliding |
| `ADMIN_COOKIE_SECURE` | `false` | `true` khi admin chạy qua HTTPS; dev HTTP phải để `false` (cookie `Secure` không gửi qua HTTP) |
| `POSTGRES_*` | `cinema` / `5434` | Chỉ dùng bởi Docker Compose |
| `REDIS_PORT` | `6380` | Chỉ dùng bởi Docker Compose (host port của `cinema-redis`) |

## Quy ước

### API

- Endpoint chi tiết nhận `id` (`/api/movies/:id`, `/api/theaters/:id`, `/api/bookings/:id`), không dùng `slug`. `slug` chỉ là dữ liệu trả về cho FE dựng URL.

### Admin SSR

- Template trong `web/templates/admin/`: `layout.html` định nghĩa `layout` và `{{block "content" .}}`; mỗi trang một file `{{define "content"}}`, được parse riêng cùng layout và gọi bằng `c.Render(200, "admin/<đường dẫn không .html>", view)` (ví dụ `admin/login`, `admin/movies/list`).
- Thêm trang mới: tạo file `.html` và một view struct khai báo ngay trong file handler tương ứng (có `Title`, và `CSRFToken` nếu trang có form; trang protected thêm `AdminEmail`). Không dùng `map[string]any`; template bật `missingkey=error`. Sửa template phải chạy lại app (embed).
- Header: `layout.html` có `{{block "userbar" .}}{{end}}` (mặc định rỗng nên trang public/lỗi không cần field thêm); trang đã đăng nhập `{{define "userbar"}}` hiện email + nút Logout (form POST `/admin/logout` có `_csrf`). Xem `dashboard.html`.
- CSS: Bootstrap 5.3.3 vendored tại `web/static/bootstrap.min.css` (không CDN, chạy offline) + `admin.css` tông trung tính. Không thêm JS khi chưa cần.
- Route admin: `g := e.Group("/admin")` có `GET/POST /login`, `POST /logout` (public) và `protected := g.Group("", adminSession)` cho trang cần đăng nhập (`GET /admin` dashboard; M2–M6 đăng ký vào `protected`). Echo khởi tạo bằng `echo.NewWithConfig(echo.Config{NoGroupAutoRegister404Routes: true})` nên group có middleware không tự thêm route catch-all → sai path vẫn 404, sai method vẫn 405. `AdminCSRF` và `AdminNoStore` gắn `e.Use` toàn cục với Skipper theo `middleware.IsAdminPath`; `main.go` dựng middleware session rồi truyền vào `handlers.RegisterRoutes` (`routes.go` không import `repositories`).
- Session admin (lưu Redis, không JWT)
- Mọi response dưới `/admin` có `Cache-Control: no-store` (`AdminNoStore`) để nút Back sau logout không hiện trang cũ.
- Lỗi dưới `/admin` (404 sai path, 405 sai method, 400 CSRF, 500...) render `admin/error.html`
- CSRF: token đọc từ form `_csrf` hoặc header `X-CSRF-Token`, cookie `_csrf` HttpOnly, SameSite=Lax, path `/admin`. Mọi form admin phải có `<input type="hidden" name="_csrf" value="{{.CSRFToken}}">`.
- Mọi text hiển thị trên UI (label, nút, tiêu đề, thông báo lỗi) viết bằng tiếng Anh; `<html lang="en">`.

### Lỗi trả về

Mọi lỗi đều là JSON cùng dạng, do `internal/handlers/error_handler.go` render (riêng path `/admin`, `/admin/*` render trang HTML `admin/error.html` cùng status code):

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
- Mỗi `tx.Exec` chỉ một câu SQL (pgx extended protocol không nhận nhiều câu); nhiều câu dùng `execAll(tx, ...)`. Luôn viết `Rollback` theo thứ tự ngược.
- Ràng buộc nghiệp vụ nằm ở DB (`showtimes_no_overlap`, `tickets_one_active_per_seat`, `users_email_lower_key`, composite FK `tickets_booking_id_showtime_id_fkey`); tên constraint dùng để map lỗi Postgres sau này.

### Model

- `internal/models`, một file một bảng; ID `int64`, cột nullable dùng pointer, enum Postgres là kiểu `string` có const (`models.UserRoleAdmin`).
- Tiền `numeric(12,2)` dùng `decimal.Decimal` (shopspring); `cast_members` jsonb là `models.CastMembers` (tự Scan/Value); soft delete qua `gorm.DeletedAt`.
- Tag `default:` phản chiếu DEFAULT của DB nên `Create` không cần set `Role`, `Status`, `Format`, `Currency`, `QRCode`. Riêng `IsActive` (bool) không có tag `default` vì GORM sẽ ghi đè `false` thành `true`; service phải set `IsActive` tường minh khi tạo.
