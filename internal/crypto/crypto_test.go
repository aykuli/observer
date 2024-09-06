// Package crypto provides methods to encode with public key and decode with private key functionality
package crypto

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEncryptor(t *testing.T) {
	tMsg := `Lorem ipsum dolor sit amet, consectetur adipiscing elit. Ut convallis tristique urna. Nulla molestie libero ullamcorper ante faucibus semper. Sed euismod lectus vel magna luctus, ac vulputate urna scelerisque. Curabitur id sodales lectus. Class aptent taciti sociosqu ad litora torquent per conubia nostra, per inceptos himenaeos. Praesent eget ipsum placerat, tristique mi pretium, pharetra ex. Fusce tristique rhoncus velit, fermentum mattis turpis facilisis malesuada.`
	dir := "."
	pubKeyPath, privKeyPath, err := GenerateRsaKeys(dir)
	require.NoError(t, err)
	require.FileExists(t, dir+"/rsa.pub")
	require.FileExists(t, dir+"/rsa")

	enc, err := NewEncryptor(pubKeyPath)
	require.NoError(t, err)
	dec, err := NewDecryptor(privKeyPath)
	require.NoError(t, err)
	require.NotNil(t, enc.pubKey)
	require.NotNil(t, dec.privKey)

	encrypted, err := enc.Encrypt([]byte(tMsg))
	require.NoError(t, err)
	require.NotEmpty(t, encrypted)

	decrypted, err := dec.Decrypt(encrypted)
	require.NoError(t, err)
	require.NotEmpty(t, decrypted)

	require.Equal(t, decrypted, tMsg)

	err = os.Remove(pubKeyPath)
	require.NoError(t, err)
	err = os.Remove(privKeyPath)
	require.NoError(t, err)

	require.NoFileExists(t, pubKeyPath)
	require.NoFileExists(t, privKeyPath)
}
