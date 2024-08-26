# qbrdt

`qbrdt` is a Go-based wrapper around the qBittorrent API that integrates Real-Debrid, allowing seamless use of qBittorrent with popular automation tools like Sonarr, Radarr, and others.

## Table of Contents

- [Features](#features)
- [Installation](#installation)
- [Usage](#usage)
- [Configuration](#configuration)
- [Contributing](#contributing)
- [License](#license)

## Features

- **Real-Debrid Integration**: Automatically handle torrents through Real-Debrid, ensuring faster and more reliable downloads.
- **qBittorrent Wrapper**: Acts as an intermediary between qBittorrent and automation tools.
- **Compatible with Sonarr/Radarr**: Easily integrate with tools that manage and automate your media library.
- **Lightweight and Fast**: Written in Go, ensuring optimal performance and low resource consumption.

## Installation

### Prerequisites

- Go 1.18+ installed on your system.
- A running instance of qBittorrent with WebUI enabled.
- A Real-Debrid account with a valid API token.

### Clone the Repository

```bash
git clone https://github.com/yourusername/qbrdt.git
cd qbrdt
```

### Build the Binary

```bash
go build -o qbrdt cmd/qbrdt/main.go
```

## Usage

```bash
./qbrdt
```



## Contributing

Contributions are welcome! Please fork the repository and create a pull request with your changes. Ensure you follow the coding standards and include tests for any new features or bug fixes.

## License

This project is licensed under the MIT License. See the LICENSE file for more details.