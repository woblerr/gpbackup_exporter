package gpbckpstruct

type History struct {
	BackupConfigs []BackupConfig `yaml:"backupconfigs"`
}

type BackupConfig struct {
	BackupDir             string             `yaml:"backupdir"`
	BackupVersion         string             `yaml:"backupversion"`
	Compressed            bool               `yaml:"compressed"`
	CompressionType       string             `yaml:"compressiontype"`
	DatabaseName          string             `yaml:"databasename"`
	DatabaseVersion       string             `yaml:"databaseversion"`
	DataOnly              bool               `yaml:"dataonly"`
	DateDeleted           string             `yaml:"datedeleted"`
	ExcludeRelations      []string           `yaml:"excluderelations"`
	ExcludeSchemaFiltered bool               `yaml:"excludeschemafiltered"`
	ExcludeSchemas        []string           `yaml:"excludeschemas"`
	ExcludeTableFiltered  bool               `yaml:"excludetablefiltered"`
	IncludeRelations      []string           `yaml:"includerelations"`
	IncludeSchemaFiltered bool               `yaml:"includeschemafiltered"`
	IncludeSchemas        []string           `yaml:"includeschemas"`
	IncludeTableFiltered  bool               `yaml:"includetablefiltered"`
	Incremental           bool               `yaml:"incremental"`
	LeafPartitionData     bool               `yaml:"leafpartitiondata"`
	MetadataOnly          bool               `yaml:"metadataonly"`
	Plugin                string             `yaml:"plugin"`
	PluginVersion         string             `yaml:"pluginversion"`
	RestorePlan           []RestorePlanEntry `yaml:"restoreplan"`
	SingleDataFile        bool               `yaml:"singledatafile"`
	Timestamp             string             `yaml:"timestamp"`
	EndTime               string             `yaml:"endtime"`
	WithoutGlobals        bool               `yaml:"withoutgoals"`
	WithStatistics        bool               `yaml:"withstatistics"`
	Status                string             `yaml:"status"`
}

type RestorePlanEntry struct {
	Timestamp string   `yaml:"timestamp"`
	TableFQNs []string `yaml:"tablefqdn"`
}
