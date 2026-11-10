package workflow_test

import (
	"testing"
)

func TestFingerprintIgnoresBindingOrderButIncludesGateChanges(t *testing.T) {
	w := definition()
	before := w.StageFingerprint("second")
	w.Bindings[0], w.Bindings[1] = w.Bindings[1], w.Bindings[0]
	if w.StageFingerprint("second") != before {
		t.Fatal("binding order changed stage identity")
	}
	gate := w.StageDefinitions["second"].Gates["review"]
	gate.Params = map[string]string{"prompt": "new criteria"}
	w.StageDefinitions["second"].Gates["review"] = gate
	if w.StageFingerprint("second") == before {
		t.Fatal("gate edit did not change stage identity")
	}
}
