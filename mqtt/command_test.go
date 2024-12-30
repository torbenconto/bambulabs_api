package mqtt

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewCommand(t *testing.T) {
	cmd := NewCommand(Print)
	assert.Equal(t, Print, cmd.Type)
	assert.NotNil(t, cmd.fields)
}

func TestCommand_AddField(t *testing.T) {
	cmd := NewCommand(Print)
	cmd.AddField("key", "value")
	assert.Equal(t, "value", cmd.fields["key"])
}

func TestCommand_JSON(t *testing.T) {
	cmd := NewCommand(Print)
	cmd.AddCommandField("testParam")
	cmd.AddField("extra", "data")

	jsonStr, err := cmd.JSON()
	assert.NoError(t, err)

	var result map[string]interface{}
	err = json.Unmarshal([]byte(jsonStr), &result)
	assert.NoError(t, err)

	data, ok := result["print"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "testParam", data["command"])
	assert.Equal(t, "data", data["extra"])
}
