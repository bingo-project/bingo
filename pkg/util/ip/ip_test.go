package ip

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContains(t *testing.T) {
	ip := "10.10.10.10"
	cidr := "10.10.10.0/24"

	// Test ip contains ip
	ret := Contains(ip, ip)
	assert.True(t, ret)

	// Test error cidr
	ret = Contains("10.10.10.0/100", ip)
	assert.False(t, ret)

	// Test contains in cidr
	ret = Contains(cidr, ip)
	assert.True(t, ret)

	// Test not contains in cidr
	ret = Contains(cidr, "100.100.100.100")
	assert.False(t, ret)
}

func TestContainsInCIDR(t *testing.T) {
	ip := "10.10.10.10"
	cidr := []string{"10.10.10.0/24"}

	// Test contains
	ret := ContainsInCIDR(cidr, ip)
	assert.True(t, ret)

	// Test not contains
	ret = ContainsInCIDR(cidr, "100.100.100.100")
	assert.False(t, ret)
}
