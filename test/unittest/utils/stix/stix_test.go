package stix_test

import (
	"errors"
	"soarca/models/cacao"
	"soarca/utils/stix"
	"testing"

	"github.com/go-playground/assert/v2"
)

func TestStringEquals(t *testing.T) {

	stix := stix.New()

	var1 := cacao.Variable{Type: cacao.VariableTypeString}
	var1.Value = "a"
	var1.Name = "__var1__"
	vars := cacao.NewVariables(var1)

	result, err := stix.Evaluate("__var1__:value = a", vars)
	assert.Equal(t, err, nil)
	assert.Equal(t, result, true)
	result, err = stix.Evaluate("__var1__:value = b", vars)
	assert.Equal(t, result, false)
	assert.Equal(t, err, nil)

	result, err = stix.Evaluate("__var1__:value = 1", vars)
	assert.Equal(t, result, false)
	assert.Equal(t, err, nil)

	result, err = stix.Evaluate("__var1__:value > b", vars)
	assert.Equal(t, result, false)
	assert.Equal(t, err, nil)

	result, err = stix.Evaluate("__var1__:value < b", vars)
	assert.Equal(t, result, true)
	assert.Equal(t, err, nil)

	result, err = stix.Evaluate("__var1__:value <= b", vars)
	assert.Equal(t, result, true)
	assert.Equal(t, err, nil)

	result, err = stix.Evaluate("__var1__:value >= b", vars)
	assert.Equal(t, result, false)
	assert.Equal(t, err, nil)

	result, err = stix.Evaluate("a =  b", vars)
	assert.Equal(t, result, false)
	assert.Equal(t, err, errors.New("comparisons can only contain 3 parts as per STIX specification"))
	result, err = stix.Evaluate("a = b c", vars)
	assert.Equal(t, result, false)
	assert.Equal(t, err, errors.New("comparisons can only contain 3 parts as per STIX specification"))

}

func TestIntEquals(t *testing.T) {
	stix := stix.New()

	var1 := cacao.Variable{Type: cacao.VariableTypeLong}
	var1.Value = "1000"
	var1.Name = "__var1__"
	vars := cacao.NewVariables(var1)

	result, err := stix.Evaluate("__var1__:value = 1000", vars)
	assert.Equal(t, err, nil)
	assert.Equal(t, result, true)
	result, err = stix.Evaluate("__var1__:value = 9999", vars)
	assert.Equal(t, result, false)
	assert.Equal(t, err, nil)

	result, err = stix.Evaluate("__var1__:value = 10000", vars)
	assert.Equal(t, result, false)
	assert.Equal(t, err, nil)

	result, err = stix.Evaluate("__var1__:value > 999", vars)
	assert.Equal(t, result, true)
	assert.Equal(t, err, nil)

	result, err = stix.Evaluate("__var1__:value < 1001", vars)
	assert.Equal(t, result, true)
	assert.Equal(t, err, nil)

	result, err = stix.Evaluate("__var1__:value <= 1000", vars)
	assert.Equal(t, result, true)
	assert.Equal(t, err, nil)

	result, err = stix.Evaluate("__var1__:value >= 1000", vars)
	assert.Equal(t, result, true)
	assert.Equal(t, err, nil)

	result, err = stix.Evaluate("__var1__:value >= a", vars)
	assert.Equal(t, result, false)
	assert.NotEqual(t, err, nil)

	result, err = stix.Evaluate("a =  b", vars)
	assert.Equal(t, result, false)
	assert.Equal(t, err, errors.New("comparisons can only contain 3 parts as per STIX specification"))
	result, err = stix.Evaluate("a = b c", vars)
	assert.Equal(t, result, false)
	assert.Equal(t, err, errors.New("comparisons can only contain 3 parts as per STIX specification"))
}

