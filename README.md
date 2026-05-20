# homelab

A lightweight, standard-library-only Go media server and download manager. This project is designed to facilitate streaming local video files and managing remote file downloads via a simple HTTP interface.

## Features

- **Video Streaming**: Supports chunked streaming of various video formats (`.mp4`, `.webm`, `.avi`, `.mkv`, `.mov`) using HTTP `Range` headers.
- **File Management**: List files available for streaming through an easy-to-use endpoint.
- **Download Manager**: Concurrently handle and track file downloads from remote URLs.
- **Zero Third-Party Dependencies**: Built entirely using the Go Standard Library.
- **Cross-Platform**: Easy to build for both Linux and Windows.

## Project Structure

- `cmd/main.go`: Application entry point and route configuration.
- `internals/handlers/streaming/`: Logic for file listing and chunked video delivery.
- `internals/handlers/downloads/`: Download manager and progress tracking implementation.
- `internals/utils/`: Helper functions.
- `bin/`: Compiled binaries.

## API Endpoints

- `GET /`: Lists all streamable files.
- `GET /file?name=<filename>`: Streams the specified file in chunks.
- `POST /download`: Starts a new download (Expects a JSON body with `url` and `filename`).
- `GET /try`: Triggers a sample download for testing.

## Getting Started

### Prerequisites

- Go (version 1.25 or higher recommended)
- `make` (optional, for build automation)

### Installation

1. Clone the repository:
   ```bash
   git clone <repo-url>
   cd homelab
   ```

2. Run the application:
   ```bash
   make run
   ```
   *The server will start on [http://localhost:4080](http://localhost:4080).*

### Building

- To build for your current system:
  ```bash
  make build
  ```
- To cross-compile for Windows:
  ```bash
  make windows
  ```
  *Binaries are saved in the `bin/` directory.*

## License

[MIT](LICENSE) (or specify your license)
