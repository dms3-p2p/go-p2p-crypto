package crypto

import (
	"bytes"
	"fmt"
	"io"

	pb "github.com/dms3-p2p/go-p2p-crypto/pb"
	"golang.org/x/crypto/ed25519"
)

type Ed25519PrivateKey struct {
	k ed25519.PrivateKey
}

type Ed25519PublicKey struct {
	k ed25519.PublicKey
}

func GenerateEd25519Key(src io.Reader) (PrivKey, PubKey, error) {
	pub, priv, err := ed25519.GenerateKey(src)
	if err != nil {
		return nil, nil, err
	}

	return &Ed25519PrivateKey{
			k: priv,
		},
		&Ed25519PublicKey{
			k: pub,
		},
		nil
}

func (k *Ed25519PrivateKey) Type() pb.KeyType {
	return pb.KeyType_Ed25519
}

func (k *Ed25519PrivateKey) Bytes() ([]byte, error) {
	return MarshalPrivateKey(k)
}

func (k *Ed25519PrivateKey) Raw() ([]byte, error) {
	// Intentionally redundant for backwards compatibility.
	// Issue: #36
	// TODO: Remove the second copy of the public key at some point in the future.
	buf := make([]byte, len(k.k)+ed25519.PublicKeySize)
	copy(buf, k.k)
	copy(buf[len(k.k):], k.pubKeyBytes())

	return buf, nil
}

func (k *Ed25519PrivateKey) pubKeyBytes() []byte {
	return k.k[ed25519.PrivateKeySize-ed25519.PublicKeySize:]
}

func (k *Ed25519PrivateKey) Equals(o Key) bool {
	edk, ok := o.(*Ed25519PrivateKey)
	if !ok {
		return false
	}

	return bytes.Equal(k.k, edk.k)
}

func (k *Ed25519PrivateKey) GetPublic() PubKey {
	return &Ed25519PublicKey{k: k.pubKeyBytes()}
}

func (k *Ed25519PrivateKey) Sign(msg []byte) ([]byte, error) {
	return ed25519.Sign(k.k, msg), nil
}

func (k *Ed25519PublicKey) Type() pb.KeyType {
	return pb.KeyType_Ed25519
}

func (k *Ed25519PublicKey) Bytes() ([]byte, error) {
	return MarshalPublicKey(k)
}

func (k *Ed25519PublicKey) Raw() ([]byte, error) {
	return k.k, nil
}

func (k *Ed25519PublicKey) Equals(o Key) bool {
	edk, ok := o.(*Ed25519PublicKey)
	if !ok {
		return false
	}

	return bytes.Equal(k.k, edk.k)
}

func (k *Ed25519PublicKey) Verify(data []byte, sig []byte) (bool, error) {
	return ed25519.Verify(k.k, data, sig), nil
}

func UnmarshalEd25519PublicKey(data []byte) (PubKey, error) {
	if len(data) != 32 {
		return nil, fmt.Errorf("expect ed25519 public key data size to be 32")
	}
	return &Ed25519PublicKey{
		k: ed25519.PublicKey(data),
	}, nil
}

func UnmarshalEd25519PrivateKey(data []byte) (PrivKey, error) {
	switch len(data) {
	case ed25519.PrivateKeySize + ed25519.PublicKeySize:
		// Remove the redundant public key. See issue #36.
		redundantPk := data[ed25519.PrivateKeySize:]
		if !bytes.Equal(data[len(data)-ed25519.PublicKeySize:], redundantPk) {
			return nil, fmt.Errorf("expected redundant ed25519 public key to be redundant")
		}

		// No point in storing the extra data.
		newKey := make([]byte, ed25519.PrivateKeySize)
		copy(newKey, data[:ed25519.PrivateKeySize])
		data = newKey
	case ed25519.PrivateKeySize:
	default:
		return nil, fmt.Errorf(
			"expected ed25519 data size to be %d or %d",
			ed25519.PrivateKeySize,
			ed25519.PublicKeySize+ed25519.PublicKeySize,
		)
	}
	return &Ed25519PrivateKey{
		k: ed25519.PrivateKey(data),
	}, nil
}
