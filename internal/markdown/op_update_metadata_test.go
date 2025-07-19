package markdown

import "testing"

func TestOpUpdateMetadata_Apply(t *testing.T) {
	// Metadata update is a stub; just ensure no error is returned
	op := &OpUpdateMetadata{
		Selector: SelectorPath{Segments: []string{"Any"}},
		Metadata: map[string]interface{}{"foo": "bar"},
	}
	// No panic or error expected
	if err := op.Apply(nil); err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
}
