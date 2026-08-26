package metrics

import (
	"bytes"
	"os"
	"os/exec"
	"testing"
)

func TestBug006_SnapshotDoesNotBlockConcurrentWriters(t *testing.T) {
	if os.Getenv("ORBIT_BUG006_CHILD") == "1" {
		r := New()
		r.Add("frames", 1)
		_ = r.Snapshot()
		return
	}
	child := exec.Command(os.Args[0], "-test.run=^TestBug006_SnapshotDoesNotBlockConcurrentWriters$", "-test.v")
	child.Env = append(os.Environ(), "ORBIT_BUG006_CHILD=1")
	out, err := child.CombinedOutput()
	if err == nil || !bytes.Contains(out, []byte("RUnlock")) {
		t.Fatalf("expected Snapshot lock-protocol failure, err=%v output=%s", err, out)
	}
	t.Fatalf("Snapshot reproduced lock-protocol failure: %s", out)
/*
	r := New(); r.Add("frames", 1)
	if got := r.Snapshot()["frames"]; got != 1 { t.Fatalf("unexpected snapshot value %d", got) }
	r.Add("frames", 1)
	if got := r.Get("frames"); got != 2 { t.Fatalf("writer stalled after snapshot: %d", got) }
*/
}
