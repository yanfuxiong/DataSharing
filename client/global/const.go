package global

const (
	ClientVersion = "2.3.39"

	ProtocolID     = "/libp2p/dcutr"
	HostProtocolID = "host_register"

	HOST_ID                   = "12345" // This HOST ID is pseudo to test
	ProtocolDirectID          = "/instruction/cross_share/1.0.0"
	ProtocolImageTransmission = "/ipfs/protocol/cross_share/1.0.0"
	ProtocolFileTransmission  = "/ipfs/protocol/cross_share/1.0.1"
	DefaultPort               = 0

	// This is the maximum length of messages between clients,    32KB
	P2PMsgMaxLength = 32 * 1024

	//This is the length of SrcFileList removed from file drop messages between clients,
	P2PMsgMagicLength = 400

	//This is the length of FilePath and FileName removed from FileInfo struct
	FileInfoMagicLength = 80

	//This is the length of content removed from String Array
	StringArrayMagicLength = 5
)
