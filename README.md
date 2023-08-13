
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

- Go (version 1.16 or higher)
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
lsweb [flags] <url>
```

### Flags

- `-L, --list`: List downloadable links from the provided URL. This is also the default action if no flags are provided.
- `-D, --download`: Download the files. By default, files are downloaded sequentially.
- `-S, --sim`: Download files simultaneously.
- `-O, --output`: Specify the output format. Available formats: `json`, `txt`, `num`, `html`. Default is `txt`.
- `-F, --file`: Specify a file to write the output.
- `--gh`: Accept a GitHub URL and list all release assets.

### Examples

1. List downloadable links from a website:
   ```bash
   lsweb -L --url https://example.com
   ```

2. Download files from a website:
   ```bash
   lsweb -D --url https://example.com
   ```

3. Download files simultaneously:
   ```bash
   lsweb -D -S --url https://example.com
   ```

4. List GitHub release assets:
   ```bash
   lsweb --gh --url https://github.com/telegramdesktop/tdesktop/releases/tag/v4.8.10
   ```

## Contributing

Contributions are welcome! Please feel free to submit a pull request or open an issue.

## TODO  

1. add --limit flag
2. add --filter flag (use regex to filter files)
3. add --ic flag (ignore certificate)

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

---
Authored by: **hemzaz the frogodile** 🐸🐊