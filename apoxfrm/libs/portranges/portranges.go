package portranges

import (
	"fmt"
	"strconv"

	"github.com/PaloAltoNetworks/cns-customer/apoxfrm/libs/portspec"
)

// TrimPortRange returns ranges such that if no entries in exist in filteredPortMap, the
// complete sports are returned. However, if filteredPortMap has entries, the ranges
// returned are intersection of sports and filteredPortMap.
func TrimPortRange(sports string, filteredPortMap map[int]struct{}) ([]string, error) {

	// return early if there are no ports in policy
	// remove this when we remove ports from ext networks
	if len(filteredPortMap) == 0 {
		return []string{sports}, nil
	}

	pspec, err := portspec.NewPortSpecFromString(sports, nil)
	if err != nil {
		return []string{}, err
	}

	// single value
	if pspec.Min == pspec.Max {
		if _, ok := filteredPortMap[int(pspec.Min)]; ok {
			return []string{sports}, nil
		}
		return []string{}, nil
	}

	// range
	includePorts := []int{}
	for i := uint32(pspec.Min); i <= uint32(pspec.Max); i++ {
		if _, ok := filteredPortMap[int(i)]; !ok {
			continue
		}
		includePorts = append(includePorts, int(i))
	}

	return buildRanges(includePorts), nil
}

// CreatePortList take a map of of ports mentioned in the port clause of network policy and
// converts it into a port list
func CreatePortList(portMap map[string]interface{}) []string {
	portList := make([]string, len(portMap))
	var sliceIndex = 0
	for k := range portMap {
		portList[sliceIndex] = k
		sliceIndex++
	}
	return portList
}

// fmtRange will return a string array with one member in the form
// - { "start:end" } when start and end are two different numbers
// - { "start" } when start and end are same number
func fmtRange(start, end int) []string {
	var r string
	if start != end {
		r = fmt.Sprintf("%d:%d", start, end)
	} else {
		r = strconv.Itoa(start)
	}
	return []string{r}
}

// buildRangesR is a recursive function to return a list of ranges
// representing the numbers in the ports list.
// ports list is expected to be sorted in ascending order.
func buildRangesR(ports []int, start, curr int) []string {
	if len(ports) == 1 {
		return fmtRange(start, curr)
	}
	// len(ports) > 1
	sports := []string{}
	if ports[0]+1 != ports[1] {
		sports = fmtRange(start, curr)
		start = ports[1]
	}
	return append(sports, buildRangesR(ports[1:], start, ports[1])...)
}

// buildRanges returns a list of ranges to represent ports in the ports list.
// ports list is expected to be sorted in ascending order.
func buildRanges(ports []int) []string {
	if len(ports) == 0 {
		return []string{}
	}
	return buildRangesR(ports, ports[0], ports[0])
}
