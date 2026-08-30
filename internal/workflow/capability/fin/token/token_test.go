package token

import (
	"testing"

	"github.com/go-playground/assert/v2"
)

func TestGenerateProducesHighEntropyDistinctTokens(t *testing.T) {
	first, err := Generate()
	if err != nil {
		t.Fatal(err)
	}
	second, err := Generate()
	if err != nil {
		t.Fatal(err)
	}
	if first == second {
		t.Fatal("expected two generated tokens to differ")
	}
	assert.Equal(t, len(first), tokenBytes*2) // hex-encoded
}

func TestHashIsDeterministicAndDistinct(t *testing.T) {
	hashA := Hash("token-a")
	hashAAgain := Hash("token-a")
	hashB := Hash("token-b")

	assert.Equal(t, hashA, hashAAgain)
	if hashA == hashB {
		t.Fatal("expected different tokens to hash differently")
	}
	if hashA == "token-a" {
		t.Fatal("expected Hash to not return the plaintext token")
	}
}

func TestEqual(t *testing.T) {
	assert.Equal(t, Equal("secret", "secret"), true)
	assert.Equal(t, Equal("secret", "different"), false)
	assert.Equal(t, Equal("", ""), true)
}
