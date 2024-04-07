package nex

import (
	"fmt"
	"strings"

	"golang.org/x/mod/semver"
)

// LibraryVersion represents a NEX library version
type LibraryVersion struct {
	Major             int
	Minor             int
	Patch             int
	GameSpecificPatch string
	semver            string
}

// Copy returns a new copied instance of LibraryVersion
func (lv *LibraryVersion) Copy() *LibraryVersion {
	return &LibraryVersion{
		Major:             lv.Major,
		Minor:             lv.Minor,
		Patch:             lv.Patch,
		GameSpecificPatch: lv.GameSpecificPatch,
		semver:            fmt.Sprintf("v%d.%d.%d", lv.Major, lv.Minor, lv.Patch),
	}
}

func (lv *LibraryVersion) semverCompare(compare string) int {
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

	return semver.Compare(lv.semver, compare)
}

// GreaterOrEqual compares if the given semver is greater than or equal to the current version
func (lv *LibraryVersion) GreaterOrEqual(compare string) bool {
	return lv.semverCompare(compare) != -1
}

// LessOrEqual compares if the given semver is lesser than or equal to the current version
func (lv *LibraryVersion) LessOrEqual(compare string) bool {
	return lv.semverCompare(compare) != 1
}

// NewPatchedLibraryVersion returns a new LibraryVersion with a game specific patch
func NewPatchedLibraryVersion(major, minor, patch int, gameSpecificPatch string) *LibraryVersion {
	return &LibraryVersion{
		Major:             major,
		Minor:             minor,
		Patch:             patch,
		GameSpecificPatch: gameSpecificPatch,
		semver:            fmt.Sprintf("v%d.%d.%d", major, minor, patch),
	}
}

// NewLibraryVersion returns a new LibraryVersion
func NewLibraryVersion(major, minor, patch int) *LibraryVersion {
	return &LibraryVersion{
		Major:  major,
		Minor:  minor,
		Patch:  patch,
		semver: fmt.Sprintf("v%d.%d.%d", major, minor, patch),
	}
}

// LibraryVersions contains a set of the NEX version that the server uses
type LibraryVersions struct {
	Main         *LibraryVersion
	DataStore    *LibraryVersion
	MatchMaking  *LibraryVersion
	Ranking      *LibraryVersion
	Ranking2     *LibraryVersion
	Messaging    *LibraryVersion
	Utility      *LibraryVersion
	NATTraversal *LibraryVersion
}

// SetDefault sets the default NEX protocol versions
func (lvs *LibraryVersions) SetDefault(version *LibraryVersion) {
	lvs.Main = version
	lvs.DataStore = version.Copy()
	lvs.MatchMaking = version.Copy()
	lvs.Ranking = version.Copy()
	lvs.Ranking2 = version.Copy()
	lvs.Messaging = version.Copy()
	lvs.Utility = version.Copy()
	lvs.NATTraversal = version.Copy()
}

// NewLibraryVersions returns a new set of LibraryVersions
func NewLibraryVersions() *LibraryVersions {
	return &LibraryVersions{}
}
