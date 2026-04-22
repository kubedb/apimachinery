package v1alpha1

type MongoSource struct {
	ConnectionInfo ConnectionInfo `yaml:"connectionInfo" json:"connectionInfo"`
	Mongoshake     *Mongoshake    `yaml:"mongoshake" json:"mongoshake,omitempty"`
}
type MongoTarget struct {
	ConnectionInfo ConnectionInfo `yaml:"connectionInfo" json:"connectionInfo"`
}
type Mongoshake struct {
	// SyncMode: full, incr, or fullSync
	SyncMode string `yaml:"syncMode" json:"syncMode,omitempty" config:"sync_mode"`

	MongoSslRootCaFile   string `yaml:"mongoSslRootCaFile" json:"mongoSslRootCaFile,omitempty" config:"mongo_ssl_root_ca_file"`
	MongoSslClientCaFile string `yaml:"mongoSslClientCaFile" json:"mongoSslClientCaFile,omitempty" config:"mongo_ssl_root_ca_file"`

	FilterOpTypes        []string `yaml:"filterOpTypes" json:"filterOpTypes,omitempty" config:"filter.op_types"`
	FilterNamespaceBlack []string `yaml:"filterNamespaceBlack" json:"filterNamespaceBlack,omitempty" config:"filter.namespace.black"`
	FilterNamespaceWhite []string `yaml:"filterNamespaceWhite" json:"filterNamespaceWhite,omitempty" config:"filter.namespace.white"`
	FilterPassSpecialDb  []string `yaml:"filterPassSpecialDb" json:"filterPassSpecialDb,omitempty" config:"filter.pass.special.db"`

	// ---------------- BOOLS → POINTERS ----------------
	FilterDDLEnable *bool `yaml:"filterDdlEnable" json:"filterDdlEnable,omitempty" config:"filter.ddl_enable"`
	FilterOplogGids *bool `yaml:"filterOplogGids" json:"filterOplogGids,omitempty" config:"filter.oplog.gids"`

	CheckpointStartPosition int64 `yaml:"checkpointStartPosition" json:"checkpointStartPosition,omitempty" config:"checkpoint.start_position" type:"date"`

	TransformNamespace []string `yaml:"transformNamespace" json:"transformNamespace,omitempty" config:"transform.namespace"`

	FullSyncReaderCollectionParallel    int `yaml:"fullSyncReaderCollectionParallel" json:"fullSyncReaderCollectionParallel,omitempty" config:"full_sync.reader.collection_parallel"`
	FullSyncReaderWriteDocumentParallel int `yaml:"fullSyncReaderWriteDocumentParallel" json:"fullSyncReaderWriteDocumentParallel,omitempty" config:"full_sync.reader.write_document_parallel"`
	FullSyncReaderDocumentBatchSize     int `yaml:"fullSyncReaderDocumentBatchSize" json:"fullSyncReaderDocumentBatchSize,omitempty" config:"full_sync.reader.document_batch_size"`
	FullSyncReaderFetchBatchSize        int `yaml:"fullSyncReaderFetchBatchSize" json:"fullSyncReaderFetchBatchSize,omitempty" config:"full_sync.reader.fetch_batch_size"`
	FullSyncReaderParallelThread        int `yaml:"fullSyncReaderParallelThread" json:"fullSyncReaderParallelThread,omitempty" config:"full_sync.reader.parallel_thread"`

	FullSyncReaderParallelIndex     string `yaml:"fullSyncReaderParallelIndex" json:"fullSyncReaderParallelIndex,omitempty" config:"full_sync.reader.parallel_index"`
	FullSyncReaderSplitMaxChunkSize int    `yaml:"fullSyncReaderSplitMaxChunkSize" json:"fullSyncReaderSplitMaxChunkSize,omitempty" config:"full_sync.reader.split_max_chunk_size"`

	// ---------------- BOOLS → POINTERS ----------------
	FullSyncCollectionDrop               *bool `yaml:"fullSyncCollectionDrop" json:"fullSyncCollectionDrop,omitempty" config:"full_sync.collection_exist_drop"`
	FullSyncReaderOplogStoreDisk         *bool `yaml:"fullSyncReaderOplogStoreDisk" json:"fullSyncReaderOplogStoreDisk,omitempty" config:"full_sync.reader.oplog_store_disk"`
	FullSyncExecutorInsertOnDupUpdate    *bool `yaml:"fullSyncExecutorInsertOnDupUpdate" json:"fullSyncExecutorInsertOnDupUpdate,omitempty" config:"full_sync.executor.insert_on_dup_update"`
	FullSyncExecutorFilterOrphanDocument *bool `yaml:"fullSyncExecutorFilterOrphanDocument" json:"fullSyncExecutorFilterOrphanDocument,omitempty" config:"full_sync.executor.filter.orphan_document"`
	FullSyncExecutorMajorityEnable       *bool `yaml:"fullSyncExecutorMajorityEnable" json:"fullSyncExecutorMajorityEnable,omitempty" config:"full_sync.executor.majority_enable"`
	FullSyncDoNotShardDest               *bool `yaml:"fullSyncDoNotShardDest" json:"fullSyncDoNotShardDest,omitempty" config:"full_sync.do_not_shard_destination"`

	FullSyncCreateIndex                 string `yaml:"fullSyncCreateIndex" json:"fullSyncCreateIndex,omitempty" config:"full_sync.create_index"`
	FullSyncReaderOplogStoreDiskMaxSize int64  `yaml:"fullSyncReaderOplogStoreDiskMaxSize" json:"fullSyncReaderOplogStoreDiskMaxSize,omitempty" config:"full_sync.reader.oplog_store_disk_max_size"`

	IncrSyncReaderFetchBatchSize int `yaml:"incrSyncReaderFetchBatchSize" json:"incrSyncReaderFetchBatchSize,omitempty" config:"incr_sync.reader.fetch_batch_size"`
	IncrSyncWorker               int `yaml:"incrSyncWorker" json:"incrSyncWorker,omitempty" config:"incr_sync.worker"`
	IncrSyncTunnelWriteThread    int `yaml:"incrSyncTunnelWriteThread" json:"incrSyncTunnelWriteThread,omitempty" config:"incr_sync.tunnel.write_thread"`

	// +optional
	ExtraConfiguration map[string]string `yaml:"extraConfiguration" json:"extraConfiguration,omitempty"`
}
