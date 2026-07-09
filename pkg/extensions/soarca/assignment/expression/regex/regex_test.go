package regex

import (
	"testing"

	"github.com/go-playground/assert/v2"
)

const httpResponse = `HTTP/1.1 200 OK
Content-Type: application/json
Authorization: Bearer abc.def.ghi
X-Request-Id: 7f3a2e1b
Set-Cookie: session=alpha; Path=/
Set-Cookie: tracking=beta; Path=/

{"client_ip": "192.168.1.42", "users": ["alice@example.test", "bob@example.test"]}
`

func TestName(t *testing.T) {
	regex := New()
	assert.Equal(t, regex.GetEngineName(), "regex")
}

func TestFullMatchNoCaptureGroup(t *testing.T) {
	regex := New()
	result, err := regex.Execute(httpResponse, `\d+\.\d+\.\d+\.\d+`)
	assert.Equal(t, err, nil)
	assert.Equal(t, result, "192.168.1.42")
}

func TestSingleCaptureGroup(t *testing.T) {
	regex := New()
	result, err := regex.Execute(httpResponse, `Bearer (\S+)`)
	assert.Equal(t, err, nil)
	assert.Equal(t, result, "abc.def.ghi")
}

func TestStatusCode(t *testing.T) {
	regex := New()
	result, err := regex.Execute(httpResponse, `HTTP/\d\.\d (\d+)`)
	assert.Equal(t, err, nil)
	assert.Equal(t, result, "200")
}

func TestMultipleMatchesNoCaptureGroup(t *testing.T) {
	regex := New()
	result, err := regex.Execute(httpResponse, `[\w.]+@[\w.]+`)
	assert.Equal(t, err, nil)
	assert.Equal(t, result, "alice@example.test\nbob@example.test")
}

func TestMultipleMatchesWithCaptureGroup(t *testing.T) {
	regex := New()
	result, err := regex.Execute(httpResponse, `Set-Cookie: (\w+)=`)
	assert.Equal(t, err, nil)
	assert.Equal(t, result, "session\ntracking")
}

func TestSecondCaptureGroupIgnored(t *testing.T) {
	regex := New()
	result, err := regex.Execute("user=alice role=admin", `user=(\w+) role=(\w+)`)
	assert.Equal(t, err, nil)
	assert.Equal(t, result, "alice")
}

func TestNoMatch(t *testing.T) {
	regex := New()
	result, err := regex.Execute(httpResponse, `nonexistent-header: (\S+)`)
	assert.Equal(t, err, nil)
	assert.Equal(t, result, "")
}

func TestEmptySource(t *testing.T) {
	regex := New()
	result, err := regex.Execute("", `\d+`)
	assert.Equal(t, err, nil)
	assert.Equal(t, result, "")
}

func TestCaseInsensitiveFlag(t *testing.T) {
	regex := New()
	result, err := regex.Execute("Hello HELLO hello", `(?i)hello`)
	assert.Equal(t, err, nil)
	assert.Equal(t, result, "Hello\nHELLO\nhello")
}

func TestUnicodeLetters(t *testing.T) {
	regex := New()
	result, err := regex.Execute("héllo wörld", `\p{L}+`)
	assert.Equal(t, err, nil)
	assert.Equal(t, result, "héllo\nwörld")
}

func TestEscapedSpecialChars(t *testing.T) {
	regex := New()
	result, err := regex.Execute("price is $42.50 today", `\$(\d+\.\d+)`)
	assert.Equal(t, err, nil)
	assert.Equal(t, result, "42.50")
}

func TestInvalidPattern(t *testing.T) {
	regex := New()
	result, err := regex.Execute(httpResponse, `(unclosed`)
	assert.NotEqual(t, err, nil)
	assert.Equal(t, result, "")
}

func TestInvalidGroupReference(t *testing.T) {
	regex := New()
	result, err := regex.Execute(httpResponse, `*invalid`)
	assert.NotEqual(t, err, nil)
	assert.Equal(t, result, "")
}

func TestJsonFieldExtraction(t *testing.T) {
	regex := New()
	result, err := regex.Execute(httpResponse, `"client_ip":\s*"([^"]+)"`)
	assert.Equal(t, err, nil)
	assert.Equal(t, result, "192.168.1.42")
}

func TestRequestIdExtraction(t *testing.T) {
	regex := New()
	result, err := regex.Execute(httpResponse, `X-Request-Id: (\S+)`)
	assert.Equal(t, err, nil)
	assert.Equal(t, result, "7f3a2e1b")
}
