package clipboard

import (
	"context"
	"log"
	rtkCommon "rtk-cross-share/common"
	rtkGlobal "rtk-cross-share/global"
	rtkPlatform "rtk-cross-share/platform"
	rtkUtils "rtk-cross-share/utils"
	"sync/atomic"
	"time"
)

var kDefClipboardData = rtkCommon.ClipBoardData{
	SourceID:	"",
	Hash:		"",
	TimeStamp:	0,
	FmtType:	"",
	ExtData:	nil,
}

var lastClipboardData atomic.Value

var dstPasteIpAddr string
var pasteEventIp = make(chan string, 10)

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

func InitClipboard() {
	dstPasteIpAddr = ""

	rtkPlatform.SetCopyImageCallback(func(filesize rtkCommon.FileSize, imageHeader rtkCommon.ImgHeader, data []byte) {
		log.Printf("[%s %d] WatchClipboardCopyImg and UpdateImageClipboardData", rtkUtils.GetFuncName(), rtkUtils.GetLine())
		updateImageClipboardData(rtkGlobal.NodeInfo.ID, filesize, imageHeader, data)

		

	})

	rtkPlatform.SetPasteImageCallback(func() {
		for i := 0; i < rtkUtils.GetClientCount(); i++ {
			pasteEventIp <- dstPasteIpAddr
		}
	})
}

func WatchClipboardText(ctx context.Context, ipAddr string, resultChan chan<- rtkCommon.ClipBoardData) {
	contentText := make(chan string)
	rtkUtils.GoSafe(func() {rtkPlatform.WatchClipboardText(ctx, contentText)})
	var lastHash string
	var lastTimeStamp uint64

	for {
		select {
		case <-ctx.Done():
			return
		case text := <-contentText:
			updateTextClipboardData(rtkGlobal.NodeInfo.ID, text)
		case <-time.After(100 * time.Millisecond):
			lastData := GetLastClipboardData()
			if lastData.SourceID == rtkGlobal.NodeInfo.ID && lastData.FmtType == rtkCommon.TEXT_CB {
				currentHash := lastData.Hash
				currentTimeStamp := lastData.TimeStamp

				if !rtkUtils.ContentEqual([]byte(lastHash), []byte(currentHash)) && lastTimeStamp != currentTimeStamp {
					if extData, ok := lastData.ExtData.(rtkCommon.ExtDataText); ok {
						if extData.Text == "" {
							continue
						}
						log.Printf("[WatchClipboardText][%s] - got new message: %s", ipAddr, string(extData.Text))
						lastHash = currentHash
						lastTimeStamp = currentTimeStamp
						resultChan <- lastData
					} else {
						log.Printf("[%s %d] Err: Invalid text extData", rtkUtils.GetFuncName(), rtkUtils.GetLine())
					}
				}

			}
		}
	}
}

func WatchClipboardImg(ctx context.Context, ipAddr string, resultChan chan<- rtkCommon.ClipBoardData) {
	var lastHash string
	var lastTimeStamp uint64

	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(100 * time.Millisecond):
			lastData := GetLastClipboardData()
			if lastData.SourceID == rtkGlobal.NodeInfo.ID && lastData.FmtType == rtkCommon.IMAGE_CB {
				currentHash := lastData.Hash
				currentTimeStamp := lastData.TimeStamp
				if !rtkUtils.ContentEqual([]byte(lastHash), []byte(currentHash)) && lastTimeStamp != currentTimeStamp {
					if extData, ok := lastData.ExtData.(rtkCommon.ExtDataImg); ok {
						log.Printf("[WatchClipboardImg][%s] - got new Image  Wight:%d Height:%d, content len:[%d] \n\n",
							ipAddr,
							extData.Header.Width,
							extData.Header.Height,
							len(extData.Data))
						lastHash = currentHash
						lastTimeStamp = currentTimeStamp
						resultChan <- lastData
					} else {
						log.Printf("[%s %d] Err: Invalid text extData", rtkUtils.GetFuncName(), rtkUtils.GetLine())
					}
				}

			}
		}
	}
}

func WatchClipboardPasteImg(ctx context.Context, ipAddr string, id string, resultChan chan<- bool) {

	for {
		select {
		case <-ctx.Done():
			return
		case pasteFromIP := <-pasteEventIp:
			if pasteFromIP == ipAddr {
				lastData := GetLastClipboardData()
				if lastData.SourceID == id {
					if _, ok := lastData.ExtData.(rtkCommon.ExtDataImg); ok {
						log.Printf("[WatchClipboardImgPaste] get paste event from [%s]", ipAddr)
						resultChan <- true
					} else {
						log.Printf("[%s %d] Err: Invalid text extData", rtkUtils.GetFuncName(), rtkUtils.GetLine())
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

func SetupDstPasteImage(id string, ipAddr string, content []byte, imgHeader rtkCommon.ImgHeader, dataSize uint32) {
	filesize := rtkCommon.FileSize{
		SizeHigh: 0,
		SizeLow:  dataSize,
	}

	dstPasteIpAddr = ipAddr
	log.Printf("[%s %d] SetupDstPasteImage and UpdateImage, id=%s", rtkUtils.GetFuncName(), rtkUtils.GetLine(), id)
	updateImageClipboardData(id, filesize, imgHeader, content)
	rtkPlatform.GoSetupDstPasteImage(ipAddr, content, imgHeader, dataSize)
}

func SetupDstPasteFile(desc, fileName, platform string, fileSizeHigh uint32, fileSizeLow uint32) {
	rtkPlatform.GoSetupDstPasteFile(desc, fileName, platform, fileSizeHigh, fileSizeLow)
}
