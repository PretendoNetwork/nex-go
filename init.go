package nex

import "github.com/PretendoNetwork/plogger-go"

var logger = plogger.NewLogger()

func init() {
	initErrorsData()
}
