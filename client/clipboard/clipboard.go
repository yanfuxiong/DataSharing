package clipboard

import (
	"bytes"
	"context"
	"log"
	rtkCommon "rtk-cross-share/client/common"
	rtkGlobal "rtk-cross-share/client/global"
	rtkPlatform "rtk-cross-share/client/platform"
	rtkUtils "rtk-cross-share/client/utils"
	rtkMisc "rtk-cross-share/misc"
	"sync/atomic"
	"time"
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

func updateXClipHead(id string, nText, nImage, nHtml, nRtf int64) {
	log.Printf("[%s] ID:[%s] Text:%d Image:%d Html:%d Rtf:%d", rtkMisc.GetFuncInfo(), id, nText, nImage, nHtml, nRtf)

	clipboardData := rtkCommon.ClipBoardData{
		SourceID:  id,
		FmtType:   rtkCommon.XCLIP_CB,
		Hash:      "",
		TimeStamp: uint64(time.Now().Unix()),
		ExtData: rtkCommon.ExtDataXClip{
			Text:     nil,
			Image:    nil,
			Html:     nil,
			Rtf:      nil,
			TextLen:  nText,
			ImageLen: nImage,
			HtmlLen:  nHtml,
			RtfLen:   nRtf,
		},
	}
	updateLastClipboardData(clipboardData)
}

func updateXClipData(id string, cbText, cbImage, cbHtml, cbRtf []byte) {
	log.Printf("[%s] ID:[%s] Text:%d Image:%d Html:%d Rtf:%d", rtkMisc.GetFuncInfo(), id, len(cbText), len(cbImage), len(cbHtml), len(cbRtf))
	hash, _ := rtkUtils.CreateMD5Hash(bytes.Join([][]byte{cbText, cbImage, cbHtml, cbRtf}, nil))
	clipboardData := rtkCommon.ClipBoardData{
		SourceID:  id,
		FmtType:   rtkCommon.XCLIP_CB,
		Hash:      hash.B58String(),
		TimeStamp: uint64(time.Now().Unix()),
		ExtData: rtkCommon.ExtDataXClip{
			Text:     cbText,
			Image:    cbImage,
			Html:     cbHtml,
			Rtf:      cbRtf,
			TextLen:  int64(len(cbText)),
			ImageLen: int64(len(cbImage)),
			HtmlLen:  int64(len(cbHtml)),
			RtfLen:   int64(len(cbRtf)),
		},
	}
	updateLastClipboardData(clipboardData)
}

func init() {
	rtkPlatform.SetCopyXClipCallback(updateClipboardFromPlatform)
}

func updateClipboardFromPlatform(cbText, cbImage, cbHtml, cbRtf []byte) {
	updateXClipData(rtkGlobal.NodeInfo.ID, cbText, cbImage, cbHtml, cbRtf)

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
						log.Printf("[WatchXClipData][%s] - got new data text:%d, image:%d, html:%d, rtf:%d", ipAddr, extData.TextLen, extData.ImageLen, extData.HtmlLen, extData.RtfLen)
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

func SetupDstPasteXClipHead(id string, nText, nImage, nHtml, nRtf int64) {
	updateXClipHead(id, nText, nImage, nHtml, nRtf)
}

func SetupDstPasteXClipData(id string, text, image, html, rtf []byte) {
	updateXClipData(id, text, image, html, rtf)
	rtkPlatform.GoSetupDstPasteXClipData(text, image, html, rtf)
}
