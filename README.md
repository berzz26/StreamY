root folder struct:
video-platform/
│
├── cmd/
│   ├── api/
│   │   └── main.go
│   │
│   └── worker/
│       └── main.go
│
├── internal/
│   ├── config/
│   │   └── config.go
│   │
│   ├── database/
│   │   ├── postgres.go
│   │   └── migrations/
│   │       └── 001_create_videos_table.sql
│   │
│   ├── models/
│   │   └── video.go
│   │
│   ├── repository/
│   │   └── video_repository.go
│   │
│   ├── storage/
│   │   ├── local.go
│   │   └── paths.go
│   │
│   ├── upload/
│   │   ├── handler.go
│   │   ├── service.go
│   │   └── validator.go
│   │
│   ├── transcoder/
│   │   ├── ffmpeg.go
│   │   ├── hls.go
│   │   ├── worker.go
│   │   └── job.go
│   │
│   ├── streaming/
│   │   ├── handler.go
│   │   └── service.go
│   │
│   ├── queue/
│   │   └── memory_queue.go
│   │
│   ├── middleware/
│   │   ├── logger.go
│   │   └── recovery.go
│   │
│   └── utils/
│       ├── file.go
│       ├── response.go
│       └── errors.go
│
├── storage/
│   ├── originals/
│   └── processed/
│
├── scripts/
│   ├── dev.sh
│   └── ffmpeg_test.sh
│
├── web/
│   └── test-player/
│       ├── index.html
│       └── player.js
│
├── .env
├── .gitignore
├── docker-compose.yml
├── go.mod
├── go.sum
├── Makefile
└── README.md