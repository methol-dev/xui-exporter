package parse

import (
	"testing"
)

func TestParseSubscription_Success(t *testing.T) {
	html := `<html>
<body>
<template id="subscription-data"
  data-sid="uk2jf33cdnzjn2dg"
  data-downloadbyte="6150124543"
  data-uploadbyte="267143927"
  data-totalbyte="536870912000"
  data-expire="1769184000"
></template>
</body>
</html>`

	result, err := ParseSubscription([]byte(html))
	if err != nil {
		t.Fatalf("Expected success, got error: %v", err)
	}

	if result.SID != "uk2jf33cdnzjn2dg" {
		t.Errorf("Expected SID 'uk2jf33cdnzjn2dg', got '%s'", result.SID)
	}

	if result.DownloadByte != 6150124543 {
		t.Errorf("Expected DownloadByte 6150124543, got %d", result.DownloadByte)
	}

	if result.UploadByte != 267143927 {
		t.Errorf("Expected UploadByte 267143927, got %d", result.UploadByte)
	}

	if result.TotalByte != 536870912000 {
		t.Errorf("Expected TotalByte 536870912000, got %d", result.TotalByte)
	}

	if result.Expire != 1769184000 {
		t.Errorf("Expected Expire 1769184000, got %d", result.Expire)
	}
}

func TestParseSubscription_MissingTemplate(t *testing.T) {
	html := `<html><body><div>No template here</div></body></html>`

	_, err := ParseSubscription([]byte(html))
	if err == nil {
		t.Fatal("Expected error for missing template, got nil")
	}
}

func TestParseSubscription_MissingSID(t *testing.T) {
	html := `<html>
<body>
<template id="subscription-data"
  data-downloadbyte="6150124543"
  data-uploadbyte="267143927"
  data-totalbyte="536870912000"
  data-expire="1769184000"
></template>
</body>
</html>`

	_, err := ParseSubscription([]byte(html))
	if err == nil {
		t.Fatal("Expected error for missing sid, got nil")
	}
}

func TestParseSubscription_InvalidDownloadByte(t *testing.T) {
	html := `<html>
<body>
<template id="subscription-data"
  data-sid="test123"
  data-downloadbyte="not-a-number"
  data-uploadbyte="267143927"
  data-totalbyte="536870912000"
  data-expire="1769184000"
></template>
</body>
</html>`

	_, err := ParseSubscription([]byte(html))
	if err == nil {
		t.Fatal("Expected error for invalid downloadbyte, got nil")
	}
}

func TestParseSubscription_ZeroQuota(t *testing.T) {
	html := `<html>
<body>
<template id="subscription-data"
  data-sid="test123"
  data-downloadbyte="6150124543"
  data-uploadbyte="267143927"
  data-totalbyte="0"
  data-expire="1769184000"
></template>
</body>
</html>`

	_, err := ParseSubscription([]byte(html))
	if err == nil {
		t.Fatal("Expected error for zero quota, got nil")
	}
}

func TestParseSubscription_NegativeExpire(t *testing.T) {
	html := `<html>
<body>
<template id="subscription-data"
  data-sid="test123"
  data-downloadbyte="6150124543"
  data-uploadbyte="267143927"
  data-totalbyte="536870912000"
  data-expire="-1"
></template>
</body>
</html>`

	_, err := ParseSubscription([]byte(html))
	if err == nil {
		t.Fatal("Expected error for negative expire, got nil")
	}
}

func TestParseSubscription_SubPageData(t *testing.T) {
	html := `<!doctype html>
<html lang="en">
  <head>
    <title>Subscription</title>
    <script type="module" crossorigin src="/sub/assets/subpage.js"></script>
    <script>window.X_UI_BASE_PATH="/sub/";window.__SUB_PAGE_DATA__={"announce":"","download":"108.78GB","downloadByte":116799356178,"emails":["user"],"enabled":true,"expire":1795104000,"links":["vless://example\u0026type=tcp"],"sId":"uk2jf33cdnzjn2dg","total":"500.00GB","totalByte":536870912000,"upload":"5.31GB","uploadByte":5702004572};</script>
  </head>
  <body>
    <div id="app"></div>
  </body>
</html>`

	result, err := ParseSubscription([]byte(html))
	if err != nil {
		t.Fatalf("Expected success, got error: %v", err)
	}

	if result.SID != "uk2jf33cdnzjn2dg" {
		t.Errorf("Expected SID 'uk2jf33cdnzjn2dg', got '%s'", result.SID)
	}
	if result.DownloadByte != 116799356178 {
		t.Errorf("Expected DownloadByte 116799356178, got %d", result.DownloadByte)
	}
	if result.UploadByte != 5702004572 {
		t.Errorf("Expected UploadByte 5702004572, got %d", result.UploadByte)
	}
	if result.TotalByte != 536870912000 {
		t.Errorf("Expected TotalByte 536870912000, got %d", result.TotalByte)
	}
	if result.Expire != 1795104000 {
		t.Errorf("Expected Expire 1795104000, got %d", result.Expire)
	}
}

func TestParseSubscription_SubPageDataSpacedAssignment(t *testing.T) {
	html := `<html><head><script>
window.__SUB_PAGE_DATA__ = {"sId":"spaced","downloadByte":1,"uploadByte":2,"totalByte":3,"expire":4};
</script></head></html>`

	result, err := ParseSubscription([]byte(html))
	if err != nil {
		t.Fatalf("Expected success, got error: %v", err)
	}
	if result.SID != "spaced" || result.DownloadByte != 1 || result.UploadByte != 2 || result.TotalByte != 3 || result.Expire != 4 {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func TestParseSubscription_SubPageDataMissingSID(t *testing.T) {
	html := `<html><head><script>window.__SUB_PAGE_DATA__={"downloadByte":1,"uploadByte":2,"totalByte":3,"expire":4};</script></head></html>`

	_, err := ParseSubscription([]byte(html))
	if err == nil {
		t.Fatal("Expected error for missing sId, got nil")
	}
}

func TestParseSubscription_SubPageDataZeroQuota(t *testing.T) {
	html := `<html><head><script>window.__SUB_PAGE_DATA__={"sId":"test123","downloadByte":1,"uploadByte":2,"totalByte":0,"expire":4};</script></head></html>`

	_, err := ParseSubscription([]byte(html))
	if err == nil {
		t.Fatal("Expected error for zero quota, got nil")
	}
}

func TestParseSubscription_SubPageDataInvalidJSON(t *testing.T) {
	html := `<html><head><script>window.__SUB_PAGE_DATA__={not-json};</script></head></html>`

	_, err := ParseSubscription([]byte(html))
	if err == nil {
		t.Fatal("Expected error for invalid JSON, got nil")
	}
}

func TestParseSubscription_NegativeDownload(t *testing.T) {
	html := `<html>
<body>
<template id="subscription-data"
  data-sid="test123"
  data-downloadbyte="-100"
  data-uploadbyte="267143927"
  data-totalbyte="536870912000"
  data-expire="1769184000"
></template>
</body>
</html>`

	_, err := ParseSubscription([]byte(html))
	if err == nil {
		t.Fatal("Expected error for negative download, got nil")
	}
}
