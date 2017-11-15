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
	ImageMongoDBOperator  = "kubedb/mongodb-operator"
	ImageMongoDB          = "library/mongo"
<<<<<<< HEAD
	ImageRedisOperator    = "kubedb/mongodb-operator"
	ImageRedis            = "library/redis"
=======
>>>>>>> 02805cad028e70c6eb403b343ff09113ae3d9f9f
)

const (
	OperatorName       = "kubedb-operator"
	OperatorContainer  = "operator"
	OperatorPortName   = "web"
	OperatorPortNumber = 8080
)
