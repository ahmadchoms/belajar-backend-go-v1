package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPasswordHashing(t *testing.T) {
	password := "rahasia123"

	hash, err := HashPassword(password)

	// Memastikan tidak ada error saat hashing
	assert.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.NotEqual(t, password, hash) // Hash ga boleh sama kayak plain text

	isValid := CheckPasswordHash(password, hash)
	assert.True(t, isValid, "Password harusnya valid")

	isInvalid := CheckPasswordHash("blabla", hash)
	assert.False(t, isInvalid, "Password salah harusnya return false")
}
