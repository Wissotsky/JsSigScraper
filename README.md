# JavaScript Signature Scanner

This command-line tool written in Go allows you to detect JavaScript signatures on webpages given a URL. It fetches the HTML of a webpage, identifies and downloads JavaScript files, and checks them against predefined regular expression "signatures" specified in a TOML configuration file to identify JavaScript signatures used on the page.

## Introduction

This tool provides insight into the JavaScript signatures employed by a webpage. It can:

- Accept a URL as input.
- Retrieve the HTML content of the webpage.
- Identify and download JavaScript files linked within the HTML.
- Analyze JavaScript files for JavaScript signatures.
- Output the detected JavaScript signature names as newline-separated text for easy chaining with other POSIX scripts.
- Allow users to specify and customize regex signatures using a TOML configuration file.

## Installation

1. Clone this repository.
2. Make sure you have Go (Golang) installed.
3. Navigate to the project directory.
4. Run the tool with the following command:
   
   ```shell
   go run JsSigScraper.go -url <URL> -config <config-file> [-keepjs] [-useragent <user-agent>]
   ```

   Example:

   ```shell
   go run JsSigScraper.go -url https://example.com -config example-config.toml -keepjs -useragent Mozilla/5.0
   ```

## Configuration

The TOML configuration file (`example-config.toml`) includes user-defined regex signatures for JavaScript signatures. Here's an example structure:

```toml
[[signatures]]
name = "Signature Name 1"
regex = "your-regex-here-1"

[[signatures]]
name = "Signature Name 2"
regex = "your-regex-here-2"

# Add more signatures as needed
```

## Output

The output consists of detected JavaScript signature names separated by newlines:

```plaintext
Signature Name 1
Signature Name 2
```

## External Dependencies

- HTML parsing library (e.g., `golang.org/x/net/html`)
- TOML parsing library (e.g., `github.com/BurntSushi/toml`)
