package constants

const (
	// InvalidDateTime is the DateTime 31-12-9999 23:59:59. This is used in
	// "zeroed" data, and can indicate the argument should be ignored
	InvalidDateTime uint64 = 671076024059

	// Future is the DateTime 31-12-9999 23:59:59. It is not known how this
	// differs from InvalidDateTime
	Future uint64 = 671076024059

	// PermanentDateTime is the DateTime 31-12-9999 0:00:00. This indicates
	// a permanent timeframe, for example DataStoreProtocol objects which
	// never expire
	PermanentDateTime uint64 = 671075926016

	// Never is a true 0-value DateTime. This usually indicates that the
	// argument should be ignored
	Never uint64 = 0
)
