package sync

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRunSyncRound_ErrorIsolation(t *testing.T) {
	// UID001 fails, UID002 and UID003 succeed - failures should be isolated
	var calledUIDs []string
	fn := func(ctx context.Context, uid string) error {
		calledUIDs = append(calledUIDs, uid)
		if uid == "UID001" {
			return errors.New("API error for UID001")
		}
		return nil
	}

	s := newTestSyncer(&mockQuanyuClient{}, &mockSyncStorage{}, &mockDeviceProvider{
		uids: []string{"UID001", "UID002", "UID003"},
	})

	s.runSyncRound(context.Background(), "detail", fn)

	assert.Equal(t, []string{"UID001", "UID002", "UID003"}, calledUIDs,
		"all UIDs should be attempted even when one fails")
}

func TestRunSyncRound_EmptyUIDList(t *testing.T) {
	fn := func(ctx context.Context, uid string) error {
		t.Fatal("should not be called with empty UID list")
		return nil
	}

	s := newTestSyncer(&mockQuanyuClient{}, &mockSyncStorage{}, &mockDeviceProvider{
		uids: []string{},
	})

	// Should not panic, just log and return
	s.runSyncRound(context.Background(), "detail", fn)
}

func TestRunSyncRound_AllFail(t *testing.T) {
	var callCount int
	fn := func(ctx context.Context, uid string) error {
		callCount++
		return errors.New("always fails")
	}

	s := newTestSyncer(&mockQuanyuClient{}, &mockSyncStorage{}, &mockDeviceProvider{
		uids: []string{"UID001", "UID002"},
	})

	s.runSyncRound(context.Background(), "detail", fn)
	assert.Equal(t, 2, callCount, "all UIDs should be attempted")
}
