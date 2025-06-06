<h1 align="left">
  SimpHttp - A minimalist HTTP/HTTPS-aware domain probe
</h1>

<p align="left">
  <a href="https://opensource.org/licenses/MIT"><img src="https://img.shields.io/badge/license-MIT-_red.svg"></a>
  <a href="https://github.com/thd3r/SimpHttp/releases"><img src="https://img.shields.io/github/release/thd3r/SimpHttp.svg"></a>
  <a href="https://x.com/thd3r"><img src="https://img.shields.io/twitter/follow/thd3r.svg?logo=twitter"></a>
  <a href="https://github.com/thd3r/SimpHttp/issues"><img src="https://img.shields.io/badge/contributions-welcome-brightgreen.svg?style=flat"></a>

</p>

```sh
╔═╗┬┌┬┐┌─┐╦ ╦┌┬┐┌┬┐┌─┐
╚═╗││││├─┘╠═╣ │  │ ├─┘
╚═╝┴┴ ┴┴  ╩ ╩ ┴  ┴ ┴
          v0.1.2 latest
```

**SimpHttp** is a high-performance, lightweight HTTP scanner designed to rapidly test multiple hosts and ports using concurrent HTTP/HTTPS requests. It's ideal for developers, security researchers, and sysadmins who need to quickly verify service availability across large target lists.

## ✨ Features

* ✅ Fast TCP connectivity check before HTTP request
* 🌐 Supports both HTTP and HTTPS schemes automatically
* 📋 Reads targets from file, stdin, or CLI input
* ⚡ Concurrent scanning with user-defined thread count
* ⏱ Customizable timeout per request
* 🔎 Verbose logging for inspection and debugging
* 📦 Outputs structured results including:
  
  * Protocol used
  * HTTP status code
  * Response size
  * Redirect destination
  * Error message (if any)

## Use Cases

* Port and service discovery
* Web availability monitoring
* Mass HTTP banner grabbing

## Installation

```sh
go install -v github.com/thd3r/SimpHttp/cmd/simphttp@latest
```

## Flags

| Flag	   | Description | Example |
|----------|-------------|---------------------------------- |
| -targets | Single target, file path, or stdin | hosts.txt or example.com |
| -threads | Number of concurrent workers | 50 |
| -timeout | Timeout per request (in seconds) | 10 |
| -verbose | Enable verbose logging | -verbose |
| -version | Show SimpHttp version | -version |

## Usage

### Read from stdin

```sh
echo example.com | simphttp
```

```sh
echo https://example.com | simphttp
```

```sh
cat targets.txt | simphttp
```

### Or

```sh
simphttp -targets example.com
```

```sh
simphttp -targets https://example.com
```

```sh
simphttp -targets targets.txt
```

---
> [!TIP]
> **SimpHttp** automatically generates a report and saves it to a temporary folder.
---

## How It Works

### 1. Target Parsing
  Inputs can be a single host, list of hosts (from a file or stdin), or combination.

### 2. Connectivity Check
  Performs a low-level TCP dial to each `target:port` combination to ensure it's reachable.

### 3. HTTP Request
  Sends a GET request using a custom `http.Client` with timeouts and transport tuning.

### 4. Result Handling
  Response status, size, and redirect (if any) are extracted and printed. Errors are captured and reported per `host:port`

## Contributing

Contributions are welcome! Feel free to submit a pull request or open an issue to report bugs or suggest enhancements.
