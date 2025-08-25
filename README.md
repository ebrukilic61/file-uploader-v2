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
│   │       └── upload_handler.go # HTTP handler'lar
│   ├── domain/
│   │   ├── entities/
│   │   │   └── upload.go        # Domain entities
│   │   ├── models/
│   │   │   └── upload_dto.go    # Data Transfer Objects
│   │   └── repositories/
│   │       ├── file_repository.go # Dosya işlemleri
│   │       └── upload_repo.go   # Repository interface
│   ├── infrastructure/
│   │   └── storage/
│   │       └── local_storage.go
│   │       └── s3_storage.go
│   ├── pkg/
│   │   └── config/              # Konfigürasyon yönetimi
│   │       └── config.go
│   └── usecases/
│       └── file_uploader.go     # İş mantığı
├── env.example                  # Environment variables örneği
├── go.mod
└── go.sum
```
# Kurulum

## Gereksinimler
- Go 1.21+
- FFmpeg

## FFmpeg Kurulumu

### Windows:
```bash
winget install ffmpeg
```

### macOS:
```bash
brew install ffmpeg
```
### Linux (Ubuntu/Debian):
```
sudo apt update && sudo apt install ffmpeg
```
=======
>>>>>>> acde494ad40329b8c0d1989b584c362ce7ec4ccc

## Gereksinimler
- Go 1.21+
- FFmpeg

## FFmpeg Kurulumu

### Windows:
```bash
winget install ffmpeg
```
### macOS:
```bash
brew install ffmpeg
```
### Linux (Ubuntu/Debian):
```
sudo apt update && sudo apt install ffmpeg
```
## Klasör Konumları

### Varsayılan Konumlar:
- **Temp Klasörü**: `./temp_uploads/` (proje kökünde belirtildi config dosyası içerisinde -> cmd/server içerisinde oluşturuluyor)
- **Uploads Klasörü**: `./uploads/` (proje kökünde belirtildi config dosyası içerisinde -> cmd/server içerisinde oluşturuluyor)

## API Endpoints

### 1. Upload Status
```
GET /api/v1/upload/status?upload_id={upload_id}&filename={filename}
```

**Response:**
```json
{
  "upload_id": "test-123",
  "filename": "test.txt",
  "uploaded_chunks": [1, 2, 3]
}
```

### 2. Upload Chunk
```
POST /api/v1/upload/chunk
Content-Type: multipart/form-data

Form fields:
- upload_id: string
- chunk_index: string
- filename: string
- chunk_hash: string (optional)
- file: file
```

**Response:**
```json
{
  "status": "ok",
  "upload_id": "test-123",
  "chunk_index": 1,
  "filename": "test.txt"
}
```

### 3. Complete Upload
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
- Paralel upload desteği
- Geçici dosyalar otomatik temizlenir
- Memory-efficient streaming işlemler
