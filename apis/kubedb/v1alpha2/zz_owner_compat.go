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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Owner is retained as a thin alias for AsOwner. The accessor was renamed in
// apimachinery, but the released database operators still call Owner(), so
// dropping it outright breaks every consumer that vendors them at a pinned
// version. Delete this file once those operators have migrated to AsOwner.

// Deprecated: use AsOwner instead.
func (c *Cassandra) Owner() *metav1.OwnerReference {
	return c.AsOwner()
}

// Deprecated: use AsOwner instead.
func (c *ClickHouse) Owner() *metav1.OwnerReference {
	return c.AsOwner()
}

// Deprecated: use AsOwner instead.
func (d *DB2) Owner() *metav1.OwnerReference {
	return d.AsOwner()
}

// Deprecated: use AsOwner instead.
func (d *DocumentDB) Owner() *metav1.OwnerReference {
	return d.AsOwner()
}

// Deprecated: use AsOwner instead.
func (d *Druid) Owner() *metav1.OwnerReference {
	return d.AsOwner()
}

// Deprecated: use AsOwner instead.
func (h *HanaDB) Owner() *metav1.OwnerReference {
	return h.AsOwner()
}

// Deprecated: use AsOwner instead.
func (h *Hazelcast) Owner() *metav1.OwnerReference {
	return h.AsOwner()
}

// Deprecated: use AsOwner instead.
func (i *Ignite) Owner() *metav1.OwnerReference {
	return i.AsOwner()
}

// Deprecated: use AsOwner instead.
func (m *MSSQLServer) Owner() *metav1.OwnerReference {
	return m.AsOwner()
}

// Deprecated: use AsOwner instead.
func (m *Milvus) Owner() *metav1.OwnerReference {
	return m.AsOwner()
}

// Deprecated: use AsOwner instead.
func (n *Neo4j) Owner() *metav1.OwnerReference {
	return n.AsOwner()
}

// Deprecated: use AsOwner instead.
func (o *Oracle) Owner() *metav1.OwnerReference {
	return o.AsOwner()
}

// Deprecated: use AsOwner instead.
func (p *Pgpool) Owner() *metav1.OwnerReference {
	return p.AsOwner()
}

// Deprecated: use AsOwner instead.
func (q *Qdrant) Owner() *metav1.OwnerReference {
	return q.AsOwner()
}

// Deprecated: use AsOwner instead.
func (r *RabbitMQ) Owner() *metav1.OwnerReference {
	return r.AsOwner()
}

// Deprecated: use AsOwner instead.
func (s *Singlestore) Owner() *metav1.OwnerReference {
	return s.AsOwner()
}

// Deprecated: use AsOwner instead.
func (w *Weaviate) Owner() *metav1.OwnerReference {
	return w.AsOwner()
}

// Deprecated: use AsOwner instead.
func (z *ZooKeeper) Owner() *metav1.OwnerReference {
	return z.AsOwner()
}
