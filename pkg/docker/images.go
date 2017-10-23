package docker

const (
	ImageOperator         = "kubedb/operator"
	ImagePostgresOperator = "kubedb/pg-operator"
	ImagePostgres         = "kubedb/postgres"
	ImageMySQLOperator    = "kubedb/mysql-operator"
	ImageMySQL            = "library/mysql"
	ImageElasticOperator  = "kubedb/es-operator"
	ImageElasticsearch    = "kubedb/elasticsearch"
	ImageElasticdump      = "kubedb/elasticdump"
	ImageXdbOperator      = "kubedb/xdb-operator"
	ImageXdb              = "kubedb/xdb"
)

const (
	OperatorName       = "kubedb-operator"
	OperatorContainer  = "operator"
	OperatorPortName   = "web"
	OperatorPortNumber = 8080
)
