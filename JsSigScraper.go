package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/BurntSushi/toml"
	"golang.org/x/net/html"
)

type JavaScriptSignature struct {
	Name  string
	Regex string
}

type Configuration struct {
	Signatures []JavaScriptSignature
}

var (
	defaultUserAgent = "Mozilla/5.0"
)

func main() {
	urlPtr := flag.String("url", "", "URL to scan")
	configFilePtr := flag.String("config", "", "Path to configuration file")
	keepJS := flag.Bool("keepjs", false, "Keep JavaScript files")
	userAgent := flag.String("useragent", defaultUserAgent, "User agent to use")
	flag.Parse()

	if *urlPtr == "" || *configFilePtr == "" {
		fmt.Println("Usage: go run JsSigScraper.go -url <URL> -config <config-file> [-keepjs] [-useragent <user-agent>]")
		return
	}

	config, err := readConfig(*configFilePtr)
	if err != nil {
		log.Fatalf("Error reading configuration: %v", err)
	}

	htmlContent, err := fetchHTML(*urlPtr, *userAgent)
	if err != nil {
		log.Fatalf("Error fetching HTML: %v", err)
	}

	javascriptURLs, err := extractJSScriptURLs(htmlContent, *urlPtr)
	if err != nil {
		log.Fatalf("Error extracting JavaScript URLs: %v", err)
	}

	var wg sync.WaitGroup

	// Download JavaScript files and check for JavaScript signatures
	for _, jsURL := range javascriptURLs {
		wg.Add(1)
		go func(jsURL string) {
			defer wg.Done()

			content, err := fetchJavascript(jsURL, *userAgent)
			if err != nil {
				log.Printf("Error downloading JavaScript from %s: %v\n", jsURL, err)
				return
			}

			for _, signature := range config.Signatures {
				if matched, _ := regexp.MatchString(signature.Regex, content); matched {
					fmt.Println(signature.Name)
					break
				}
			}

			// Save the JavaScript content to a file if -keepjs is specified
			if *keepJS {
				if err := saveJavascript(jsURL, content); err != nil {
					log.Printf("Error saving JavaScript from %s: %v\n", jsURL, err)
				}
			}
		}(jsURL)
	}

	// Wait for all goroutines to finish
	wg.Wait()
}

func readConfig(configFile string) (Configuration, error) {
	var config Configuration
	if _, err := toml.DecodeFile(configFile, &config); err != nil {
		return Configuration{}, err
	}
	return config, nil
}

func fetchHTML(url, userAgent string) (string, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	request.Header.Set("User-Agent", userAgent)

	response, err := client.Do(request)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP request failed with status code: %d", response.StatusCode)
	}

	htmlBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return "", err
	}

	return string(htmlBytes), nil
}

// absolute monstrosity
func extractJSScriptURLs(htmlString, baseURL string) ([]string, error) {
	var jsURLs []string

	base, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}

	tokenizer := html.NewTokenizer(strings.NewReader(htmlString))

	for {
		tokenType := tokenizer.Next()
		switch tokenType {
		case html.ErrorToken:
			return jsURLs, nil // End of the document
		case html.StartTagToken, html.SelfClosingTagToken:
			token := tokenizer.Token()
			if token.Data == "script" {
				for _, attr := range token.Attr {
					if attr.Key == "src" {
						scriptURL, err := url.Parse(attr.Val)
						if err != nil {
							return nil, err
						}
						absoluteURL := base.ResolveReference(scriptURL).String()
						jsURLs = append(jsURLs, absoluteURL)
					}
				}
			}
		}
	}
}

func fetchJavascript(jsURL, userAgent string) (string, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	request, err := http.NewRequest("GET", jsURL, nil)
	if err != nil {
		return "", err
	}
	request.Header.Set("User-Agent", userAgent)

	response, err := client.Do(request)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP request failed with status code: %d", response.StatusCode)
	}

	jsBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return "", err
	}

	return string(jsBytes), nil
}

func saveJavascript(jsURL, content string) error {
	domain, filename, err := getFilename(jsURL)
	if err != nil {
		return err
	}

	// Create folder for domain if it doesn't exist
	if err := os.MkdirAll(domain, 0755); err != nil {
		return err
	}

	filename = filepath.Join(domain, filename)
	if err := os.WriteFile(filename, []byte(content), 0644); err != nil {
		return err
	}

	return nil
}

func getFilename(jsURL string) (string, string, error) {
	url, err := url.Parse(jsURL)
	if err != nil {
		return "", "", err
	}
	domain := url.Hostname()
	filename := filepath.Base(url.Path)
	return domain, filename, nil
}
