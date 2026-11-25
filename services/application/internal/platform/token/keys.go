package token

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
)

// PEM block types. The private file holds the two private keys (PKCS#8),
// the public file holds the two public keys (PKIX).
const (
	blockAccessPrivate  = "ACCESS PRIVATE KEY"
	blockRefreshPrivate = "REFRESH PRIVATE KEY"
	blockAccessPublic   = "ACCESS PUBLIC KEY"
	blockRefreshPublic  = "REFRESH PUBLIC KEY"
)

const minRSABits = 2048

// Keys holds one RS256 key pair for access tokens and one for refresh tokens.
type Keys struct {
	AccessPrivate  *rsa.PrivateKey
	RefreshPrivate *rsa.PrivateKey
	AccessPublic   *rsa.PublicKey
	RefreshPublic  *rsa.PublicKey
}

// LoadKeys reads both key files. Errors never include key material.
func LoadKeys(privatePath, publicPath string) (Keys, error) {
	priv, err := readBlocks(privatePath, blockAccessPrivate, blockRefreshPrivate)
	if err != nil {
		return Keys{}, err
	}
	pub, err := readBlocks(publicPath, blockAccessPublic, blockRefreshPublic)
	if err != nil {
		return Keys{}, err
	}

	var k Keys
	if k.AccessPrivate, err = parsePrivate(priv[blockAccessPrivate], blockAccessPrivate); err != nil {
		return Keys{}, err
	}
	if k.RefreshPrivate, err = parsePrivate(priv[blockRefreshPrivate], blockRefreshPrivate); err != nil {
		return Keys{}, err
	}
	if k.AccessPublic, err = parsePublic(pub[blockAccessPublic], blockAccessPublic); err != nil {
		return Keys{}, err
	}
	if k.RefreshPublic, err = parsePublic(pub[blockRefreshPublic], blockRefreshPublic); err != nil {
		return Keys{}, err
	}

	if !k.AccessPrivate.PublicKey.Equal(k.AccessPublic) {
		return Keys{}, errors.New("token: access public key does not match access private key")
	}
	if !k.RefreshPrivate.PublicKey.Equal(k.RefreshPublic) {
		return Keys{}, errors.New("token: refresh public key does not match refresh private key")
	}
	if k.AccessPublic.Equal(k.RefreshPublic) {
		return Keys{}, errors.New("token: access and refresh must use different key pairs")
	}
	return k, nil
}

// readBlocks returns the DER bytes of each wanted block type. Each must appear
// exactly once; other block types are rejected.
func readBlocks(path string, want ...string) (map[string][]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("token: read key file: %w", err)
	}
	wanted := make(map[string]bool, len(want))
	for _, w := range want {
		wanted[w] = true
	}

	out := make(map[string][]byte, len(want))
	for {
		var b *pem.Block
		b, data = pem.Decode(data)
		if b == nil {
			break
		}
		if !wanted[b.Type] {
			return nil, fmt.Errorf("token: %s: unexpected PEM block %q", path, b.Type)
		}
		if _, dup := out[b.Type]; dup {
			return nil, fmt.Errorf("token: %s: duplicate PEM block %q", path, b.Type)
		}
		out[b.Type] = b.Bytes
	}
	for _, w := range want {
		if _, ok := out[w]; !ok {
			return nil, fmt.Errorf("token: %s: missing PEM block %q", path, w)
		}
	}
	return out, nil
}

func parsePrivate(der []byte, name string) (*rsa.PrivateKey, error) {
	k, err := x509.ParsePKCS8PrivateKey(der)
	if err != nil {
		return nil, fmt.Errorf("token: %s: not a PKCS#8 private key", name)
	}
	rk, ok := k.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("token: %s: not an RSA key", name)
	}
	if rk.N.BitLen() < minRSABits {
		return nil, fmt.Errorf("token: %s: RSA key must be at least %d bits", name, minRSABits)
	}
	return rk, nil
}

func parsePublic(der []byte, name string) (*rsa.PublicKey, error) {
	k, err := x509.ParsePKIXPublicKey(der)
	if err != nil {
		return nil, fmt.Errorf("token: %s: not a PKIX public key", name)
	}
	rk, ok := k.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("token: %s: not an RSA key", name)
	}
	return rk, nil
}
