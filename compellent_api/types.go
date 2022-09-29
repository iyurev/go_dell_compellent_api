package compellent_api

type LoginResp struct {
	InstanceId string `json:"instanceId"`
	UserId     int32  `json:"userId"`
}

type FluidNasVolumeFolder struct {
	ClusterId  string `json:"clusterId"`
	FolderId   int    `json:"folderId"`
	Name       string `json:"name"`
	InstanceId string `json:"instanceId"`
}

type FluidFsNasVolume struct {
	ClusterId                    string      `json:"clusterId"`
	Name                         string      `json:"name"`
	Size                         string      `json:"size"`
	NasVolumeFolderId            interface{} `json:"nasVolumeFolderId"`
	AclToUnix777MappingEnabled   bool        `json:"aclToUnix777MappingEnabled"`
	InteroperabilityPolicy       string      `json:"interoperabilityPolicy"`
	DefaultUnixFolderPermissions string      `json:"defaultUnixFolderPermissions"`
	DefaultUnixFilePermissions   string      `json:"defaultUnixFilePermissions"`
	NasVolumeId                  int         `json:"nasVolumeId,omitempty"`
	InstanceId                   string      `json:"instanceId,omitempty"`
}

type FluidFSCluster struct {
	ClusterId    string `json:"clusterId"`
	ObjectType   string `json:"objectType"`
	InstanceName string `json:"instanceName"`
	InstanceId   string `json:"instanceId"`
}

type AccessDetails struct {
	ExportTo        string `json:"exportTo"`
	ExportToClients string `json:"exportToClients"`
	ExportToPrefix  int    `json:"exportToPrefix"`
	ReadWrite       bool   `json:"readWrite"`
	TrustUsers      string `json:"trustUsers"`
}

type FluidFsNfsExport struct {
	ClusterId           string          `json:"clusterId"`
	NasVolumeId         int             `json:"nasVolumeId"`
	FolderPath          string          `json:"folderPath"`
	KerberosV5          bool            `json:"kerberosV5"`
	KerberosV5Integrity bool            `json:"kerberosV5Integrity"`
	KerberosV5Privacy   bool            `json:"kerberosV5Privacy"`
	UnixStyle           bool            `json:"unixStyle"`
	AccessDetails       []AccessDetails `json:"accessDetails"`
	VolumeName          string          `json:"volumeName,omitempty"`
	InstanceId          string          `json:"instanceId"`
}

type Filter struct {
	FilterType string       `json:"filterType"`
	Filters    []FilterItem `json:"filters"`
}

type FilterItem struct {
	AttributeName  string `json:"attributeName"`
	AttributeValue string `json:"attributeValue"`
	FilterType     string `json:"filterType"`
}

type Response struct {
	StatusCode int
	Body       []byte
}
