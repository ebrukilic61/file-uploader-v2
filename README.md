# File Uploader

Bu proje, büyük dosyaları chunklar halinde upload etmek için tasarlanmış bir Go backend uygulamasıdır. Clean Architecture pattern'i kullanılarak geliştirilmiştir.

## Özellikler

- Chunk-based file upload
- Hash doğrulama (SHA-256)
- Upload durumu kontrolü
- Eksik chunk kontrolü
- Atomik dosya işlemleri
- Clean Architecture

## Proje Yapısı

```
File-Uploader/
├── cmd/
│   └── server/
│       └── main.go              # Ana uygulama
│── docs/
│   ├── docs.go
│   ├── swagger.yaml
├── internal/
│   ├── delivery/
│   │   └── http/
│   │     └── handlers/
│   │       └── upload_handler.go # HTTP handler'lar
│   │       └── media_handler.go
│   │       └── cleanup_handler.go
│   │     └── routers/
│   │       └── media_routers.go
│   │       └── upload_routers.go
│   ├── domain/
│   │   ├── dto/
│   │   │   └── upload_dto.go    # Data Transfer Objects
│   │   │   └── images_dto.go    
│   │   │   └── media_dto.go    
│   │   ├── entities/
│   │   │   └── images.go        # Domain entities
│   │   │   └── media.go
│   │   │   └── upload.go
│   │   ├── repositories/
│   │   │   ├── media_repo.go # Media Repository interface
│   │   │   ├── upload_repo.go # Upload Repository interface
│   │   │   └── storage_repo.go
│   │   └── mapper/
│   │       └── mapper.go    # Data Transfer Objects
│   ├── infrastructure/
│   │   ├── db/
│   │   │    └── db.go
│   │   │    └── migrate.go
│   │   ├── processor/
│   │   │   └── image.go
│   │   ├── queue/
│   │   │   └── job.go
│   │   │   └── worker_pool.go
│   │   │   └── worker.go
│   │   ├── repositories/
│   │   │   └── file_upload_repository.go
│   │   │   └── media_repository.go
│   │   │   └── media_size_repository.go
│   │   │   └── media_variant_repository.go
│   │   ├── storage/
│   │   │   └── local_storage.go
│   │   │   └── s3_storage.go
│   │   ├── usecases/
│   │   │   └── cleanup.go
│   │   │   └── file_uploader.go
│   │   │   └── media.go
│   ├── pkg/
│   │   ├── config/              # Konfigürasyon yönetimi
│   │   │   └── config.go
│   │   ├──  constants/
│   │   │   └── status.go
│   │   ├──  errors/
│   │   │   ├──  i18n/
│   │   │   │    └── en.json
│   │   │   │    └── error_translator.go
│   │   │   │    └── tr.json
│   │   │   └── error_handler.go
│   │   │   └── upload_error.go
│   │   ├──  file/
│   │   │   └── calculate_file_hash.go
│   │   │   └── copy_file.go
│   │   │   └── make_key.go
│   │   │   └── validate_file_hash.go
│   │   ├──  response/
│   │   │   └── response.go
├── .env                  # Environment
├── .gitignore
├── go.mod
└── go.sum
```
# Kurulum

## Gereksinimler
- Go 1.21+
- Redis 8

## Klasör Konumları

### Varsayılan Konumlar:
- **Temp Klasörü**: `./temp_uploads/` (proje kökünde belirtildi config dosyası içerisinde -> cmd/server içerisinde oluşturuluyor)
- **Uploads Klasörü**: `./uploads/` (proje kökünde belirtildi config dosyası içerisinde -> cmd/server içerisinde oluşturuluyor)
- **Media Klasörü**: `./uploads/media` (uploads klasörü içerisind ebulunmaktadır. image dosyalarının orijinal (original) ve varyant (variant) halleri bu dosya içerisinde tutulmaktadır. Aynı zamanda buradaki veriler veri tabanı tablolarına aktarılmaktadır.)

## API Endpoints

### 1. Upload Chunk
```
GET /api/v1/upload/chunk?upload_id={upload_id}&filename={filename}&chunk_index={chunk_index}
```

**Response:**
```json
{
    "status": "queued",
    "upload_id": "upload_id",
    "chunk_index": 1,
    "filename": "file.filetype",
    "message": "chunk işleme kuyruğuna alındı"
}
```

### 2. Complete Upload
```
POST /api/v1/upload/complete
Content-Type: multipart/form-data

Form fields:
- upload_id: string
- total_chunks: int
- filename: string
```

**Response:**
```json
{
  "status": "ok",
  "message": "Dosya birleştirildi",
  "filename": "test.txt"
}
```

### 3. Cancel Upload
```
POST /api/v1/upload/cancel
Content-Type: multipart/form-data

Form fields:
- upload_id: string
- filename: string
```

**Response**
```json
{
    "status": "queued",
    "message": "Upload iptal edildi"
}
```

### 4. Upload Status 
```
GET /api/v1/upload/status
Content-Type: multipart/form-data

Form fields:
- upload_id: string
- filename: string
```

**Response**
```json
{
    "upload_id": "upload_id",
    "filename": "filename.filetype",
    "uploaded_chunks": 0,
    "status": "failed/completed"
}
```

## Güvenlik Özellikleri

- **Hash Doğrulama**: Chunk'ların bütünlüğünü kontrol etmek için SHA-256 hash kullanılır
- **Idempotent Upload**: Aynı chunk'ın tekrar gönderilmesi durumunda hata vermez
- **Dosya Yolu Güvenliği**: `filepath.Base()` kullanılarak path traversal saldırıları önlenir
- **Atomik İşlemler**: Geçici dosyalar kullanılarak dosya yazma işlemleri atomik hale getirilir

## Hata Yönetimi

- Eksik parametreler için 400 Bad Request
- Dosya işlem hataları için 500 Internal Server Error
- Hash doğrulama başarısız olursa 400 Bad Request
- Eksik chunk'lar varsa 400 Bad Request

## Performans

- Chunk boyutu ayarlanabilir (varsayılan: 10 MB)
- Paralel upload desteği (worker-redis-server yapısı)
- Geçici dosyalar otomatik temizlenir
- Memory-efficient streaming işlemler
"# file-uploader-v2" 
