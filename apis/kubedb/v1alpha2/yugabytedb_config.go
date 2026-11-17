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

package v1alpha2

import (
	"bufio"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"kubedb.dev/apimachinery/apis/kubedb"
)

// YugabyteDBConfigKeys are the only keys accepted in spec.configuration (secret or inline) and in the
// applyConfig of a Reconfigure OpsRequest.
var YugabyteDBConfigKeys = []string{
	kubedb.YugabyteDBMasterFlagsKey,
	kubedb.YugabyteDBTServerFlagsKey,
	kubedb.YugabyteDBPGConfKey,
	kubedb.YugabyteDBHBAConfKey,
}

// yugabyteDBOperatorOwnedFlags are the flags the operator writes itself and re-asserts after the user
// configuration. A user value for any of them is rejected. Flags with the prefix
// yugabyteDBOperatorOwnedFlagPrefix are rejected too.
var yugabyteDBOperatorOwnedFlags = map[string]struct{}{
	"rpc_bind_addresses":              {},
	"server_broadcast_addresses":      {},
	"master_addresses":                {},
	"tserver_master_addrs":            {},
	"fs_data_dirs":                    {},
	"certs_dir":                       {},
	"certs_for_client_dir":            {},
	"cert_node_filename":              {},
	"use_node_to_node_encryption":     {},
	"use_client_to_server_encryption": {},
	"allow_insecure_connections":      {},
	"ysql_enable_auth":                {},
	"use_cassandra_authentication":    {},
	"memory_limit_hard_bytes":         {},
}

const yugabyteDBOperatorOwnedFlagPrefix = "placement_"

var (
	yugabyteDBFlagNameRE = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)
	yugabyteDBPGKeyRE    = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_.]*$`)
)

// IsYugabyteDBOperatorOwnedFlag reports whether the operator owns the flag.
func IsYugabyteDBOperatorOwnedFlag(name string) bool {
	if _, ok := yugabyteDBOperatorOwnedFlags[name]; ok {
		return true
	}
	return strings.HasPrefix(name, yugabyteDBOperatorOwnedFlagPrefix)
}

// YugabyteDBOperatorOwnedFlags returns the sorted names of the flags the operator owns (not the placement_ prefix).
func YugabyteDBOperatorOwnedFlags() []string {
	names := make([]string, 0, len(yugabyteDBOperatorOwnedFlags))
	for n := range yugabyteDBOperatorOwnedFlags {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}

// ParseYugabyteDBFlags parses the content of a master.flags or tserver.flags value. Each non empty line that
// does not start with # is "--name=value" or "name=value". It returns the flags in order of appearance.
func ParseYugabyteDBFlags(content string) ([][2]string, error) {
	var flags [][2]string
	sc := bufio.NewScanner(strings.NewReader(content))
	sc.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)
	lineNo := 0
	for sc.Scan() {
		lineNo++
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		line = strings.TrimPrefix(line, "--")
		name, value, found := strings.Cut(line, "=")
		name = strings.TrimSpace(name)
		if !found {
			return nil, fmt.Errorf("line %d: %q is not in the form --name=value or name=value", lineNo, sc.Text())
		}
		if !yugabyteDBFlagNameRE.MatchString(name) {
			return nil, fmt.Errorf("line %d: %q is not a valid flag name", lineNo, name)
		}
		flags = append(flags, [2]string{name, strings.TrimSpace(value)})
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	return flags, nil
}

// ValidateYugabyteDBConfigEntry validates the value of one configuration key.
func ValidateYugabyteDBConfigEntry(key, content string) error {
	switch key {
	case kubedb.YugabyteDBMasterFlagsKey, kubedb.YugabyteDBTServerFlagsKey:
		flags, err := ParseYugabyteDBFlags(content)
		if err != nil {
			return fmt.Errorf("%s: %v", key, err)
		}
		for _, f := range flags {
			if IsYugabyteDBOperatorOwnedFlag(f[0]) {
				return fmt.Errorf("%s: flag %q is owned by the operator and cannot be set", key, f[0])
			}
		}
	case kubedb.YugabyteDBPGConfKey:
		sc := bufio.NewScanner(strings.NewReader(content))
		for sc.Scan() {
			line := strings.TrimSpace(sc.Text())
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			name := strings.FieldsFunc(line, func(r rune) bool { return r == '=' || r == ' ' || r == '\t' })
			if len(name) < 2 || !yugabyteDBPGKeyRE.MatchString(name[0]) {
				return fmt.Errorf("%s: %q is not in the form name = value", key, line)
			}
		}
	case kubedb.YugabyteDBHBAConfKey:
		sc := bufio.NewScanner(strings.NewReader(content))
		for sc.Scan() {
			line := strings.TrimSpace(sc.Text())
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			if len(strings.Fields(line)) < 4 {
				return fmt.Errorf("%s: %q is not a valid pg_hba.conf rule", key, line)
			}
		}
	default:
		return fmt.Errorf("unknown configuration key %q, allowed keys are %s", key, strings.Join(YugabyteDBConfigKeys, ", "))
	}
	return nil
}

// ValidateYugabyteDBConfigData validates every entry of a configuration map. The errors are sorted by key.
func ValidateYugabyteDBConfigData(data map[string]string) []error {
	keys := make([]string, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var errs []error
	for _, k := range keys {
		if err := ValidateYugabyteDBConfigEntry(k, data[k]); err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}
