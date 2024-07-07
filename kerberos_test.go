package nex

import (
	"encoding/hex"
	"testing"

	"github.com/PretendoNetwork/nex-go/v2/types"
	"github.com/stretchr/testify/assert"
)

func TestDeriveGuestKey(t *testing.T) {
	pid := types.NewPID(100)
	password := []byte("MMQea3n!fsik")
	result := DeriveKerberosKey(pid, password)
	assert.Equal(t, "9ef318f0a170fb46aab595bf9644f9e1", hex.EncodeToString(result))
}
