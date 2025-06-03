package clipboard

import (
	"context"
	"log"
	rtkCommon "rtk-cross-share/client/common"
	rtkGlobal "rtk-cross-share/client/global"
	rtkPlatform "rtk-cross-share/client/platform"
	rtkUtils "rtk-cross-share/client/utils"
	rtkMisc "rtk-cross-share/misc"
	"sync/atomic"
	"time"

	"golang.design/x/clipboard"
)

var kDefClipboardData = rtkCommon.ClipBoardData{
	SourceID:  "",
	Hash:      "",
	TimeStamp: 0,
	FmtType:   "",
	ExtData:   nil,
}

var lastClipboardData atomic.Value

var dstPasteImgFromId string
var pasteImgEventId = make(chan string, 5)
var copyImgEventChan = make(chan struct{}, 5)

func updateLastClipboardData(cbData rtkCommon.ClipBoardData) {
	lastClipboardData.Store(kDefClipboardData)
	log.Println("updateLastClipboardData fmt = ", cbData.FmtType)
	lastClipboardData.Store(cbData)
}

func GetLastClipboardData() rtkCommon.ClipBoardData {
	val := lastClipboardData.Load()
	if val == nil {
		lastClipboardData.Store(kDefClipboardData)
		return kDefClipboardData
	}
	return val.(rtkCommon.ClipBoardData)
}

func ResetLastClipboardData() {
	log.Println("ResetLastClipboardData")
	lastClipboardData.Store(kDefClipboardData)
	rtkPlatform.GoCleanClipboard()
}

func updateTextClipboardData(id string, text string) {
	log.Println("UpdateTextClipboardData = ", text)
	hash, _ := rtkUtils.CreateMD5Hash([]byte(text))
	clipboardData := rtkCommon.ClipBoardData{
		SourceID:  id,
		FmtType:   rtkCommon.TEXT_CB,
		TimeStamp: uint64(time.Now().Unix()),
		Hash:      hash.B58String(),
		ExtData: rtkCommon.ExtDataText{
			Text: text,
		},
	}
	updateLastClipboardData(clipboardData)
}

func updateImageClipboardData(id string, filesize rtkCommon.FileSize, imageHeader rtkCommon.ImgHeader, data []byte) {
	log.Printf("UpdateImageClipboardData size=%d, id=%s", filesize, id)
	hash, _ := rtkUtils.CreateMD5Hash([]byte(data))
	clipboardData := rtkCommon.ClipBoardData{
		SourceID:  id,
		FmtType:   rtkCommon.IMAGE_CB,
		Hash:      hash.B58String(),
		TimeStamp: uint64(time.Now().Unix()),
		ExtData: rtkCommon.ExtDataImg{
			Size:   filesize,
			Header: imageHeader,
			Data:   data,
		},
	}
	updateLastClipboardData(clipboardData)
}

func init() {
	err := clipboard.Init()
	if err != nil {
		log.Fatal("clipboard init error:%+v", err)
	}

	dstPasteImgFromId = ""

	rtkPlatform.SetCopyImageCallback(func(filesize rtkCommon.FileSize, imageHeader rtkCommon.ImgHeader, data []byte) {
		log.Printf("[%s %d] WatchClipboardImg and UpdateImage", rtkMisc.GetFuncName(), rtkMisc.GetLine())

		// TODO: It should only receive JPG. Remove image decode flow
		log.Printf("(SRC) Start to convert bmp to jpg")
		startTime := time.Now().UnixNano()
		jpgData, err := rtkUtils.BmpToJpg(data, int(imageHeader.Width), int(imageHeader.Height), int(imageHeader.BitCount))
		if err != nil {
			log.Println("[CopyImage] Encode to jpeg failed, err:", err)
			return
		}
		log.Printf("(SRC) Convert bmp to jpg, size:[%d] use [%d] ms...", len(data), (time.Now().UnixNano()-startTime)/1e6)

		filesize.SizeHigh = 0
		filesize.SizeLow = uint32(len(jpgData))
		updateImageClipboardData(rtkGlobal.NodeInfo.ID, filesize, imageHeader, jpgData)

		nCount := rtkUtils.GetClientCount()
		for i := 0; i < nCount; i++ {
			copyImgEventChan <- struct{}{}
		}

	})

	rtkPlatform.SetPasteImageCallback(func() {
		log.Println("WatchClipboardPasteImg: trigger paste image")
		lastData := GetLastClipboardData()
		if imgData, ok := lastData.ExtData.(rtkCommon.ExtDataImg); ok {
			log.Printf("[Paste] H=%d,W=%d,Planes=%d,BitCnt=%d,Compress=%d",
				imgData.Header.Height, imgData.Header.Width, imgData.Header.Planes, imgData.Header.BitCount, imgData.Header.Compression)
			log.Println("[Paste] Current clipboard data src:", lastData.SourceID)

			nCount := rtkUtils.GetClientCount()
			for i := 0; i < nCount; i++ {
				pasteImgEventId <- dstPasteImgFromId
			}
		}
	})

}
func WatchClipboardText(ctx context.Context, id string, resultChan chan<- rtkCommon.ClipBoardData) {
	contentText := make(chan string)
	rtkMisc.GoSafe(func() { rtkPlatform.WatchClipboardText(ctx, contentText) })
	var lastHash string
	var lastTimeStamp uint64

	for {
		select {
		case <-ctx.Done():
			close(resultChan)
			return
		case text := <-contentText:
			updateTextClipboardData(rtkGlobal.NodeInfo.ID, text)
			lastData := GetLastClipboardData()
			if lastData.SourceID == rtkGlobal.NodeInfo.ID && lastData.FmtType == rtkCommon.TEXT_CB {
				currentHash := lastData.Hash
				currentTimeStamp := lastData.TimeStamp

				if !rtkUtils.ContentEqual([]byte(lastHash), []byte(currentHash)) || lastTimeStamp != currentTimeStamp {
					if extData, ok := lastData.ExtData.(rtkCommon.ExtDataText); ok {
						if extData.Text == "" {
							continue
						}

						ipAddr, _ := rtkUtils.GetClientIp(id)
						log.Printf("[WatchClipboardText][%s] - got new message: %s", ipAddr, string(extData.Text))
						lastHash = currentHash
						lastTimeStamp = currentTimeStamp
						resultChan <- lastData
					} else {
						log.Printf("[%s %d] Err: Invalid text extData", rtkMisc.GetFuncName(), rtkMisc.GetLine())
					}
				}

			}
		}
	}
}

