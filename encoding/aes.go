package encoding

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
)

// AESencryptCBC encrypts packet with AES-CBC using
// random generated AES initialization vector
func AESencryptCBC(key, packet []byte) ([]byte, error) {
	// LoRa limits messages to 256 bytes long
	packetLen := len(packet)
	if packetLen > 255 {
		return nil, errors.New("Packet too long")
	}
	// Allocate buffer for encryption:
	// - IV
	// - Aligned to aes.blocksize packet
	alignedLength := alignPacketLength(packetLen)
	buf := make([]byte, alignedLength+aes.BlockSize)
	// Initialize buffer with random values (first 16 bytes will be used as IV)
	_, err := rand.Read(buf)
	if err != nil {
		return nil, err
	}
	// First byte carries actual payload length:
	buf[0] = byte(packetLen)
	// Copy packet into aes.blocksize aligned buffer
	alignedPacket := make([]byte, alignedLength)
	copy(alignedPacket, packet)
	// Encrypt using AES CBC
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	encryptor := cipher.NewCBCEncrypter(block, buf[:aes.BlockSize])
	encryptor.CryptBlocks(buf[aes.BlockSize:], alignedPacket)

	return buf, nil
}

// AESdecryptCBC decrypts message encrypted with AES-CBC.
// Message must be prepended with AES-IV (initialization vector)
// AESBlock size long (usually 16 bytes)
func AESdecryptCBC(key, packet []byte) ([]byte, error) {
	packetLen := len(packet)
	// Messages less that 2 AES block size are invalid (IV + 1 block)
	if packetLen < aes.BlockSize*2 {
		return nil, errors.New("Packet too short")
	}
	// AES encrypted message must be multiple of AES block size
	if packetLen%aes.BlockSize != 0 {
		return nil, errors.New("Invalid packet length")
	}
	// First 16 bytes (AES block size) is IV (AES Initial Value)
	// Also, first byte carries actual payload length
	iv := packet[:aes.BlockSize]
	payloadLen := int(iv[0])
	// Ensure that payload length is correct: less that packet size - IV
	if payloadLen > packetLen-aes.BlockSize {
		return nil, errors.New("Invalid packet payload size")
	}
	// Decrypt using AES CBC
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	decryptor := cipher.NewCBCDecrypter(block, iv)
	decryptedPacket := make([]byte, packetLen-aes.BlockSize)
	decryptor.CryptBlocks(decryptedPacket, packet[aes.BlockSize:])

	return decryptedPacket[:payloadLen], nil
}

func alignPacketLength(packetLen int) int {
	if packetLen%aes.BlockSize != 0 {
		return (packetLen/aes.BlockSize + 1) * aes.BlockSize
	}
	return packetLen
}
