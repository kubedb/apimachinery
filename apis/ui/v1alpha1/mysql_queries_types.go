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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	ResourceKindMySQLQueries = "MySQLQueries"
	ResourceMySQLQueries     = "mysqlqueries"
	ResourceMySQLQuerieses   = "mysqlqueries"
)

// MySQLQueriesSpec defines the desired state of MySQLQueries
type MySQLQueriesSpec struct {
	Queries []MySQLQuerySpec `json:"queries" protobuf:"bytes,1,rep,name=queries"`
}

type MySQLQuerySpec struct {
	StartTime             *metav1.Time `json:"startTime" protobuf:"bytes,1,opt,name=startTime"`
	UserHost              string       `json:"userHost" protobuf:"bytes,2,opt,name=userHost"`
	QueryTimeMilliSeconds string       `json:"queryTimeMilliSeconds" protobuf:"bytes,3,opt,name=queryTimeMilliSeconds"`
	LockTimeMilliSeconds  string       `json:"lockTimeMilliSeconds" protobuf:"bytes,4,opt,name=lockTimeMilliSeconds"`
	RowsSent              int64        `json:"rowsSent" protobuf:"varint,5,opt,name=rowsSent"`
	RowsExamined          int64        `json:"rowsExamined" protobuf:"varint,6,opt,name=rowsExamined"`
	DB                    string       `json:"db" protobuf:"bytes,7,opt,name=db"`
	LastInsertId          int64        `json:"lastInsertId,omitempty" protobuf:"varint,8,opt,name=lastInsertId"`
	InsertId              int64        `json:"insertId,omitempty" protobuf:"varint,9,opt,name=insertId"`
	ServerId              int64        `json:"serverId,omitempty" protobuf:"varint,10,opt,name=serverId"`
	SQLText               string       `json:"sqlText,omitempty" protobuf:"bytes,11,opt,name=sqlText"`
	ThreadId              int64        `json:"threadId,omitempty" protobuf:"varint,12,opt,name=threadId"`
}

// MySQLQueries is the Schema for the MySQLQueries API

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type MySQLQueries struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`

	Spec MySQLQueriesSpec `json:"spec,omitempty" protobuf:"bytes,2,opt,name=spec"`
}

// MySQLQueriesList contains a list of MySQLQueries

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type MySQLQueriesList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty" protobuf:"bytes,1,opt,name=metadata"`
	Items           []MySQLQueries `json:"items" protobuf:"bytes,2,rep,name=items"`
}

func init() {
	SchemeBuilder.Register(&MySQLQueries{}, &MySQLQueriesList{})
}