func TestFloatEquals(t *testing.T) {
	stix := stix.New()

	var1 := cacao.Variable{Type: cacao.VariableTypeFloat}
	var1.Value = "1000.0"
	var1.Name = "__var1__"
	vars := cacao.NewVariables(var1)

	result, err := stix.Evaluate("__var1__:value = 1000.0", vars)
	assert.Equal(t, err, nil)
	assert.Equal(t, result, true)

	result, err = stix.Evaluate("__var1__:value = 1000.000000000000000000001", vars)
	assert.Equal(t, err, nil)
	assert.Equal(t, result, true)

	result, err = stix.Evaluate("__var1__:value = 1000.000001", vars)
	assert.Equal(t, err, nil)
	assert.Equal(t, result, false)

	result, err = stix.Evaluate("__var1__:value = 9999", vars)
	assert.Equal(t, result, false)
	assert.Equal(t, err, nil)

	result, err = stix.Evaluate("__var1__:value = 10000", vars)
	assert.Equal(t, result, false)
	assert.Equal(t, err, nil)

	result, err = stix.Evaluate("__var1__:value > 999", vars)
	assert.Equal(t, result, true)
	assert.Equal(t, err, nil)

	result, err = stix.Evaluate("__var1__:value < 1001", vars)
	assert.Equal(t, result, true)
	assert.Equal(t, err, nil)

	result, err = stix.Evaluate("__var1__:value <= 1000", vars)
	assert.Equal(t, result, true)
	assert.Equal(t, err, nil)

	result, err = stix.Evaluate("__var1__:value >= 1000", vars)
	assert.Equal(t, result, true)
	assert.Equal(t, err, nil)

	result, err = stix.Evaluate("__var1__:value >= a", vars)
	assert.Equal(t, result, false)
	assert.NotEqual(t, err, nil)

	result, err = stix.Evaluate("a =  b", vars)
	assert.Equal(t, result, false)
	assert.Equal(t, err, errors.New("comparisons can only contain 3 parts as per STIX specification"))
	result, err = stix.Evaluate("a = b c", vars)
	assert.Equal(t, result, false)
	assert.Equal(t, err, errors.New("comparisons can only contain 3 parts as per STIX specification"))
}

func TestIp4AddressEquals(t *testing.T) {
	stix := stix.New()
	var1 := cacao.Variable{Type: cacao.VariableTypeIpv4Address}
	var1.Value = "10.0.0.30"
	var1.Name = "__var1__"
	vars := cacao.NewVariables(var1)

	result, err := stix.Evaluate("__var1__:value = 10.0.0.30", vars)
	assert.Equal(t, err, nil)
	assert.Equal(t, result, true)

	result, err = stix.Evaluate("__var1__:value IN 10.0.0.0/8", vars)
	assert.Equal(t, err, nil)
	assert.Equal(t, result, true)

	result, err = stix.Evaluate("__var1__:value IN 10.30.0.0/16", vars)
	assert.Equal(t, err, nil)
	assert.Equal(t, result, false)

	result, err = stix.Evaluate("__var1__:value != 10.0.0.31", vars)
	assert.Equal(t, err, nil)
	assert.Equal(t, result, true)
}

func TestIp6AddressEquals(t *testing.T) {
	stix := stix.New()
	var1 := cacao.Variable{Type: cacao.VariableTypeIpv6Address}
	var1.Value = "2001:db8::1"
	var1.Name = "__var1__"
	vars := cacao.NewVariables(var1)

	result, err := stix.Evaluate("__var1__:value = 2001:db8::1", vars)
	assert.Equal(t, err, nil)
	assert.Equal(t, result, true)

	result, err = stix.Evaluate("__var1__:value IN 2001:db8::1/64", vars)
	assert.Equal(t, err, nil)
	assert.Equal(t, result, true)

	result, err = stix.Evaluate("__var1__:value IN 2001:db81::1/64", vars)
	assert.Equal(t, err, nil)
	assert.Equal(t, result, false)

	result, err = stix.Evaluate("__var1__:value != 2001:db8::2", vars)
	assert.Equal(t, err, nil)
	assert.Equal(t, result, true)
}

func TestMacAddressEquals(t *testing.T) {
	stix := stix.New()
	var1 := cacao.Variable{Type: cacao.VariableTypeMacAddress}
	var1.Value = "BC-24-11-00-00-01"
	var1.Name = "__var1__"
	vars := cacao.NewVariables(var1)

	result, err := stix.Evaluate("__var1__:value = BC-24-11-00-00-01", vars)
	assert.Equal(t, err, nil)
	assert.Equal(t, result, true)

	var2 := cacao.Variable{Type: cacao.VariableTypeMacAddress}
	var2.Value = "BC:24:11:00:00:01"
	var2.Name = "__var2__"
	vars2 := cacao.NewVariables(var2)

	result, err = stix.Evaluate("__var2__:value = BC:24:11:00:00:01", vars2)
	assert.Equal(t, err, nil)
	assert.Equal(t, result, true)

	// Mixed notations
	result, err = stix.Evaluate("__var1__:value = BC:24:11:00:00:01", vars)
	assert.Equal(t, err, nil)
	assert.Equal(t, result, true)

	result, err = stix.Evaluate("__var1__:value > BC:24:11:00:00:00", vars)
	assert.Equal(t, err, nil)
	assert.Equal(t, result, true)

	result, err = stix.Evaluate("__var1__:value < BC:24:11:00:00:02", vars)
	assert.Equal(t, err, nil)
	assert.Equal(t, result, true)
}

