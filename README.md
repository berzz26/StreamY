# StreamY — Concurrent Media Processing Pipeline

## Demo 
**1.5x speed, the transcoding and final speeds are highlighted in the terminal and seek forward (at 1:02) to see the video player in action!**

https://github.com/user-attachments/assets/a50c3e3a-62db-4641-8885-b1e521e63c22

For the optimization process and to see before and after speeds, see the [optimization thread](https://x.com/berzzdotdev/status/2094271027174154309/video/1).

---

## Overview

StreamY is a backend-focused video processing and streaming system written in Go.

The goal of the project is to understand how video processing pipelines work and eventually build a distributed system capable of processing media workloads across multiple machines.

The current system takes an uploaded video, probes its metadata, dynamically plans the required renditions, encodes and segments those renditions into HLS, and stores the resulting streaming assets in MinIO.

The current focus is **single-machine media processing and optimization**. Distributed processing, worker scheduling, caching, and CDN/edge infrastructure are planned as later stages.

---

# Current Pipeline
<img width="4074" height="1070" alt="streamy architecture" src="https://github.com/user-attachments/assets/44fadb52-079e-4fca-a029-31473ac4eeff" />


Encoding is currently **sequential across renditions**, while object-storage uploads are performed concurrently. This was determined through benchmarking on the development machine.

---

# Performance

The pipeline started as a simple CPU-based implementation and was progressively optimized.

Test workload:

```text
Source size:       36,455,468 bytes (~36.5 MB)
Duration:          34.709 seconds
Source resolution: 1816 × 1080
Frame rate:        30 FPS
Source codec:      H.264
Renditions:        1080p, 720p, 480p, 360p
```

Development hardware:

```text
CPU:       AMD Ryzen 5 5600H
RAM:       16 GB
GPU:       NVIDIA RTX 3050 Laptop GPU
Encoder:   H.264 NVENC
Storage:   MinIO
Database:  PostgreSQL
```

### Optimization progression

| Pipeline            |                      Total time | Main change                                   |
| ------------------- | ------------------------------: | --------------------------------------------- |
| CPU encoding        |             ~105,000 ms (~105s) | Initial implementation                        |
| GPU encoding        |              50,213 ms (~50.2s) | H.264 NVENC                                   |
| Concurrent uploads  |             ~19,116 ms (~19.1s) | Concurrent MinIO uploads                      |
| Concurrent encoding | ~14,526–17,217 ms (~14.5–17.2s) | Parallel FFmpeg/NVENC                         |
| **Current best**    |            **9,117 ms (~9.1s)** | Sequential encoding + concurrent uploads + p4 |

The original CPU-based pipeline took roughly **1 minute 45 seconds** for the test video. Moving encoding to the RTX 3050 reduced this to approximately **50 seconds**.

The major bottleneck then turned out not to be encoding, but object-storage uploads. The original sequential upload path took **41,690 ms (~41.7s)**. Introducing concurrent uploads reduced this to approximately **1,057 ms (~1.06s)** in the best run — roughly a **39× reduction in upload time**.

The current best benchmark is:

```text
Probe:        112 ms
Transcode:  7,371 ms
Upload:     1,057 ms
Database:       8 ms
Total:      9,117 ms
```

This represents roughly an **11.5× reduction in total processing time** compared with the original CPU-based pipeline.

---

# Why Encoding Is Sequential

The renditions are independent workloads, so they can theoretically be encoded concurrently.

That was tested using Go's `errgroup` to run multiple FFmpeg processes simultaneously.

However, on the RTX 3050 Laptop GPU, concurrent NVENC encoding did not improve the end-to-end pipeline.

Parallel encoding reduced the transcode wall-clock time from roughly **7.3 seconds to ~6.7 seconds**, but upload time increased substantially, resulting in total processing times between approximately **14.5–17.2 seconds**.

The best sequential configuration remained approximately **9.1 seconds**.

This demonstrated an important distinction in the pipeline:

```text
GPU encoding
    → compute/resource constrained
    → more parallelism caused contention

MinIO uploads
    → I/O constrained
    → concurrency significantly improved throughput
```

For the current hardware and workload, sequential rendition encoding is therefore faster overall.

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

Each directory inside `cmd/` represents a separate executable.

---

## cmd/api/

```text
cmd/api/main.go
```

Main HTTP API server.

Responsibilities include:

* Starting the HTTP server
* Loading configuration
* Connecting to PostgreSQL
* Initializing dependencies
* Handling video upload requests
* Exposing video and streaming endpoints

---

## cmd/worker/

```text
cmd/worker/main.go
```

Runs the background media-processing worker.

The worker claims videos that need processing and runs the media pipeline independently from the API server.

Keeping processing outside the API process prevents long-running FFmpeg workloads from blocking upload requests.

---

# internal/

Contains the application's private business logic.

---

# internal/config/

```text
internal/config/
```

Responsible for application configuration.

Configuration includes things such as:

```env
PORT=
DATABASE_URL=
MINIO_ENDPOINT=
MINIO_ACCESS_KEY=
MINIO_SECRET_KEY=
MINIO_BUCKET=
```

Configuration is kept outside the application code so the same application can run against different environments.

---

# internal/database/

```text
internal/database/
```

Responsible for PostgreSQL connectivity and database migrations.

PostgreSQL stores application metadata and media-processing metadata rather than the actual video segments.

---

# internal/models/

```text
internal/models/
```

Contains the domain models used throughout the system.

Important models include:

```text
Video
VideoProbe
VideoRendition
VideoSegment
Run
```

### Video

Stores the primary metadata and processing state of a video.

```text
id
title
status
original_path
original_size
duration_seconds
processed_path
error_message
created_at
updated_at
```

### VideoProbe

Stores metadata extracted from the original media using `ffprobe`.

This includes information such as:

* Container format
* Duration
* Bitrate
* Video codec
* Video profile
* Resolution
* Frame rate
* Pixel format
* Color information
* Audio codec
* Audio bitrate
* Audio sample rate
* Audio channels

The probe information is kept separate from the main video record because it represents detailed media metadata rather than application state.

### VideoRendition

Represents a planned output representation of the source video.

A rendition contains information such as:

```text
video_id
width
height
bitrate
frame rate
```

For example:

```text
1080p
720p
480p
360p
```

The rendition planner determines which representations should be generated based on the source video's characteristics.

### VideoSegment

Represents an individual HLS media segment belonging to a rendition.

It stores metadata such as:

```text
variant_id
segment_index
duration_seconds
segment_path
```

This allows the database to describe the generated HLS structure without storing the actual media data inside PostgreSQL.

### Run

Stores timing and execution information for a media-processing run.

It is used to benchmark individual pipeline stages and compare different implementations.

---

# internal/repository/

```text
internal/repository/
```

Responsible for database access.

The repository layer isolates SQL and persistence logic from the media-processing pipeline.

Typical operations include:

```text
CreateVideo()
GetVideoByID()
ClaimNextVideo()
CreateVideoProbe()
UpdateVideoDuration()
UpdateVideoStatus()
MarkVideoFailed()
CreateRun()
```

---

# internal/storage/

```text
internal/storage/
```

Contains the object-storage integration.

StreamY currently uses **MinIO** as its S3-compatible object store.

The storage layer is responsible for:

* Uploading individual files
* Uploading directories
* Determining content types
* Uploading HLS playlists
* Uploading HLS segments
* Concurrent object uploads

---

## Upload Pipeline

The original storage implementation uploaded every file sequentially:

```text
segment_000.ts → MinIO
                     ↓
segment_001.ts → MinIO
                     ↓
segment_002.ts → MinIO
                     ↓
...
```

This became the largest bottleneck after GPU encoding was introduced.

The current implementation uses a worker pool:

```text
                    Upload Jobs
                        │
              ┌─────────┼─────────┐
              ▼         ▼         ▼
           Worker 1  Worker 2  Worker 3 ...
              │         │         │
              └─────────┼─────────┘
                        ▼
                      MinIO
```

The current benchmark uses **7 concurrent upload workers**, which produced the best result for the test environment.

The number is configurable and should not be considered universally optimal.

---

# internal/transcoder/

```text
internal/transcoder/
```

Contains the core media-processing pipeline.

This is currently the most important part of StreamY.

The package is responsible for:

* Probing input media
* Planning renditions
* Encoding video
* Generating HLS segments
* Creating playlists
* Measuring processing performance

---

## probe.go

Uses `ffprobe` to inspect the source media before transcoding.

The probe extracts structured JSON containing the source's format and stream metadata.

This information is converted into the `VideoProbe` model and stored in PostgreSQL.

The rendition planner then uses this information to make decisions about the output representations.

---

## ffmpeg.go

Contains the FFmpeg execution logic.

The current encoding pipeline uses:

```text
CUDA
  ↓
NVENC
  ↓
H.264
  ↓
HLS
```

The current NVENC configuration uses the `p4` preset after benchmarking it against `p5`.

---

## Rendition Planner

The rendition planner takes the source video's characteristics and determines which resolutions should be generated.

The system currently supports resolutions up to 4K:

```text
2160p
1440p
1080p
720p
480p
360p
240p
```

Only resolutions that make sense for the source video are selected.

For example, a 1080p source should not produce a 1440p or 2160p rendition.

The planner also derives the appropriate width while preserving the source aspect ratio.

---

## Encoding

Each planned rendition is encoded using FFmpeg and generated into its own temporary directory:

```text
processed/
└── <video-id>/
    ├── 720p/
    │   ├── index.m3u8
    │   ├── segment_000.ts
    │   ├── segment_001.ts
    │   └── ...
    │
    ├── 480p/
    │   ├── index.m3u8
    │   └── ...
    │
    ├── 360p/
    │   └── ...
    │
    └── 240p/
        └── ...
```

Each rendition has its own media playlist and segments.

A master playlist is generated after the renditions have completed.

---

# Processing Worker

The worker follows this general flow:

```text
Claim Video
    │
    ▼
Probe Video
    │
    ▼
Store Probe Metadata
    │
    ▼
Plan Renditions
    │
    ▼
Encode Renditions
    │
    ▼
Generate HLS
    │
    ▼
Upload Assets Concurrently
    │
    ▼
Generate / Upload Master Playlist
    │
    ▼
Update Video Status
    │
    ▼
Delete Temporary Files
```

The original uploaded video is currently kept only as temporary processing input. Once processing has successfully completed, the temporary source and local processed files are removed.

The durable media assets are stored in MinIO.

---

# internal/streaming/

```text
internal/streaming/
```

Responsible for serving the generated HLS assets.

The streaming layer works with the generated playlists and segments rather than performing media processing itself.

Typical requests include:

```text
GET /stream/:videoID/index.m3u8
GET /stream/:videoID/720p/index.m3u8
GET /stream/:videoID/720p/segment_000.ts
```

The generated master playlist allows an HLS client to select an appropriate rendition.

---

# internal/upload/

```text
internal/upload/
```

Responsible for handling incoming video uploads.

The general flow is:

```text
Upload request
      ↓
Validate
      ↓
Store temporary source
      ↓
Create video record
      ↓
Worker processes video
```

The API does not perform the expensive FFmpeg work itself.

---

# storage/

The local filesystem is used only as temporary working storage during processing.

```text
storage/
```

The uploaded source is temporarily stored locally, processed by FFmpeg, and removed after the resulting streaming assets have been persisted to MinIO.

MinIO is the durable object store for generated HLS assets.

---

# web/

```text
web/
```

Contains the lightweight frontend/player used to test the streaming pipeline.

It is primarily a development and validation tool rather than the focus of the project.

---

# scripts/

```text
scripts/
```

Contains development and media-processing utility scripts.

These are useful for testing FFmpeg behavior and automating common development workflows.

---

# docker-compose.yml

Used to run local infrastructure required by StreamY.

The development environment currently uses services such as:

```text
PostgreSQL
MinIO
```

Additional infrastructure will be introduced as the system evolves toward distributed processing.

---

# Current Architecture

The current system is intentionally single-machine:

```text
                         ┌──────────────┐
                         │   API Server │
                         └──────┬───────┘
                                │
                                ▼
                           PostgreSQL
                                │
                                │
                         Video Processing
                                │
                                ▼
                         ┌──────────────┐
                         │    Worker    │
                         └──────┬───────┘
                                │
                ┌───────────────┼───────────────┐
                │               │               │
                ▼               ▼               ▼
             FFmpeg          FFmpeg          FFmpeg
             NVENC           NVENC           NVENC
                │               │               │
                └───────────────┼───────────────┘
                                │
                         Concurrent Uploads
                                │
                                ▼
                             MinIO
```

Encoding itself is currently sequential across renditions because concurrent NVENC workloads performed worse on the available GPU.

The upload stage is concurrent because it is dominated by I/O and benefits significantly from multiple operations being in flight.

---

# Earlier Architecture

The original implementation was much simpler:

```text
Upload
  │
  ▼
Temporary Video
  │
  ▼
CPU FFmpeg
  │
  ▼
HLS
  │
  ▼
Sequential MinIO Uploads
  │
  ▼
PostgreSQL
```

There was initially no GPU acceleration, dynamic rendition planning, detailed probing, or concurrent upload worker pool.

The original implementation took approximately:

```text
~105,000 ms
(~1m45s)
```

for the test workload.

The current optimized single-machine pipeline reaches:

```text
9,117 ms
(~9.1s)
```

---

# Design Direction

The current implementation is intentionally being treated as a **V1 single-machine media-processing pipeline**.

The next stage of the project is to explore distributed media processing.

The long-term direction is to move from:

```text
                Single Worker
                     │
                  FFmpeg
                     │
                  MinIO
```

toward:

```text
                       Job Queue
                           │
             ┌─────────────┼─────────────┐
             ▼             ▼             ▼
          Worker A      Worker B      Worker C
             │             │             │
            GPU           GPU           GPU
             │             │             │
             └─────────────┼─────────────┘
                           ▼
                         MinIO
```

This will eventually allow StreamY to process multiple videos and/or independent media-processing jobs across multiple machines.

Before introducing that complexity, the single-machine pipeline provides a measured baseline against which distributed implementations can be compared.

---

# Goals

The project is being developed around a few core goals:

* Understand video encoding and HLS processing
* Build a production-style backend in Go
* Understand where bottlenecks occur in media workloads
* Benchmark before and after optimizations
* Explore concurrency and parallelism
* Build a distributed media-processing architecture
* Eventually explore caching, CDN/edge delivery, scheduling, and fault tolerance

The focus is not on building a complete end-to-end video platform immediately. The primary focus is **understanding and building the media-processing pipeline itself**, then scaling it from a single machine into a distributed system. 
