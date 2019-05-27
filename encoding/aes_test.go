package encoding

import (
	"crypto/aes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAesCbc(t *testing.T) {
	key := []byte{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1}
	runs := [][]byte{
		{1},
		{1, 2, 3},
		{1, 2, 3, 4, 5, 6, 7},
		{1, 2, 3, 4, 5, 6, 7, 8},
		{1, 2, 3, 4, 5, 6, 7, 8, 1, 2, 3, 4, 5, 6},
	}
	for _, packet := range runs {
		// Encrypt packet
		encrypted, err := AESencryptCBC(key, packet)
		require.NoError(t, err)
		assert.NotEqual(t, packet, encrypted)
		// Decrypt it back
		decrypted, err := AESdecryptCBC(key, encrypted)
		assert.NoError(t, err)
		assert.Equal(t, packet, decrypted)
	}
}

func TestAlignPacketLength(t *testing.T) {
	runs := map[int]int{
		0:  0,
		1:  aes.BlockSize,
		8:  aes.BlockSize,
		16: aes.BlockSize,
		17: aes.BlockSize * 2,
		29: aes.BlockSize * 2,
		31: aes.BlockSize * 2,
	}
	for size, expected := range runs {
		res := alignPacketLength(size)
		assert.Equal(t, expected, res)
	}
}
