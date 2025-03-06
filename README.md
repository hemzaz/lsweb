# lsweb

`lsweb` is a command-line tool designed to list and download files from websites. It acts as an `ls` command for websites, providing a quick way to view and fetch downloadable content from a given URL.

## Features

- List downloadable links from a website.
- Download files directly to the current working directory.
- Supports simultaneous and sequential downloading.
- Dynamic and colorful progress bar for each download.
- Automatically extracts links from JSON, XML, and HTML content.
- Special flag for fetching GitHub release assets.

## Installation

### Prerequisites

- Go (version 1.22 or higher)
- `goreleaser` for building

### Building from Source

1. Clone the repository:
   ```bash
   git clone https://github.com/hemzaz/lsweb.git
   ```

2. Navigate to the project directory:
   ```bash
   cd lsweb
   ```

3. Build using `goreleaser`:
   ```bash
   make build
   ```

## Usage

```bash
lsweb [flags]
```

### Flags

- `-u`: URL to fetch links from
- `-f`: File to fetch links from
- `-o`: Output format (json, txt, num, html)
- `-filter`: Regex to filter links
- `-limit`: Limit the number of links to fetch
- `-ic`: Ignore certificate errors
- `-gh`: Fetch GitHub releases
- `-download`: Download the files
- `-list`: List the links (default: true)
- `-sim`: Download files simultaneously
- `-max-concurrent`: Maximum number of concurrent downloads (default: 5)
- `-overwrite`: Overwrite existing files when downloading
- `-timeout`: Timeout in seconds for HTTP requests (default: 60)
- `-version`: Show version information

### Examples

1. List downloadable links from a website:
   ```bash
   lsweb -u https://example.com
   ```

2. Download files from a website:
   ```bash
   lsweb -download -u https://example.com
   ```

3. Download files simultaneously:
   ```bash
   lsweb -download -sim -u https://example.com
   ```

4. List GitHub release assets:
   ```bash
   lsweb -gh -u https://github.com/telegramdesktop/tdesktop/
   ```

https://github.com/hemzaz/lsweb/assets/1830915/e621f153-31b8-48e9-babd-ca174e1cd3ca


## Contributing

Contributions are welcome! Please feel free to submit a pull request or open an issue.

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

---
Authored by: **hemzaz the frogodile** 🐸🐊
