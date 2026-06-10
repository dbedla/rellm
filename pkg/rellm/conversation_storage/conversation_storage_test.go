package conversation_storage

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestConversationStorage(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "conversation_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() {
		err := os.RemoveAll(tempDir)
		if err != nil {
			t.Fatalf("failed to remove temp dir: %v", err)
		}
	}()

	storage := New(tempDir)

	// 1. Test Restore from empty directory
	msgs, err := storage.RestoreLatest()
	if err != nil {
		t.Errorf("RestoreLatest from empty dir failed: %v", err)
	}
	if len(msgs) != 0 {
		t.Errorf("expected empty conversation, got %d messages", len(msgs))
	}

	// 2. Test Store and RestoreSpecific
	msg1 := json.RawMessage(`{"role": "user", "content": "hello"}`)
	msg2 := json.RawMessage(`{"role": "assistant", "content": "hi there"}`)
	originalConversation := []json.RawMessage{msg1, msg2}

	err = storage.Store(originalConversation)
	if err != nil {
		t.Fatalf("Store failed: %v", err)
	}

	// We need to wait a second to ensure different timestamps if we store again
	// but for single store we can just find the file
	files, _ := os.ReadDir(tempDir)
	var fileName string
	for _, f := range files {
		fileName = f.Name()
	}

	restored, err := storage.RestoreSpecific(fileName)
	if err != nil {
		t.Errorf("RestoreSpecific failed: %v", err)
	}

	var v1, v2 []any
	dataOrig, _ := json.Marshal(originalConversation)
	dataRestored, _ := json.Marshal(restored)
	err = json.Unmarshal(dataOrig, &v1)
	assert.NoError(t, err)
	err = json.Unmarshal(dataRestored, &v2)
	assert.NoError(t, err)

	if !reflect.DeepEqual(v1, v2) {
		t.Errorf("restored conversation does not match original")
	}

	// 3. Test RestoreLatest
	time.Sleep(1 * time.Second)
	msg3 := json.RawMessage(`{"role": "user", "content": "latest message"}`)
	latestConversation := []json.RawMessage{msg3}
	err = storage.Store(latestConversation)
	if err != nil {
		t.Fatalf("Store second failed: %v", err)
	}

	restoredLatest, err := storage.RestoreLatest()
	if err != nil {
		t.Errorf("RestoreLatest failed: %v", err)
	}

	dataLatest, _ := json.Marshal(latestConversation)
	dataRestoredLatest, _ := json.Marshal(restoredLatest)
	var v3, v4 []any
	err = json.Unmarshal(dataLatest, &v3)
	assert.NoError(t, err)
	err = json.Unmarshal(dataRestoredLatest, &v4)
	assert.NoError(t, err)

	if !reflect.DeepEqual(v3, v4) {
		t.Errorf("restored latest conversation does not match expected")
	}

	// 4. Test RestoreSpecific with non-existent file
	msgs, err = storage.RestoreSpecific("non-existent.json")
	if err != nil {
		t.Errorf("RestoreSpecific for non-existent file failed: %v", err)
	}
	if len(msgs) != 0 {
		t.Errorf("expected empty conversation for non-existent file, got %d messages", len(msgs))
	}
}
