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
	"database/sql"
	"strconv"
	"strings"

	"kubedb.dev/apimachinery/crds"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	"kmodules.xyz/client-go/apiextensions"
)

func (in MySQLDatabase) CustomResourceDefinition() *apiextensions.CustomResourceDefinition {
	return crds.MustCustomResourceDefinition(GroupVersion.WithResource(ResourceMySQLDatabases))
}

var _ Interface = &MySQLDatabase{}

func (in *MySQLDatabase) GetInit() *InitSpec {
	return in.Spec.Init
}

func (in *MySQLDatabase) GetStatus() DatabaseStatus {
	return in.Status
}

//GetAppBindingMeta returns meta info of the appbinding which has been created by schema manager
func (in *MySQLDatabase) GetAppBindingMeta() metav1.ObjectMeta {
	meta := metav1.ObjectMeta{
		Name:      in.Name + "-appbinding",
		Namespace: in.Namespace,
	}
	return meta
}

//GetVaultSecretEngineMeta returns meta info of the secret engine which has been created by schema manager
func (in *MySQLDatabase) GetVaultSecretEngineMeta() metav1.ObjectMeta {
	meta := metav1.ObjectMeta{
		Name:      in.Name + "-secret-engine",
		Namespace: in.Namespace,
	}
	return meta
}

//GetMySQLRoleMeta returns meta info of the MySQL role which has been created by schema manager
func (in *MySQLDatabase) GetMySQLRoleMeta() metav1.ObjectMeta {
	meta := metav1.ObjectMeta{
		Name:      in.Name + "-mysql-role",
		Namespace: in.Namespace,
	}
	return meta
}

//GetSecretAccessRequestMeta returns meta info of the secret access request which has been created by schema manager
func (in *MySQLDatabase) GetSecretAccessRequestMeta() metav1.ObjectMeta {
	meta := metav1.ObjectMeta{
		Name:      in.Name + "-secret-access-request",
		Namespace: in.Namespace,
	}
	return meta
}

//GetInitJobMeta returns meta info of the init job which has been created by schema manager
func (in *MySQLDatabase) GetInitJobMeta() metav1.ObjectMeta {
	meta := metav1.ObjectMeta{
		Name:      in.Name + "-init-job",
		Namespace: in.Namespace,
	}
	return meta
}

//GetMySQLAuthSecretMeta returns meta info of the mysql auth secret
func (in *MySQLDatabase) GetMySQLAuthSecretMeta() metav1.ObjectMeta {
	meta := metav1.ObjectMeta{
		Name:      in.Spec.DatabaseRef.Name + "-auth",
		Namespace: in.Spec.DatabaseRef.Namespace,
	}
	return meta
}

//GetRestoreSessionMeta returns meta info of the restore session which has been created by schema manager
func (in *MySQLDatabase) GetRestoreSessionMeta() metav1.ObjectMeta {
	meta := metav1.ObjectMeta{
		Name:      in.Name + "-restoresession",
		Namespace: in.Namespace,
	}
	return meta
}

//GetRepositoryMeta returns meta info of the repository which has been created by schema manager
func (in *MySQLDatabase) GetRepositoryMeta() metav1.ObjectMeta {
	meta := metav1.ObjectMeta{
		Name:      in.Name + "-repository",
		Namespace: in.Namespace,
	}
	return meta
}

//GetRepositorySecretMeta returns meta info of the repository which has been created by schema manager
func (in *MySQLDatabase) GetRepositorySecretMeta() metav1.ObjectMeta {
	meta := metav1.ObjectMeta{
		Name:      in.Name + "-repository-secret",
		Namespace: in.Namespace,
	}
	return meta
}

//======================================================database functions=====================================

const (
	SHOW              string = " SHOW "
	CREATE            string = " CREATE "
	DROP              string = " DROP "
	ALTER             string = " ALTER "
	DATABASE          string = " DATABASE "
	USE               string = " USE "
	SPACE             string = " "
	SEMICOLON         string = ";"
	CHARACTERSET      string = " CHARACTER SET "
	COLLATE           string = " COLLATE "
	READONLY          string = " READ ONLY "
	ENCRYPTION        string = " ENCRYPTION "
	IFNOTEXISTS       string = " IF NOT EXISTS "
	IFEXISTS          string = " IF EXISTS "
	NULLSTRING        string = ""
	ENCRYPTIONENABLE  string = "'Y'"
	ENCRYPTIONDISABLE string = "'N'"
)

