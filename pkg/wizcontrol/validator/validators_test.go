package validator

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMacValidation(t *testing.T) {

	assert := assert.New(t)
	require := require.New(t)
	mac, err := MAC(" 12:34:56:78:9a:BC ", true)
	require.NoError(err)
	assert.Equal("12:34:56:78:9a:bc", mac)

	mac, err = MAC(" 123456789abC ", true)
	require.NoError(err)
	assert.Equal("12:34:56:78:9a:bc", mac)

	mac, err = MAC(" 1234.5678.9aBc ", true)
	require.NoError(err)
	assert.Equal("12:34:56:78:9a:bc", mac)
}

func TestIPValidation(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)
	ip, err := IP(" 192.168.1.1 ", true)
	require.NoError(err)
	assert.Equal("192.168.1.1", ip)

	ip, err = IP(" 127.0.0.1 ", true)
	require.NoError(err)
	assert.Equal("127.0.0.1", ip)
}
