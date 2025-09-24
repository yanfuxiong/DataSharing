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
var copyXClipEventChan = make(chan struct{}, 5)

func updateLastClipboardData(cbData rtkCommon.ClipBoardData) {
	lastClipboardData.Store(kDefClipboardData)
	log.Printf("updateLastClipboardData from ID:[%s] fmt =[%s] ", cbData.SourceID, cbData.FmtType)
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

func updateXClipHead(id string, nText, nImage, nHtml int64) {
	log.Printf("[%s] ID:[%s] Text:%d Image:%d Html:%d", rtkMisc.GetFuncInfo(), id, nText, nImage, nHtml)

	clipboardData := rtkCommon.ClipBoardData{
		SourceID:  id,
		FmtType:   rtkCommon.XCLIP_CB,
		Hash:      "",
		TimeStamp: uint64(time.Now().Unix()),
		ExtData: rtkCommon.ExtDataXClip{
			Text:     nil,
			Image:    nil,
			Html:     nil,
			TextLen:  nText,
			ImageLen: nImage,
			HtmlLen:  nHtml,
		},
	}
	updateLastClipboardData(clipboardData)
}

func updateXClipData(id string, cbText, cbImage, cbHtml []byte) {
	log.Printf("[%s] ID:[%s] Text:%d Image:%d Html:%d", rtkMisc.GetFuncInfo(), id, len(cbText), len(cbImage), len(cbHtml))
	hash, _ := rtkUtils.CreateMD5Hash(append(append(cbText, cbImage...), cbHtml...))
	clipboardData := rtkCommon.ClipBoardData{
		SourceID:  id,
		FmtType:   rtkCommon.XCLIP_CB,
		Hash:      hash.B58String(),
		TimeStamp: uint64(time.Now().Unix()),
		ExtData: rtkCommon.ExtDataXClip{
			Text:     cbText,
			Image:    cbImage,
			Html:     cbHtml,
			TextLen:  int64(len(cbText)),
			ImageLen: int64(len(cbImage)),
			HtmlLen:  int64(len(cbHtml)),
		},
	}
	updateLastClipboardData(clipboardData)
}

func init() {
	err := clipboard.Init()
	if err != nil {
		log.Fatal("clipboard init error:%+v", err)
	}

	rtkPlatform.SetCopyXClipCallback(updateClipboardFromPlatform)
}

func updateClipboardFromPlatform(cbText, cbImage, cbHtml []byte) {
	updateXClipData(rtkGlobal.NodeInfo.ID, cbText, cbImage, cbHtml)

	nCount := rtkUtils.GetClientCount()
	for i := 0; i < nCount; i++ {
		copyXClipEventChan <- struct{}{}
	}
}

func WatchXClipData(ctx context.Context, id string, resultChan chan<- rtkCommon.ClipBoardData) {
	var lastHash string
	var lastTimeStamp uint64

	for {
		select {
		case <-ctx.Done():
			close(resultChan)
			return
		case <-copyXClipEventChan:
			lastData := GetLastClipboardData()
			if lastData.SourceID == rtkGlobal.NodeInfo.ID && lastData.FmtType == rtkCommon.XCLIP_CB {
				currentHash := lastData.Hash
				currentTimeStamp := lastData.TimeStamp

				if !rtkUtils.ContentEqual([]byte(lastHash), []byte(currentHash)) || lastTimeStamp != currentTimeStamp {
					if extData, ok := lastData.ExtData.(rtkCommon.ExtDataXClip); ok {
						ipAddr, _ := rtkUtils.GetClientIp(id)
						log.Printf("[WatchXClipData][%s] - got new data text:%d, image:%d, html:%d", ipAddr, extData.TextLen, extData.ImageLen, extData.HtmlLen)
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

func SetupDstPasteFile(desc, fileName, platform string, fileSizeHigh uint32, fileSizeLow uint32) {
	rtkPlatform.GoSetupDstPasteFile(desc, fileName, platform, fileSizeHigh, fileSizeLow)
}

func SetupDstPasteXClipHead(id string, nText, nImage, nHtml int64) {
	updateXClipHead(id, nText, nImage, nHtml)
}

func SetupDstPasteXClipData(id string, text, image, html []byte) {
	updateXClipData(id, text, image, html)
	rtkPlatform.GoSetupDstPasteXClipData(text, image, html)
}