//CreateDatabase function creates database in the server as per the configuration of d database
func (d *MySQLDatabaseConfiguration) CreateDatabase(cl *sql.DB) error {

	//make queryString string ready
	queryString := CREATE + DATABASE + IFNOTEXISTS + d.Name
	if d.CharacterSet != NULLSTRING {
		queryString += CHARACTERSET + d.CharacterSet
	}
	if d.Collation != NULLSTRING {
		queryString += COLLATE + d.Collation
	}
	if d.Encryption != NULLSTRING {
		queryString += ENCRYPTION + d.Encryption
	}
	queryString += SEMICOLON

	// execute queryString
	_, err := cl.Exec(queryString)

	// handle error
	if err != nil {
		klog.Errorf("Error while creating database\n")
		return err
	}
	klog.V(3).Infof("Database created : %s\n", d.Name)
	return nil
}

//DeleteDatabase drops the database d from the server
func (d *MySQLDatabaseConfiguration) DeleteDatabase(cl *sql.DB) error {

	//make queryString string ready
	queryString := DROP + DATABASE + IFEXISTS + d.Name + SEMICOLON

	// execute queryString
	_, err := cl.Exec(queryString)

	// handle error
	if err != nil {
		klog.Errorf("Error while dropping database\n")
		return err
	}
	klog.V(3).Infof("Dropped database : %s\n", d.Name)
	return nil
}

//AlterDatabase alters the existing database of the same name with the updated configuration
func (d *MySQLDatabaseConfiguration) AlterDatabase(cl *sql.DB) error {

	curDatabase, err := GetDatabase(d.Name, cl)

	if err != nil {
		klog.Errorf("Error fetching database\n")
		return err
	}

	//make queryString string ready
	queryString := ""
	if d.CharacterSet != curDatabase.CharacterSet && d.CharacterSet != "" {
		queryString += CHARACTERSET + d.CharacterSet
	}
	if d.Collation != curDatabase.Collation && d.Collation != "" {
		queryString += COLLATE + d.Collation
	}
	if d.Encryption != curDatabase.Encryption && d.Encryption != "" {
		queryString += ENCRYPTION + d.Encryption
	}
	if d.ReadOnly != curDatabase.ReadOnly {
		queryString += READONLY + strconv.Itoa(int(d.ReadOnly))
	}
	if queryString == "" {
		return nil
	}
	queryString = ALTER + DATABASE + d.Name + queryString + SEMICOLON

	//execute queryString
	klog.Infof("Altering database : ", queryString)
	_, err = cl.Exec(queryString)

	// handle error
	if err != nil {
		klog.Errorf("Error while altering database\n")
		return err
	}
	klog.V(3).Infof("Database altered : %s\n", d.Name)
	return nil
}

//GetDatabase fetches database with the name provided and maps it into a database structure
func GetDatabase(name string, cl *sql.DB) (ret MySQLDatabaseConfiguration, err error) {

	//make query string ready
	query := SHOW + CREATE + DATABASE + name + SEMICOLON

	//execute query
	res, err := cl.Query(query)

	//handle error
	if err != nil {
		klog.Errorf("Error fetching database\n")
		return ret, err
	}

	//process
	for res.Next() {
		var dbname, retquery string
		res.Scan(&dbname, &retquery)
		ret.Name = dbname
		split := strings.Split(retquery, " ")
		pre1 := ""
		pre2 := ""
		pre3 := ""
		for _, s := range split {
			if pre2+SPACE+pre1 == "CHARACTER SET" {
				ret.CharacterSet = s
			}
			if pre1 == "COLLATE" {
				ret.Collation = s
			}
			if pre3+SPACE+pre2+SPACE+pre1 == "READ ONLY =" {
				x, _ := strconv.Atoi(s)
				ret.ReadOnly = int32(x)
			}
			if s == "ENCRYPTION='N'" {
				ret.Encryption = "'N'"
			}
			if s == "ENCRYPTION='Y'" {
				ret.Encryption = "'Y'"
			}
			pre3 = pre2
			pre2 = pre1
			pre1 = s
		}
	}
	return ret, nil
}
