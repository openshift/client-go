package securitycontextconstraints

import (
	"github.com/golang/glog"

	securityapi "github.com/openshift/origin/pkg/security/apis/security"
)

// ByRestrictions is a helper to sort SCCs in order of most restrictive to least restrictive.
type ByRestrictions []*securityapi.SecurityContextConstraints

func (s ByRestrictions) Len() int {
	return len(s)
}
func (s ByRestrictions) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s ByRestrictions) Less(i, j int) bool {
	return pointValue(s[i]) < pointValue(s[j])
}

// The following constants define the weight of the restrictions and used for
// calculating the points of the particular SCC. The lower the number, the more
// restrictive SCC is. Make sure that weak restrictions are always valued
// higher than the combination of the strong restrictions.

type points int

const (
	privilegedPoints points = 20

	hostVolumePoints       points = 10
	nonTrivialVolumePoints points = 5

	runAsAnyUserPoints points = 4
	runAsNonRootPoints points = 3
	runAsRangePoints   points = 2
	runAsUserPoints    points = 1

	noPoints points = 0
)

// pointValue places a value on the SCC based on the settings of the SCC that can be used
// to determine how restrictive it is.  The lower the number, the more restrictive it is.
func pointValue(constraint *securityapi.SecurityContextConstraints) points {
	totalPoints := noPoints

	if constraint.AllowPrivilegedContainer {
		totalPoints += privilegedPoints
	}

	// add points based on volume requests
	totalPoints += volumePointValue(constraint)

	// the map contains points for both RunAsUser and SELinuxContext
	// strategies by taking advantage that they have identical strategy names
	strategiesPoints := map[string]points{
		string(securityapi.RunAsUserStrategyRunAsAny):         runAsAnyUserPoints,
		string(securityapi.RunAsUserStrategyMustRunAsNonRoot): runAsNonRootPoints,
		string(securityapi.RunAsUserStrategyMustRunAsRange):   runAsRangePoints,
		string(securityapi.RunAsUserStrategyMustRunAs):        runAsUserPoints,
	}

	strategyType := string(constraint.SELinuxContext.Type)
	points, found := strategiesPoints[strategyType]
	if found {
		totalPoints += points
	} else {
		glog.Warningf("SELinuxContext type %q has no point value, this may cause issues in sorting SCCs by restriction", strategyType)
	}

	strategyType = string(constraint.RunAsUser.Type)
	points, found = strategiesPoints[strategyType]
	if found {
		totalPoints += points
	} else {
		glog.Warningf("RunAsUser type %q has no point value, this may cause issues in sorting SCCs by restriction", strategyType)
	}

	return totalPoints
}

// volumePointValue returns a score based on the volumes allowed by the SCC.
// Allowing a host volume will return a score of 10.  Allowance of anything other
// than Secret, ConfigMap, EmptyDir, DownwardAPI, Projected, and None will result in
// a score of 5.  If the SCC only allows these trivial types, it will have a
// score of 0.
func volumePointValue(scc *securityapi.SecurityContextConstraints) points {
	hasHostVolume := false
	hasNonTrivialVolume := false
	for _, v := range scc.Volumes {
		switch v {
		case securityapi.FSTypeHostPath, securityapi.FSTypeAll:
			hasHostVolume = true
			// nothing more to do, this is the max point value
			break
		// it is easier to specifically list the trivial volumes and allow the
		// default case to be non-trivial so we don't have to worry about adding
		// volumes in the future unless they're trivial.
		case securityapi.FSTypeSecret, securityapi.FSTypeConfigMap, securityapi.FSTypeEmptyDir,
			securityapi.FSTypeDownwardAPI, securityapi.FSProjected, securityapi.FSTypeNone:
			// do nothing
		default:
			hasNonTrivialVolume = true
		}
	}

	if hasHostVolume {
		return hostVolumePoints
	}
	if hasNonTrivialVolume {
		return nonTrivialVolumePoints
	}
	return noPoints
}
