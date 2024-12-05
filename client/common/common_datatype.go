package common

type FileSize struct {
	SizeHigh uint32
	SizeLow  uint32
}

type FileInfo struct {
	FileSize_ FileSize
	FilePath  string
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
	Size     FileSize
	FilePath string
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
