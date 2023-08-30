package nex

import (
	"fmt"
	"strings"

	"golang.org/x/mod/semver"
)

// NEXVersion represents a NEX library version
type NEXVersion struct {
	Major             int
	Minor             int
	Patch             int
	GameSpecificPatch string
	semver            string
}

// Copy returns a new copied instance of NEXVersion
func (nexVersion *NEXVersion) Copy() *NEXVersion {
	return &NEXVersion{
		Major:             nexVersion.Major,
		Minor:             nexVersion.Minor,
		Patch:             nexVersion.Patch,
		GameSpecificPatch: nexVersion.GameSpecificPatch,
		semver:            fmt.Sprintf("v%d.%d.%d", nexVersion.Major, nexVersion.Minor, nexVersion.Patch),
	}
}

func (nexVersion *NEXVersion) semverCompare(compare string) int {
	if !strings.HasPrefix(compare, "v") {
		// * Faster than doing "v" + string(compare)
		var b strings.Builder

		b.WriteString("v")
		b.WriteString(compare)

		compare = b.String()
	}

	if !semver.IsValid(compare) {
		// * The semver package returns 0 (equal) for invalid semvers in semver.Compare
		return 0
	}

	return semver.Compare(nexVersion.semver, compare)
}

// GreaterOrEqual compares if the given semver is greater than or equal to the current version
func (nexVersion *NEXVersion) GreaterOrEqual(compare string) bool {
	return nexVersion.semverCompare(compare) != -1
}

// LessOrEqual compares if the given semver is lesser than or equal to the current version
func (nexVersion *NEXVersion) LessOrEqual(compare string) bool {
	return nexVersion.semverCompare(compare) != 1
}

// NewPatchedNEXVersion returns a new NEXVersion with a game specific patch
func NewPatchedNEXVersion(major, minor, patch int, gameSpecificPatch string) *NEXVersion {
	return &NEXVersion{
		Major:             major,
		Minor:             minor,
		Patch:             patch,
		GameSpecificPatch: gameSpecificPatch,
		semver:            fmt.Sprintf("v%d.%d.%d", major, minor, patch),
	}
}

// NewNEXVersion returns a new NEXVersion
func NewNEXVersion(major, minor, patch int) *NEXVersion {
	return &NEXVersion{
		Major:  major,
		Minor:  minor,
		Patch:  patch,
		semver: fmt.Sprintf("v%d.%d.%d", major, minor, patch),
	}
}
