package common

type FileActionType string

const (
	P2PFileActionType_Drop FileActionType = "FileActionType_Drop"
	P2PFileActionType_Drag FileActionType = "FileActionType_Drag"
)

type FileType string

const (
	P2PFile_Type_Single   FileType = "File_Type_Single"
	P2PFile_Type_Multiple FileType = "File_Type_Multiple"
)

type FileSize struct {
	SizeHigh uint32
	SizeLow  uint32
}

type FileInfo struct {
	FileSize_ FileSize
	FilePath  string //full path
	FileName  string //this must start with folder name, eg: folderName/aaa/bbb/ccc.txt
}

type ImgHeader struct {
	Width       int32
	Height      int32
	Planes      uint16
	BitCount    uint16
	Compression uint32
}

type TransFmtType string

const (
	TEXT_CB 	TransFmtType = "TEXT_CB"
	IMAGE_CB 	TransFmtType = "IMAGE_CB"
	FILE_DROP 	TransFmtType = "FILE_DROP"
)

type ExtDataText struct {
	Text string
}

type ExtDataFile struct {
	SrcFileList   []FileInfo
	ActionType    FileActionType
	FileType      FileType
	TimeStamp     uint64
	FolderList    []string // must start with folder name and end with '/'.  eg: folderName/aaa/bbb/
	TotalDescribe string   // eg: 820MB /1.2GB
	TotalSize     uint64
}

type ExtDataImg struct {
	Size   FileSize
	Header ImgHeader
	Data   []byte
}

type ClipBoardData struct {
	SourceID  string
	Hash      string
	TimeStamp uint64
	FmtType   TransFmtType
	ExtData   interface{} // ExtDataText, ExtDataImg, ExtDataFile(future)
}