func TestHashEquals(t *testing.T) {
	stix := stix.New()
	md5 := cacao.Variable{Type: cacao.VariableTypeMd5Has}
	md5.Value = "d41d8cd98f00b204e9800998ecf8427e"
	md5.Name = "__md5__"

	sha1 := cacao.Variable{Type: cacao.VariableTypeMd5Has}
	sha1.Value = "da39a3ee5e6b4b0d3255bfef95601890afd80709"
	sha1.Name = "__sha1__"

	sha224 := cacao.Variable{Type: cacao.VariableTypeHash}
	sha224.Value = "d14a028c2a3a2bc9476102bb288234c415a2b01f828ea62ac5b3e42f"
	sha224.Name = "__sha224__"

	sha256 := cacao.Variable{Type: cacao.VariableTypeSha256}
	sha256.Value = "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
	sha256.Name = "__sha256__"

	sha384 := cacao.Variable{Type: cacao.VariableTypeHash}
	sha384.Value = "38b060a751ac96384cd9327eb1b1e36a21fdb71114be07434c0cc7bf63f6e1da274edebfe76f65fbd51ad2f14898b95b"
	sha384.Name = "__sha384__"

	sha512 := cacao.Variable{Type: cacao.VariableTypeHash}
	sha512.Value = "cf83e1357eefb8bdf1542850d66d8007d620e4050b5715dc83f4a921d36ce9ce47d0d13c5d85f2b0ff8318d2877eec2f63b931bd47417a81a538327af927da3e"
	sha512.Name = "__sha512__"

	vars := cacao.NewVariables(md5, sha1, sha224, sha256, sha384, sha512)

	result, err := stix.Evaluate("__md5__:value = d41d8cd98f00b204e9800998ecf8427e", vars)
	assert.Equal(t, err, nil)
	assert.Equal(t, result, true)

	result, err = stix.Evaluate("__sha1__:value = da39a3ee5e6b4b0d3255bfef95601890afd80709", vars)
	assert.Equal(t, err, nil)
	assert.Equal(t, result, true)

	result, err = stix.Evaluate("__sha224__:value = d14a028c2a3a2bc9476102bb288234c415a2b01f828ea62ac5b3e42f", vars)
	assert.Equal(t, err, nil)
	assert.Equal(t, result, true)

	result, err = stix.Evaluate("__sha256__:value = e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855", vars)
	assert.Equal(t, err, nil)
	assert.Equal(t, result, true)

	result, err = stix.Evaluate("__sha384__:value = 38b060a751ac96384cd9327eb1b1e36a21fdb71114be07434c0cc7bf63f6e1da274edebfe76f65fbd51ad2f14898b95b", vars)
	assert.Equal(t, err, nil)
	assert.Equal(t, result, true)

	result, err = stix.Evaluate("__sha512__:value = cf83e1357eefb8bdf1542850d66d8007d620e4050b5715dc83f4a921d36ce9ce47d0d13c5d85f2b0ff8318d2877eec2f63b931bd47417a81a538327af927da3e", vars)
	assert.Equal(t, err, nil)
	assert.Equal(t, result, true)

}

func TestUriEquals(t *testing.T) {
	stix := stix.New()

	var1 := cacao.Variable{Type: cacao.VariableTypeUri}
	var1.Value = "https://google.com"
	var1.Name = "__var1__"
	vars := cacao.NewVariables(var1)

	result, err := stix.Evaluate("__var1__:value = https://google.com", vars)
	assert.Equal(t, err, nil)
	assert.Equal(t, result, true)

	result, err = stix.Evaluate("__var1__:value = https://example.com", vars)
	assert.Equal(t, err, nil)
	assert.Equal(t, result, false)

	result, err = stix.Evaluate("__var1__:value > https://example.com", vars)
	assert.Equal(t, err, errors.New("operator: "+">"+" not valid or implemented"))
	assert.Equal(t, result, false)

}

func TestUuidEquals(t *testing.T) {
	stix := stix.New()

	var1 := cacao.Variable{Type: cacao.VariableTypeUuid}
	var1.Value = "ec887691-9a21-4ccf-8fae-360c13a819d1"
	var1.Name = "__var1__"
	vars := cacao.NewVariables(var1)

	result, err := stix.Evaluate("__var1__:value = ec887691-9a21-4ccf-8fae-360c13a819d1", vars)
	assert.Equal(t, err, nil)
	assert.Equal(t, result, true)

	result, err = stix.Evaluate("__var1__:value = ec887691-9a21-4ccf-8fae-360c13a819d2", vars)
	assert.Equal(t, err, nil)
	assert.Equal(t, result, false)

	result, err = stix.Evaluate("__var1__:value != ec887691-9a21-4ccf-8fae-360c13a819d2", vars)
	assert.Equal(t, err, nil)
	assert.Equal(t, result, true)

	result, err = stix.Evaluate("__var1__:value > ec887691-9a21-4ccf-8fae-360c13a819d2", vars)
	assert.Equal(t, err, errors.New("operator: "+">"+" not valid or implemented"))
	assert.Equal(t, result, false)

}
