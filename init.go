package nex

import (
	"github.com/PretendoNetwork/nex-go/types"
	"github.com/PretendoNetwork/plogger-go"
)

var logger = plogger.NewLogger()

func init() {
	initResultCodes()

	types.RegisterVariantType(1, types.NewPrimitiveS64(0))
	types.RegisterVariantType(2, types.NewPrimitiveF64(0))
	types.RegisterVariantType(3, types.NewPrimitiveBool(false))
	types.RegisterVariantType(4, types.NewString(""))
	types.RegisterVariantType(5, types.NewDateTime(0))
	types.RegisterVariantType(6, types.NewPrimitiveU64(0))
}
