package nex

// NEXVersion represents a NEX library version
type NEXVersion struct {
	Major             int
	Minor             int
	Patch             int
	GameSpecificPatch string
}

// Copy returns a new copied instance of NEXVersion
func (nexVersion *NEXVersion) Copy() *NEXVersion {
	var copied *NEXVersion

	if nexVersion.GameSpecificPatch != "" {
		copied = NewPatchedNEXVersion(nexVersion.Major, nexVersion.Minor, nexVersion.Patch, nexVersion.GameSpecificPatch)
	} else {
		copied = NewNEXVersion(nexVersion.Major, nexVersion.Minor, nexVersion.Patch)
	}

	return copied
}

// NewPatchedNEXVersion returns a new NEXVersion with a game specific patch
func NewPatchedNEXVersion(major, minor, patch int, gameSpecificPatch string) *NEXVersion {
	return &NEXVersion{
		Major:             major,
		Minor:             minor,
		Patch:             patch,
		GameSpecificPatch: gameSpecificPatch,
	}
}

// NewNEXVersion returns a new NEXVersion
func NewNEXVersion(major, minor, patch int) *NEXVersion {
	return &NEXVersion{
		Major: major,
		Minor: minor,
		Patch: patch,
	}
}
