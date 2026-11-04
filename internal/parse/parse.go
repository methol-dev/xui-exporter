package parse

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"golang.org/x/net/html"
)

// ParsedSubscription represents the raw data extracted from the subscription HTML
type ParsedSubscription struct {
	SID          string
	DownloadByte int64
	UploadByte   int64
	TotalByte    int64
	Expire       int64
}

// subPageData is the JSON object assigned to window.__SUB_PAGE_DATA__
// on current 3x-ui subscription pages.
type subPageData struct {
	SID          string `json:"sId"`
	DownloadByte int64  `json:"downloadByte"`
	UploadByte   int64  `json:"uploadByte"`
	TotalByte    int64  `json:"totalByte"`
	Expire       int64  `json:"expire"`
}

// ParseSubscription parses the subscription HTML.
// Current 3x-ui pages embed JSON in window.__SUB_PAGE_DATA__.
// Older pages expose the same numbers on template#subscription-data.
func ParseSubscription(htmlBytes []byte) (ParsedSubscription, error) {
	doc, err := html.Parse(strings.NewReader(string(htmlBytes)))
	if err != nil {
		return ParsedSubscription{}, fmt.Errorf("failed to parse HTML: %w", err)
	}

	if script := findSubPageDataScript(doc); script != "" {
		return parseSubPageData(script)
	}

	templateNode := findTemplateNode(doc)
	if templateNode == nil {
		return ParsedSubscription{}, fmt.Errorf("subscription data not found: expected window.__SUB_PAGE_DATA__ or template#subscription-data")
	}

	attrs := extractAttributes(templateNode)

	// Parse required fields
	sid, ok := attrs["data-sid"]
	if !ok || sid == "" {
		return ParsedSubscription{}, fmt.Errorf("data-sid attribute missing or empty")
	}

	downloadByte, err := parsePositiveInt64(attrs, "data-downloadbyte")
	if err != nil {
		return ParsedSubscription{}, fmt.Errorf("data-downloadbyte: %w", err)
	}

	uploadByte, err := parsePositiveInt64(attrs, "data-uploadbyte")
	if err != nil {
		return ParsedSubscription{}, fmt.Errorf("data-uploadbyte: %w", err)
	}

	totalByte, err := parsePositiveInt64(attrs, "data-totalbyte")
	if err != nil {
		return ParsedSubscription{}, fmt.Errorf("data-totalbyte: %w", err)
	}

	expire, err := parsePositiveInt64(attrs, "data-expire")
	if err != nil {
		return ParsedSubscription{}, fmt.Errorf("data-expire: %w", err)
	}

	// Validate field values according to requirements
	if totalByte <= 0 {
		return ParsedSubscription{}, fmt.Errorf("data-totalbyte must be positive (got %d)", totalByte)
	}

	if expire <= 0 {
		return ParsedSubscription{}, fmt.Errorf("data-expire must be positive (got %d)", expire)
	}

	if downloadByte < 0 {
		return ParsedSubscription{}, fmt.Errorf("data-downloadbyte must be non-negative (got %d)", downloadByte)
	}

	if uploadByte < 0 {
		return ParsedSubscription{}, fmt.Errorf("data-uploadbyte must be non-negative (got %d)", uploadByte)
	}

	return ParsedSubscription{
		SID:          sid,
		DownloadByte: downloadByte,
		UploadByte:   uploadByte,
		TotalByte:    totalByte,
		Expire:       expire,
	}, nil
}

func parseSubPageData(script string) (ParsedSubscription, error) {
	payload, err := subPageDataJSON(script)
	if err != nil {
		return ParsedSubscription{}, err
	}

	var data subPageData
	dec := json.NewDecoder(strings.NewReader(payload))
	if err := dec.Decode(&data); err != nil {
		return ParsedSubscription{}, fmt.Errorf("window.__SUB_PAGE_DATA__: %w", err)
	}

	if data.SID == "" {
		return ParsedSubscription{}, fmt.Errorf("sId missing or empty")
	}
	if data.TotalByte <= 0 {
		return ParsedSubscription{}, fmt.Errorf("totalByte must be positive (got %d)", data.TotalByte)
	}
	if data.Expire <= 0 {
		return ParsedSubscription{}, fmt.Errorf("expire must be positive (got %d)", data.Expire)
	}
	if data.DownloadByte < 0 {
		return ParsedSubscription{}, fmt.Errorf("downloadByte must be non-negative (got %d)", data.DownloadByte)
	}
	if data.UploadByte < 0 {
		return ParsedSubscription{}, fmt.Errorf("uploadByte must be non-negative (got %d)", data.UploadByte)
	}

	return ParsedSubscription{
		SID:          data.SID,
		DownloadByte: data.DownloadByte,
		UploadByte:   data.UploadByte,
		TotalByte:    data.TotalByte,
		Expire:       data.Expire,
	}, nil
}

// subPageDataJSON returns the object literal assigned to window.__SUB_PAGE_DATA__.
func subPageDataJSON(script string) (string, error) {
	const key = "__SUB_PAGE_DATA__"
	idx := strings.Index(script, key)
	if idx < 0 {
		return "", fmt.Errorf("window.__SUB_PAGE_DATA__ not found")
	}

	rest := strings.TrimLeft(script[idx+len(key):], " \t\r\n")
	if !strings.HasPrefix(rest, "=") {
		return "", fmt.Errorf("window.__SUB_PAGE_DATA__ assignment not found")
	}
	rest = strings.TrimLeft(rest[1:], " \t\r\n")
	if rest == "" || rest[0] != '{' {
		return "", fmt.Errorf("window.__SUB_PAGE_DATA__ is not a JSON object")
	}
	return rest, nil
}

// findSubPageDataScript returns the inline script that assigns window.__SUB_PAGE_DATA__.
func findSubPageDataScript(n *html.Node) string {
	if n.Type == html.ElementNode && n.Data == "script" {
		text := scriptText(n)
		if strings.Contains(text, "__SUB_PAGE_DATA__") {
			return text
		}
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if result := findSubPageDataScript(c); result != "" {
			return result
		}
	}
	return ""
}

func scriptText(n *html.Node) string {
	var b strings.Builder
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.TextNode {
			b.WriteString(c.Data)
		}
	}
	return b.String()
}

// findTemplateNode recursively searches for template#subscription-data element
func findTemplateNode(n *html.Node) *html.Node {
	if n.Type == html.ElementNode && n.Data == "template" {
		// Check if it has id="subscription-data"
		for _, attr := range n.Attr {
			if attr.Key == "id" && attr.Val == "subscription-data" {
				return n
			}
		}
	}

	// Recursively search children
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if result := findTemplateNode(c); result != nil {
			return result
		}
	}

	return nil
}

// extractAttributes extracts all attributes from a node into a map
func extractAttributes(n *html.Node) map[string]string {
	attrs := make(map[string]string)
	for _, attr := range n.Attr {
		attrs[attr.Key] = attr.Val
	}
	return attrs
}

// parsePositiveInt64 parses an int64 from the attributes map
func parsePositiveInt64(attrs map[string]string, key string) (int64, error) {
	val, ok := attrs[key]
	if !ok {
		return 0, fmt.Errorf("attribute %s not found", key)
	}

	parsed, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid integer value: %w", err)
	}

	return parsed, nil
}
