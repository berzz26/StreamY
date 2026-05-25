# StreamY — Project Structure Guide

## Overview

StreamY is a backend-focused video streaming platform written in Go.

The system handles:

* Video uploads
* Transcoding using FFmpeg
* HLS chunk generation
* Multi-resolution streaming
* HTTP delivery of video chunks
* Background processing workers

The architecture is intentionally designed similar to real-world production backend systems.

---

# Root Project Structure

```text
streamy/
│
├── cmd/
├── internal/
├── storage/
├── scripts/
├── web/
│
├── .env
├── .gitignore
├── docker-compose.yml
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

---

# cmd/

Contains executable applications.

Each folder inside `cmd/` becomes its own binary.

This is standard production Go architecture.

---

## cmd/api/

```text
cmd/api/main.go
```

Main HTTP API server.

Responsibilities:

* Starts the HTTP server
* Loads configuration
* Connects to PostgreSQL
* Initializes routes
* Registers dependencies
* Handles uploads and streaming APIs

Example responsibilities:

```text
POST /upload
GET  /videos/:id
GET  /stream/:id/master.m3u8
```

---

## cmd/worker/

```text
cmd/worker/main.go
```

Background transcoding worker.

Responsibilities:

* Picks transcoding jobs
* Executes FFmpeg
* Generates HLS chunks
* Creates video variants
* Updates processing status

This process runs separately from the API.

Why?

Because transcoding is CPU intensive and should never block API requests.

---

# internal/

Contains all private application logic.

Anything inside `internal/` cannot be imported outside the module.

This is where the real backend system lives.

---

# internal/config/

```text
internal/config/config.go
```

Responsible for application configuration.

Loads:

```env
PORT=
DATABASE_URL=
STORAGE_PATH=
FFMPEG_PATH=
```

Purpose:

* Centralized config management
* Environment variable loading
* Avoid hardcoded values

---

# internal/database/

```text
internal/database/
    postgres.go
    migrations/
```

Responsible for database connectivity and schema management.

---

## postgres.go

Creates PostgreSQL connection pools.

Responsibilities:

* Connect to DB
* Configure pooling
* Verify connectivity

Usually returns:

```go
*pgxpool.Pool
```

---

## migrations/

Contains SQL schema migration files.

Example:

```sql
001_create_videos_table.sql
```

Purpose:

* Version controlled database schema
* Easy reproducible setup
* Production-safe DB changes

---

# internal/models/

```text
internal/models/video.go
```

Contains pure domain structs.

Example:

```go
type Video struct {
    ID string
    Title string
}
```

Purpose:

* Shared data models
* DB representations
* API representations
* Internal entity definitions

Important:

Models should NOT contain business logic.

---

# internal/repository/

```text
internal/repository/video_repository.go
```

Responsible for database queries.

Purpose:

* Isolate SQL logic
* Separate persistence from business logic

Typical methods:

```go
CreateVideo()
GetVideoByID()
UpdateStatus()
```

Repositories should ONLY interact with the database.

---

# internal/storage/

```text
internal/storage/
    local.go
    paths.go
```

Abstracts physical file storage.

Initially:

* Local filesystem

Later:

* S3
* MinIO
* Cloud object storage

Purpose:

* Business logic should not care where files live
* Easier migration to cloud storage later

---

## local.go

Handles:

* Save files
* Read files
* Delete files
* Create directories

---

## paths.go

Centralizes storage path generation.

Example:

```text
/storage/originals/
/storage/processed/
```

Avoids path duplication throughout the project.

---

# internal/upload/

```text
internal/upload/
    handler.go
    service.go
    validator.go
```

Responsible for the upload pipeline.

Flow:

```text
Upload request
    ↓
Validate file
    ↓
Save original video
    ↓
Insert DB record
    ↓
Queue transcoding job
```

---

## handler.go

HTTP layer only.

Responsibilities:

* Parse multipart form
* Read uploaded file
* Return JSON response

Should NOT contain business logic.

---

## service.go

Core upload workflow.

Responsibilities:

* Validation
* File naming
* File storage
* DB insertion
* Queue publishing

This is where most upload logic lives.

---

## validator.go

Responsible for validating uploaded videos.

Examples:

* File type validation
* MIME validation
* Max file size
* Extension checks

---

# internal/transcoder/

```text
internal/transcoder/
    ffmpeg.go
    hls.go
    worker.go
    job.go
```

Core media processing system.

This is the heart of StreamY.

---

## ffmpeg.go

Executes FFmpeg commands.

Responsibilities:

* Video transcoding
* Resolution conversion
* Audio conversion
* Segment generation

Usually uses:

```go
exec.Command()
```

---

## hls.go

Handles HLS-specific generation.

Responsibilities:

* Playlist generation
* Segment settings
* Multi-resolution setup
* Adaptive streaming logic

---

## worker.go

Background worker loop.

Responsibilities:

* Listen for jobs
* Process videos
* Handle retries
* Update statuses

Example:

```go
for job := range queue {
    process(job)
}
```

---

## job.go

Defines transcoding job structures.

Example:

```go
type Job struct {
    VideoID string
    InputPath string
}
```

---

# internal/streaming/

```text
internal/streaming/
    handler.go
    service.go
