package gjkr

import (
	"testing"

	"github.com/keep-network/keep-core/pkg/internal/testutils"

	"github.com/keep-network/keep-core/pkg/protocol/group"
)

func TestFilterSymmetricKeyGeneratingMembers(t *testing.T) {
	member := (&LocalMember{
		memberCore: &memberCore{
			logger: &testutils.MockLogger{},
			ID:     13,
			group:  group.NewGroup(8, 15),
		},
	}).InitializeEphemeralKeysGeneration().
		InitializeSymmetricKeyGeneration()

	messages := []*EphemeralPublicKeyMessage{
		&EphemeralPublicKeyMessage{senderID: 11},
		&EphemeralPublicKeyMessage{senderID: 14},
	}

	member.MarkInactiveMembers(messages)

	assertAcceptsFrom(member.memberCore, 13, t) // should accept from self
	assertAcceptsFrom(member.memberCore, 11, t)
	assertAcceptsFrom(member.memberCore, 14, t)
	assertNotAcceptFrom(member.memberCore, 12, t)
	assertNotAcceptFrom(member.memberCore, 15, t)
}

func TestFilterCommitmentsVefiryingMembers(t *testing.T) {
	member := (&LocalMember{
		memberCore: &memberCore{
			logger: &testutils.MockLogger{},
			ID:     93,
			group:  group.NewGroup(49, 96),
		},
	}).InitializeEphemeralKeysGeneration().
		InitializeSymmetricKeyGeneration().
		InitializeCommitting().
		InitializeCommitmentsVerification()

	sharesMessages := []*PeerSharesMessage{
		&PeerSharesMessage{senderID: 91},
		&PeerSharesMessage{senderID: 92},
		&PeerSharesMessage{senderID: 94},
	}

	commitmentsMessages := []*MemberCommitmentsMessage{
		&MemberCommitmentsMessage{senderID: 92},
		&MemberCommitmentsMessage{senderID: 94},
		&MemberCommitmentsMessage{senderID: 95},
	}

	member.MarkInactiveMembers(sharesMessages, commitmentsMessages)

	// should accept from self
	assertAcceptsFrom(member.memberCore, 93, t)

	// 92 and 94 sent both shares message and commitments message
	assertAcceptsFrom(member.memberCore, 92, t)
	assertAcceptsFrom(member.memberCore, 94, t)

	// 95 did not send shares message
	assertNotAcceptFrom(member.memberCore, 95, t)

	// 91 did not send commitments message
	assertNotAcceptFrom(member.memberCore, 91, t)

	// 96 did not send shares message nor commitments message
	assertNotAcceptFrom(member.memberCore, 96, t)
}

func TestFilterSharingMembers(t *testing.T) {
	member := (&LocalMember{
		memberCore: &memberCore{
			logger: &testutils.MockLogger{},
			ID:     24,
			group:  group.NewGroup(13, 24),
		},
	}).InitializeEphemeralKeysGeneration().
		InitializeSymmetricKeyGeneration().
		InitializeCommitting().
		InitializeCommitmentsVerification().
		InitializeSharesJustification().
		InitializeQualified().
		InitializeSharing()

	messages := []*MemberPublicKeySharePointsMessage{
		&MemberPublicKeySharePointsMessage{senderID: 21},
		&MemberPublicKeySharePointsMessage{senderID: 23},
	}

	member.MarkInactiveMembers(messages)

	assertAcceptsFrom(member.memberCore, 24, t) // should accept from self
	assertAcceptsFrom(member.memberCore, 21, t)
	assertAcceptsFrom(member.memberCore, 23, t)
	assertNotAcceptFrom(member.memberCore, 22, t)
}

func TestFilterReconstructingMember(t *testing.T) {
	member := (&LocalMember{
		memberCore: &memberCore{
			logger: &testutils.MockLogger{},
			ID:     44,
			group:  group.NewGroup(23, 44),
		},
	}).InitializeEphemeralKeysGeneration().
		InitializeSymmetricKeyGeneration().
		InitializeCommitting().
		InitializeCommitmentsVerification().
		InitializeSharesJustification().
		InitializeQualified().
		InitializeSharing().
		InitializePointsJustification().
		InitializeRevealing().
		InitializeReconstruction()

	messages := []*MisbehavedEphemeralKeysMessage{
		{senderID: 41},
	}

	member.MarkInactiveMembers(messages)

	assertAcceptsFrom(member.memberCore, 44, t) // should accept from self
	assertAcceptsFrom(member.memberCore, 41, t)
	assertNotAcceptFrom(member.memberCore, 42, t)
	assertNotAcceptFrom(member.memberCore, 43, t)
}

func assertAcceptsFrom(member *memberCore, senderID group.MemberIndex, t *testing.T) {
	if !member.group.IsOperating(senderID) {
		t.Errorf("member should accept messages from [%v]", senderID)
	}
}

func assertNotAcceptFrom(member *memberCore, senderID group.MemberIndex, t *testing.T) {
	if member.group.IsOperating(senderID) {
		t.Errorf("member should not accept messages from [%v]", senderID)
	}
}
