package global

const (
	ClientVersion = "2.3.49"

	ClientDefaultVersion        = "2.3.0" // when the other client is an old version and cannot obtain the version number, use this default version
	QueueFileTransVersionSerial = 49      // the client support file drop queue transfer since third version(serial number) 48
	ClientXClipVerSerial        = 46      // the client support XClip since third version(serial number) 46
	ClientCaptureIndexVerSerial = 48      // the client build ClientIndex color block on verification dialog

	ProtocolID     = "/libp2p/dcutr"
	HostProtocolID = "host_register"

	HOST_ID                   = "12345" // This HOST ID is pseudo to test
	ProtocolDirectID          = "/instruction/cross_share/1.0.0"
	ProtocolImageTransmission = "/ipfs/protocol/cross_share/1.0.0"
	ProtocolFileTransmission  = "/ipfs/protocol/cross_share/1.0.1"
	ProtocolFileTransQueue    = "/ipfs/protocol/cross_share/fileDataTransfer/"
	DefaultPort               = 0

	// This is the maximum length of messages between clients,    32KB
	P2PMsgMaxLength = 32 * 1024

	//This is the length of SrcFileList removed from file drop messages between clients,
	P2PMsgMagicLength = 400

	//This is the length of FilePath and FileName removed from FileInfo struct
	FileInfoMagicLength = 80

	//This is the length of content removed from String Array
	StringArrayMagicLength = 5

	//The maximum size for sending documents each time,   10GB
	SendFilesRequestMaxSize = 10 * 1024 * 1024 * 1024

	//The maximum size for cache queue size 5
	SendFilesRequestMaxQueueSize = 5
)
