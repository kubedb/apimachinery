package docker

const (
	ImageOperator              = "kubedb/operator"
	ImagePostgresOperator      = "kubedb/pg-operator"
	ImagePostgres              = "kubedb/postgres"
	ImagePostgresTools         = "kubedb/postgres-tools"
	ImageElasticsearchOperator = "kubedb/es-operator"
	ImageElasticsearch         = "kubedb/elasticsearch"
	ImageElasticsearchTools    = "kubedb/elasticsearch-tools"
	ImageMySQLOperator         = "kubedb/ms-operator"
	ImageMySQL                 = "kubedb/mysql"
	ImageMySQLTools            = "kubedb/mysql-tools"
	ImageMongoDBOperator       = "kubedb/mg-operator"
	ImageMongoDB               = "kubedb/mongo"
	ImageMongoDBTools          = "kubedb/mongo-tools"
	ImageRedisOperator         = "kubedb/rd-operator"
	ImageRedis                 = "kubedb/redis"
	ImageMemcachedOperator     = "kubedb/mc-operator"
	ImageMemcached             = "kubedb/memcached"
)

const (
	OperatorName       = "kubedb-operator"
	OperatorContainer  = "operator"
	OperatorPortName   = "web"
	OperatorPortNumber = 8080
)