func WatchClipboardImg(ctx context.Context, id string, resultChan chan<- rtkCommon.ClipBoardData) {
	var lastHash string
	var lastTimeStamp uint64

	for {
		select {
		case <-ctx.Done():
			close(resultChan)
			return
		case <-copyImgEventChan:
			lastData := GetLastClipboardData()
			if lastData.SourceID == rtkGlobal.NodeInfo.ID && lastData.FmtType == rtkCommon.IMAGE_CB {
				currentHash := lastData.Hash
				currentTimeStamp := lastData.TimeStamp
				if !rtkUtils.ContentEqual([]byte(lastHash), []byte(currentHash)) || lastTimeStamp != currentTimeStamp {
					if extData, ok := lastData.ExtData.(rtkCommon.ExtDataImg); ok {
						ipAddr, _ := rtkUtils.GetClientIp(id)
						log.Printf("[WatchClipboardImg][%s] - got new Image  Wight:%d Height:%d, content len:[%d] \n\n",
							ipAddr,
							extData.Header.Width,
							extData.Header.Height,
							len(extData.Data))
						lastHash = currentHash
						lastTimeStamp = currentTimeStamp
						resultChan <- lastData
					} else {
						log.Printf("[%s %d] Err: Invalid text extData", rtkMisc.GetFuncName(), rtkMisc.GetLine())
					}
				}

			}
		}
	}
}

func WatchClipboardPasteImg(ctx context.Context, id string, resultChan chan<- struct{}) {
	for {
		select {
		case <-ctx.Done():
			close(resultChan)
			return
		case pasteFromId := <-pasteImgEventId:
			if pasteFromId == id {
				lastData := GetLastClipboardData()
				if lastData.SourceID == id {
					if _, ok := lastData.ExtData.(rtkCommon.ExtDataImg); ok {
						ipAddr, _ := rtkUtils.GetClientIp(id)
						log.Printf("[WatchClipboardImgPaste] get paste event from [%s]", ipAddr)
						resultChan <- struct{}{}
					} else {
						log.Printf("[%s %d] Err: Invalid img extData", rtkMisc.GetFuncName(), rtkMisc.GetLine())
					}
				}
			}
		}
	}
}

func SetupDstPasteText(id string, content []byte) {
	updateTextClipboardData(id, string(content))
	rtkPlatform.GoSetupDstPasteText(content)
}

func SetupDstPasteImage(id string, desc string, content []byte, imgHeader rtkCommon.ImgHeader, dataSize uint32) {
	filesize := rtkCommon.FileSize{
		SizeHigh: 0,
		SizeLow:  dataSize,
	}

	dstPasteImgFromId = id
	log.Printf("[%s %d] SetupDstPasteImage and UpdateImage, id=%s", rtkMisc.GetFuncName(), rtkMisc.GetLine(), id)
	log.Println("[Paste] Setup paste image and wait for requirement")
	log.Printf("[Paste] H=%d,W=%d,Planes=%d,BitCnt=%d,Compress=%d, src=%s",
		imgHeader.Height, imgHeader.Width, imgHeader.Planes, imgHeader.BitCount, imgHeader.Compression, id)
	updateImageClipboardData(id, filesize, imgHeader, content)
	rtkPlatform.GoSetupDstPasteImage(desc, content, imgHeader, dataSize)
}

func SetupDstPasteFile(desc, fileName, platform string, fileSizeHigh uint32, fileSizeLow uint32) {
	rtkPlatform.GoSetupDstPasteFile(desc, fileName, platform, fileSizeHigh, fileSizeLow)
}
