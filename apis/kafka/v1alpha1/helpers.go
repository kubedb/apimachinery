package v1alpha1

import (
	dbapi "kubedb.dev/apimachinery/apis/kubedb/v1"

	core "k8s.io/api/core/v1"
)

func copyConfigurationField(cnf *dbapi.ConfigurationSpec, sec *core.LocalObjectReference) *dbapi.ConfigurationSpec {
	if sec != nil {
		if cnf == nil {
			cnf = &dbapi.ConfigurationSpec{}
		}
		cnf.SecretName = sec.Name
	}
	sec = nil
	return cnf
}
