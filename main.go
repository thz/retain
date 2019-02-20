// Copyright 2019 Tobias Hintze
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// github.com/thz/retain
// A simple filter to determine expendable named "snapshots" based on
// a retention specification.

package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

const (
	Hourly    = "2006-01-02T15"
	Daily     = "2006-01-02"
	Monthly   = "2006-01"
	Weekly    = "%d-W%02d"
	Yearly    = "2006"
	StdFormat = "2006-01-02T15:04:05MST"
)

type Retention struct {
	alias      string
	timeformat string
	duration   time.Duration
	keep       int
	snaps      map[string]*Snap
}

// Align a point in time with the retention period. Return
// the starting time of the period and the string representation
// of that period.
func (retention Retention) Align(t time.Time) (time.Time, string) {
	if retention.alias == "weekly" {
		// Find the first day of the week.
		start := t.Truncate(time.Hour * 24)
		for {
			y, w := start.ISOWeek()
			ny, nw := start.Add(-time.Hour * 24).ISOWeek()
			if ny*100+nw != y*100+w {
				break
			}
			start = start.Add(-time.Hour * 24)
		}
		y, w := start.ISOWeek()
		return start, fmt.Sprintf(Weekly, y, w)
	}

	// Just truncate to the start of the respective retention period.
	start := t.Truncate(retention.duration)
	return start, start.Format(retention.timeformat)
}

type Snap struct {
	name      string
	timestamp time.Time
}

// snapSorter
type snapSorter struct{ snaps []Snap }

func (sorter *snapSorter) Less(x, y int) bool {
	return sorter.snaps[x].timestamp.UnixNano() < sorter.snaps[y].timestamp.UnixNano()
}
func (sorter *snapSorter) Swap(x, y int) {
	sorter.snaps[x], sorter.snaps[y] = sorter.snaps[y], sorter.snaps[x]
}
func (sorter *snapSorter) Len() int {
	return len(sorter.snaps)
}

func getRetentionsFromSpec(spec string) ([]*Retention, error) {
	retentions := make([]*Retention, 0)
	for _, f := range strings.Fields(spec) {
		if len(f) < 2 {
			return nil, fmt.Errorf("invalid field in retention spec [%s]", f)
		}
		i, err := strconv.Atoi(f[1:])
		if err != nil {
			return nil, fmt.Errorf("invalid field in retention spec [%s]: %v", f, err)
		}
		switch string(f[0]) {
		case "h":
			retentions = append(retentions, &Retention{
				alias: "hourly", duration: time.Hour, timeformat: Hourly, keep: i})
		case "d":
			retentions = append(retentions, &Retention{
				alias: "daily", duration: time.Hour * 24, timeformat: Daily, keep: i})
		case "w":
			retentions = append(retentions, &Retention{
				alias: "weekly", duration: time.Hour * 24 * 7, keep: i})
		case "m":
			retentions = append(retentions, &Retention{
				alias: "monthly", timeformat: Monthly, keep: i})
		case "y":
			retentions = append(retentions, &Retention{
				alias: "yearly", timeformat: Yearly, keep: i})
		}
	}
	return retentions, nil
}

func main() {
	log.SetFormatter(&log.TextFormatter{DisableTimestamp: true})

	var flagRetentionSpec, flagInputFormat string
	flag.StringVar(&flagRetentionSpec, "r", "y5 m24 w8 d21 h72", "the retention specification")
	flag.StringVar(&flagInputFormat, "f", StdFormat, "the time parsing format for input snaps")
	flag.Parse()
	log.Printf("Working with retention spec %q.", flagRetentionSpec)

	retentions, err := getRetentionsFromSpec(flagRetentionSpec)
	if err != nil {
		log.Fatal(err)
	}

	snaps := make([]Snap, 0)
	snapkills := make(map[string]*Snap)
	snapkeeps := make(map[string]*Snap)

	// Scan the stdin for snaps.
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		tline, terr := time.Parse(flagInputFormat, line)
		if terr != nil {
			log.Warnf("Ignoring invalid input snap: %v.", terr)
			continue
		}
		inputsnap := Snap{name: line, timestamp: tline}
		snaps = append(snaps, inputsnap)
		snapkills[line] = &inputsnap
	}

	sorter := &snapSorter{snaps: snaps}
	sort.Sort(sort.Reverse(sorter)) // youngest first
	log.Printf("Working on %d input snaps.", len(snaps))
	for _, snap := range snaps {
		for _, r := range retentions {
			if r.snaps == nil {
				r.snaps = make(map[string]*Snap)
			}
			if len(r.snaps) >= r.keep {
				// Retention already satisfied. No need to keep the snap.
				continue
			}
			_, s := r.Align(snap.timestamp) // map time to retention period
			if _, have := r.snaps[s]; !have {
				newsnap := snap // copy
				r.snaps[s] = &newsnap
				snapkeeps[newsnap.name] = &newsnap
				delete(snapkills, newsnap.name)
			}
		}
	}

	// Log retained snaps.
	for _, r := range retentions {
		log.Printf("Retention %q:\n", r.alias)
		for idx, k := range sortedKeys(r.snaps) {
			log.Printf("  %3d %s -> %s\n", idx, k, r.snaps[k].name)
		}
	}

	// Output snaps which are not kept.
	for _, k := range sortedKeys(snapkills) {
		fmt.Printf("%s\n", snapkills[k].name)
	}
	log.Printf("Releasing %d snaps and keeping %d.", len(snapkills), len(snapkeeps))
}

// Helper function - return the sorted keys of the map.
func sortedKeys(m map[string]*Snap) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)

	}
	sort.Strings(keys)
	return keys
}
