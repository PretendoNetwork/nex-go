package nex

import (
	"github.com/PretendoNetwork/nex-go/v2/types"
	"github.com/PretendoNetwork/plogger-go"
)

var logger = plogger.NewLogger()

func init() {
	initResultCodes()

	types.RegisterVariantType(1, types.NewInt64(0))
	types.RegisterVariantType(2, types.NewDouble(0))
	types.RegisterVariantType(3, types.NewBool(false))
	types.RegisterVariantType(4, types.NewString(""))
	types.RegisterVariantType(5, types.NewDateTime(0))
	types.RegisterVariantType(6, types.NewUInt64(0))
}
