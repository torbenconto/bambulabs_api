package mqtt

import (
	"encoding/json"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
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

func TestCommand_AddParamField(t *testing.T) {
	cmd := NewCommand(Print)
	cmd.AddParamField("value")
	assert.Equal(t, "value", cmd.fields["param"])
}

func TestCommand_SetId(t *testing.T) {
	cmd := NewCommand(Print)
	cmd.SetId(123)
	assert.Equal(t, "123", cmd.fields["sequence_id"])
	assert.Equal(t, "123", strconv.Itoa(cmd.id))
}

func TestCommand_JSON(t *testing.T) {
	cmd := NewCommand(Print)
	cmd.AddCommandField("testCommandField")
	cmd.AddParamField("testParamField")
	cmd.AddField("extra", "data")

	jsonStr, err := cmd.JSON()
	assert.NoError(t, err)

	var result map[string]interface{}
	err = json.Unmarshal([]byte(jsonStr), &result)
	assert.NoError(t, err)

	data, ok := result["print"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "testCommandField", data["command"])
	assert.Equal(t, "testParamField", data["param"])
	assert.Equal(t, "data", data["extra"])
}
