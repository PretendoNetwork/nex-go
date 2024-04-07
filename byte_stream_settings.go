package nex

// ByteStreamSettings defines some settings for how a ByteStream should handle certain data types
type ByteStreamSettings struct {
	StringLengthSize   int
	PIDSize            int
	UseStructureHeader bool
}

// NewByteStreamSettings returns a new ByteStreamSettings
func NewByteStreamSettings() *ByteStreamSettings {
	return &ByteStreamSettings{
		StringLengthSize:   2,
		PIDSize:            4,
		UseStructureHeader: false,
	}
}
