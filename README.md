# Notly API

Notly API, not alma ve kişisel planlama akışlarını tek backend altında toplayan modüler bir Go servisidir. Kimlik doğrulama, ders ve görev yönetimi, alışkanlık takibi, takvim senkronizasyonu, paylaşılan notlar, Pomodoro oturumları ve gerçek zamanlı bildirimler için REST API sağlar.

## Öne Çıkanlar

- Katmanlı modüler monolit mimarisi: `domain`, `repository`, `service`, `http`
- JWT tabanlı kimlik doğrulama ve refresh token akışı
- PostgreSQL üzerinde SQL migration yönetimi
- Not paylaşımı, collaborator yönetimi ve dosya eki desteği
- Google Calendar entegrasyonu için OAuth tabanlı senkronizasyon altyapısı
- WebSocket üzerinden gerçek zamanlı bildirim yayınlama
- Prometheus metrikleri ve Zap tabanlı yapısal loglama
- Swagger UI ile API dokümantasyonu

## Teknoloji Yığını

- Go 1.25
- PostgreSQL
- gorilla/mux
- sqlx ve lib/pq
- golang-migrate
- golang-jwt
- go-playground/validator
- swaggo/http-swagger
- prometheus/client_golang
- zap

## Proje Yapısı

```text
cmd/api/                         Uygulama giriş noktası
docs/                            Swagger çıktıları
internal/app/                    HTTP server ve route kayıtları
internal/common/                 Ortak response, validation ve yardımcı tipler
internal/infrastructure/         Database, middleware, jobs, logging, email, websocket
internal/modules/                İş alanı modülleri
internal/modules/auth/           Auth ve email verification
internal/modules/note/           Notlar, paylaşım ve ekler
internal/modules/task/           Görevler ve alt görevler
internal/modules/calendar/       Google Calendar entegrasyonu
internal/modules/habit/          Alışkanlık takibi
internal/modules/course/         Ders ve program yönetimi
internal/modules/pomodoro/       Pomodoro oturumları
```

## Kurulum

Gereksinimler:

- Go 1.25+
- PostgreSQL 14+
- `golang-migrate` CLI, sadece migration komutlarını Makefile üzerinden çalıştırmak için gerekli

Adımlar:

```bash
git clone https://github.com/M1ralai/notly-api.git
cd notly-api
cp .env.example .env
go mod download
```

`.env` dosyasındaki PostgreSQL bilgilerini kendi lokal veritabanına göre güncelle:

```env
API_PORT=8080
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=change-me
DB_NAME=notly
DB_SSLMODE=disable
JWT_SECRET=change-me-use-a-long-random-value
```

Uygulamayı başlat:

```bash
go run cmd/api/main.go
```

Uygulama açılırken `internal/infrastructure/database/migrations` altındaki migration dosyalarını otomatik çalıştırır.

## Geliştirme Komutları

```bash
make run          # API server'ı çalıştırır
make build        # bin/api çıktısını üretir
make test         # testleri çalıştırır
make test-cover   # coverage ile test çalıştırır
make migrate-up   # bekleyen migration'ları uygular
```

## API Dokümantasyonu

Server çalışırken Swagger UI:

```text
http://localhost:8080/swagger/index.html
```

Swagger endpoint'i kod içinde localhost ile sınırlandırılmıştır. Generated dosyalar `docs/` klasöründe tutulur.

## Başlıca Endpoint Grupları

- `POST /api/auth/register`
- `POST /api/auth/login`
- `POST /api/auth/refresh`
- `GET /health`
- `GET /metrics`
- `GET /api/notes`
- `POST /api/notes`
- `GET /api/tasks`
- `GET /api/habits`
- `GET /api/calendar/status`
- `GET /api/dashboard`
- `GET /ws`

Korumalı `/api/*` route'ları `Authorization: Bearer <token>` header'ı bekler.

## Opsiyonel Entegrasyonlar

Bu değerler boş bırakıldığında ilgili özellikler development ortamında pasif veya sınırlı çalışır:

- `RESEND_API_KEY` ve `EMAIL_FROM_ADDRESS`: email verification gönderimi
- `TURNSTILE_SECRET_KEY`: Turnstile doğrulaması
- `GOOGLE_CLIENT_ID`, `GOOGLE_CLIENT_SECRET`, `GOOGLE_REDIRECT_URI`: Google Calendar OAuth
- `ENCRYPTION_KEY`: üçüncü parti token şifreleme
- `MINIO_*`: not ekleri için S3 uyumlu obje depolama

## Notlar

Bu repo sadece API kodunu içerir. Deployment scriptleri, gerçek ortam dosyaları ve lokal çalışma kalıntıları public repoya dahil edilmemiştir.

## Lisans

MIT
