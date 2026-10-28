package cd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContextIngressClass(t *testing.T) {
	assert.Equal(t, "", contextIngressClass(nil))
	assert.Equal(t, "", contextIngressClass(map[string]interface{}{}))
	assert.Equal(t, "haproxy", contextIngressClass(map[string]interface{}{"IngressClass": "haproxy"}))
}