```

Responsible for video delivery.

Serves:

* .m3u8 playlists
* .ts chunks
* future MP4 range requests

---

## handler.go

HTTP layer.

Routes:

```text
GET /stream/:id/master.m3u8
GET /stream/:id/720/segment1.ts
```

---

## service.go

Streaming business logic.

Responsibilities:

* Locate playlist files
* Locate chunks
* Verify video exists
* Handle streaming rules

Later may include:

* Signed URLs
* Authorization
* CDN headers
* Cache control

---

# internal/queue/

```text
internal/queue/memory_queue.go
```

Job queue system.

Initially:

* In-memory Go channels

Later:

* Redis
* RabbitMQ
* Kafka

Purpose:

Decouple uploads from transcoding.

Without queues:

Uploads would block while FFmpeg runs.

---

# internal/middleware/

```text
internal/middleware/
    logger.go
    recovery.go
```

Reusable HTTP middleware.

---

## logger.go

Logs requests.

Example:

```text
POST /upload 200 120ms
```

---

## recovery.go

Prevents server crashes.

Recovers from panics and returns proper HTTP errors.

Critical in production APIs.

---

# internal/utils/

```text
internal/utils/
    file.go
    response.go
    errors.go
```

Shared helper utilities.

---

## file.go

Reusable file helper functions.

Examples:

* File extension helpers
* File size helpers
* Directory utilities

---

## response.go

Standardized API responses.

Example:

```json
{
  "success": true,
  "data": {}
}
```

---

## errors.go

Shared custom errors.

Purpose:

* Reusable error definitions
* Cleaner error handling

---

# storage/

```text
storage/
    originals/
    processed/
```

Physical video storage.

---

## storage/originals/

Stores raw uploaded videos.

Example:

```text
/storage/originals/video123.mp4
```

These files are the source material for transcoding.

---

## storage/processed/

Stores generated streaming assets.

Example:

```text
/storage/processed/video123/
    master.m3u8
    segment0.ts
```

Contains:

* HLS playlists
* Segments
* Multi-resolution variants

---

# scripts/

```text
scripts/
    dev.sh
    ffmpeg_test.sh
```

Developer utility scripts.

---

## dev.sh

Used for local development automation.

Examples:

* Start API
* Start worker
* Run migrations

---

## ffmpeg_test.sh

Used for quickly testing FFmpeg commands.

Useful during transcoding experimentation.

---

# web/

```text
web/test-player/
```

Tiny frontend for testing playback.

Not intended to be the final frontend.

Purpose:

* Verify streaming works
* Test HLS playback
* Debug chunk delivery

---

## index.html

Contains:

```html
<video controls>
```

---

## player.js

Uses:

* hls.js

Loads:

```text
master.m3u8
```

for browser playback.

---

# .env

Environment variables.

Example:

```env
PORT=8080
DATABASE_URL=postgres://...
FFMPEG_PATH=ffmpeg
```

Purpose:

* Keep secrets/config outside code
* Easier deployment

Never commit this file.

---

# .gitignore

Specifies ignored files.

Examples:

```text
.env
storage/
*.log
bin/
```

---

# docker-compose.yml

Defines local infrastructure services.

Initially:

* PostgreSQL

Later:

* Redis
* MinIO
* RabbitMQ

Purpose:

Easy local development environment.

---

# go.mod

Go module definition.

Contains:

* Module name
* Dependencies
* Versioning

Equivalent to:

```text
package.json
```

for Node.js.

---

# go.sum

Dependency checksum verification.

Automatically managed by Go.

Ensures:

* Dependency integrity
* Reproducible builds

---

# Makefile

Development command shortcuts.

Example:

```make
run-api:
	go run ./cmd/api

run-worker:
	go run ./cmd/worker
```

Purpose:

* Faster workflows
* Consistent commands
* Production-like tooling

---

# README.md

Project documentation.

Should include:

* Setup instructions
* Architecture overview
* API docs
* Development workflow
* Deployment notes

---

# Overall System Flow

```text
Client Uploads Video
        ↓
Upload Service
        ↓
Store Original File
        ↓
Create DB Record
        ↓
Queue Job
        ↓
Worker Picks Job
        ↓
FFmpeg Transcoding
        ↓
Generate HLS Chunks
        ↓
Store Processed Files
        ↓
Streaming Service Delivers Chunks
        ↓
Browser Plays Video
```

---

# Long-Term Evolution

Future additions may include:

* Authentication
* CDN integration
* Distributed transcoding
* Edge streaming
* Signed URLs
* Analytics
* Watch history
* Live streaming
* Subtitles
* Recommendations
* AI moderation

The current structure is intentionally designed to scale into those systems cleanly.
