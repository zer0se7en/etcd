// Copyright 2021 The etcd Authors
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

package version

import (
	"fmt"

	"github.com/coreos/go-semver/semver"
	"go.uber.org/zap"

	"go.etcd.io/etcd/server/v3/mvcc/backend"
	"go.etcd.io/etcd/server/v3/mvcc/buckets"
)

var (
	V3_5 = semver.Version{Major: 3, Minor: 5}
	V3_6 = semver.Version{Major: 3, Minor: 6}
)

// UpdateStorageVersion updates storage version.
func UpdateStorageVersion(lg *zap.Logger, tx backend.BatchTx) error {
	tx.Lock()
	defer tx.Unlock()
	v, err := detectStorageVersion(lg, tx)
	if err != nil {
		return fmt.Errorf("cannot determine storage version: %w", err)
	}
	switch *v {
	case V3_5:
		lg.Warn("setting storage version", zap.String("storage-version", V3_6.String()))
		// All meta keys introduced in v3.6 should be filled in here.
		buckets.UnsafeSetStorageVersion(tx, &V3_6)
	case V3_6:
	default:
		lg.Warn("unknown storage version", zap.String("storage-version", v.String()))
	}
	return nil
}

func detectStorageVersion(lg *zap.Logger, tx backend.ReadTx) (*semver.Version, error) {
	v := buckets.UnsafeReadStorageVersion(tx)
	if v != nil {
		return v, nil
	}
	confstate := buckets.UnsafeConfStateFromBackend(lg, tx)
	if confstate == nil {
		return nil, fmt.Errorf("missing %q key", buckets.MetaConfStateName)
	}
	_, term := buckets.UnsafeReadConsistentIndex(tx)
	if term == 0 {
		return nil, fmt.Errorf("missing %q key", buckets.MetaTermKeyName)
	}
	copied := V3_5
	return &copied, nil
}
