package libp2p

import (
	"testing"

	"github.com/keep-network/keep-core/pkg/net/gen/pb"
	"github.com/keep-network/keep-core/pkg/net/key"
)

func TestSignAndVerify(t *testing.T) {
	privateKey, _, err := key.GenerateStaticNetworkKey()
	if err != nil {
		t.Fatal(err)
	}

	identity, err := createIdentity(privateKey)
	if err != nil {
		t.Fatal(err)
	}

	identityBytes, err := identity.Marshal()
	if err != nil {
		t.Fatal(err)
	}

	message := &pb.UnicastNetworkMessage{
		Sender:  identityBytes,
		Payload: []byte{5, 15, 25, 30, 35},
		Type:    []byte{1},
	}

	err = signMessage(message, identity.privKey)
	if err != nil {
		t.Fatal(err)
	}

	err = verifyMessageSignature(message, identity.pubKey)
	if err != nil {
		t.Fatal(err)
	}
}
