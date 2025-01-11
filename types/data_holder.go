package types

// DataInterface defines an interface to track types which have Data anywhere
// in their parent tree.
type DataInterface interface {
	HoldableObject
	DataObjectID() RVType // Returns the object identifier of the type embedding Data
}

// DataHolder is an AnyObjectHolder for types which embed Data
type DataHolder = AnyObjectHolder[DataInterface]

// NewDataHolder returns a new DataHolder
func NewDataHolder() DataHolder {
	return DataHolder{}
}
