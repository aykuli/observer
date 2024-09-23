// Package crypto provides methods to encode with public key and decode with private key functionality
// References
// |-- https://earthly.dev/blog/encrypting-data-with-ssh-keys-and-golang/
// |-- https://stackoverflow.com/questions/62348923/rs256-message-too-long-for-rsa-public-key-size-error-signing-jwt
package crypto

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha512"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"os"

	"golang.org/x/crypto/ssh"
)

// GenerateRsaKeys generates public and private RSA keys accordingly to the directory given as argument
func GenerateRsaKeys(dir string) (string, string, error) {
	reader := rand.Reader
	bitSize := 2048

	key, err := rsa.GenerateKey(reader, bitSize)
	if err != nil {
		return "", "", err
	}

	pub, err := ssh.NewPublicKey(key.Public())
	if err != nil {
		return "", "", err
	}
	pubKeyBytes := ssh.MarshalAuthorizedKey(pub)
	privKeyBytes := pem.EncodeToMemory(&pem.Block{
		Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key),
	})

	fname := dir + "/rsa"
	if err = os.WriteFile(fname+".pub", pubKeyBytes, 0644); err != nil {
		return "", "", err
	}
	if err = os.WriteFile(fname, privKeyBytes, 0644); err != nil {
		return "", "", err
	}

	return fname + ".pub", fname, nil
}

// Encryptor is the structure container for public key
type Encryptor struct {
	pubKey *rsa.PublicKey
}

// NewEncryptor creates structure Encryptor
func NewEncryptor(pubKeyPath string) (*Encryptor, error) {
	pubKeyFromFile, err := os.ReadFile(pubKeyPath)
	if err != nil {
		return nil, err
	}
	parsed, _, _, _, err := ssh.ParseAuthorizedKey(pubKeyFromFile)
	if err != nil {
		return nil, err
	}

	parsedCryptoKey := parsed.(ssh.CryptoPublicKey)
	pubCrypto := parsedCryptoKey.CryptoPublicKey()
	rsaPubKey := pubCrypto.(*rsa.PublicKey)
	return &Encryptor{pubKey: rsaPubKey}, nil
}

// Encrypt encrypts gives message in bytes to string
func (e *Encryptor) Encrypt(msg []byte) (string, error) {
	msgLen := len(msg)
	step := e.pubKey.Size() - 2*sha512.Size - 2
	var encSumm []byte
	for start := 0; start < msgLen; start += step {
		finish := start + step
		if finish > msgLen {
			finish = msgLen
		}

		encBytes, err := rsa.EncryptOAEP(sha512.New(), rand.Reader, e.pubKey, msg[start:finish], nil)
		if err != nil {
			return "", err
		}
		encSumm = append(encSumm, encBytes...)
	}

	return base64.StdEncoding.EncodeToString(encSumm), nil
}

// Decryptor is the structure container for private key
type Decryptor struct {
	privKey *rsa.PrivateKey
}

// NewDecryptor creates structure Decryptor
func NewDecryptor(privKeyPath string) (*Decryptor, error) {
	privKeyFromFile, err := os.ReadFile(privKeyPath)
	if err != nil {
		return nil, err
	}

	block, rest := pem.Decode(privKeyFromFile)
	if len(rest) > 0 {
		return nil, err
	}
	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return &Decryptor{privKey: key}, nil
}

// Decrypt decrypts given ciphertext
func (d *Decryptor) Decrypt(ciphertext string) (string, error) {
	msg, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}
	msgLen := len(msg)
	step := d.privKey.PublicKey.Size()
	var decSumm []byte
	for start := 0; start < msgLen; start += step {
		finish := start + step
		if finish > msgLen {
			finish = msgLen
		}
		decrypted, err := rsa.DecryptOAEP(sha512.New(), rand.Reader, d.privKey, msg[start:finish], nil)
		if err != nil {
			return "", err
		}
		decSumm = append(decSumm, decrypted...)
	}

	return string(decSumm), err
}
