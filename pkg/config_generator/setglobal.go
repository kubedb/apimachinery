/*
Copyright AppsCode Inc. and Contributors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package configgenerator

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

var (
	setGlobalName   = regexp.MustCompile(`^[a-z0-9_]+$`)
	setGlobalNumber = regexp.MustCompile(`^-?\d+(\.\d+)?$`)
	setGlobalSize   = regexp.MustCompile(`^(\d+)([kKmMgGtT])$`)
	setGlobalQuote  = strings.NewReplacer(`\`, `\\`, `'`, `\'`)
	setGlobalUnits  = map[string]int64{"k": 1 << 10, "m": 1 << 20, "g": 1 << 30, "t": 1 << 40}
)

// BuildSetGlobal turns MySQL/MariaDB cnf files into one atomic "SET GLOBAL k=v, ..." statement.
// Sources are applied in order, so later ones win. It returns "" when there are no variables.
// A bare flag (no "=") cannot be set live, so it is an error.
func BuildSetGlobal(sources ...map[string]string) (string, error) {
	var assigns []string
	for _, files := range sources {
		names := make([]string, 0, len(files))
		for name := range files {
			names = append(names, name)
		}
		sort.Strings(names)

		for _, name := range names {
			for _, line := range strings.Split(files[name], "\n") {
				line = strings.TrimSpace(line)
				if line == "" || line[0] == '#' || line[0] == '[' || strings.Contains(line, "=") {
					continue
				}
				return "", fmt.Errorf("%s: %q is not a key=value variable and needs a restart", name, line)
			}
			cfg := ConvertStringInToMap(files[name], []string{"="})
			for _, key := range cfg.Keys() {
				obj, _ := cfg.Get(key)
				variable := strings.ReplaceAll(strings.ToLower(key), "-", "_")
				if !setGlobalName.MatchString(variable) {
					return "", fmt.Errorf("%s: invalid variable name %q", name, key)
				}
				assigns = append(assigns, variable+"="+setGlobalLiteral(obj.(*ValueGenerator).Value))
			}
		}
	}
	if len(assigns) == 0 {
		return "", nil
	}
	return "SET GLOBAL " + strings.Join(assigns, ", "), nil
}

// setGlobalLiteral follows what SET GLOBAL accepts: numbers and ON/OFF unquoted, 1M-style sizes
// expanded to bytes, everything else a quoted string.
func setGlobalLiteral(v string) string {
	v = strings.Trim(strings.TrimSpace(v), `"'`)
	if m := setGlobalSize.FindStringSubmatch(v); m != nil {
		if n, err := strconv.ParseInt(m[1], 10, 64); err == nil {
			return strconv.FormatInt(n*setGlobalUnits[strings.ToLower(m[2])], 10)
		}
	}
	switch strings.ToUpper(v) {
	case "ON", "OFF", "TRUE", "FALSE":
		return v
	}
	if setGlobalNumber.MatchString(v) {
		return v
	}
	return "'" + setGlobalQuote.Replace(v) + "'"
}
