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

func updateXClipHead(id string, nText, nImage, nHtml int64, imgHeader rtkCommon.ImgHeader) {
	log.Printf("[%s] ID:[%s] Text:%d Image:%d (%d,%d) Html:%d", rtkMisc.GetFuncInfo(), id, nText, nImage, imgHeader.Width, imgHeader.Height, nHtml)

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
			ImgSize: rtkCommon.FileSize{
				SizeHigh: uint32(0),
				SizeLow:  uint32(nImage),
			},
			ImgHeader: imgHeader,
		},
	}
	updateLastClipboardData(clipboardData)
}

func updateXClipData(id string, cbText, cbImage, cbHtml []byte, imgHeader rtkCommon.ImgHeader) {
	log.Printf("[%s] ID:[%s] Text:%d Image:%d (%d,%d) Html:%d", rtkMisc.GetFuncInfo(), id, len(cbText), len(cbImage), imgHeader.Width, imgHeader.Height, len(cbHtml))
	imgSize := rtkCommon.FileSize{
		SizeHigh: 0,
		SizeLow:  uint32(len(cbImage)),
	}

	hash, _ := rtkUtils.CreateMD5Hash(append(append(cbText, cbImage...), cbHtml...))
	clipboardData := rtkCommon.ClipBoardData{
		SourceID:  id,
		FmtType:   rtkCommon.XCLIP_CB,
		Hash:      hash.B58String(),
		TimeStamp: uint64(time.Now().Unix()),
		ExtData: rtkCommon.ExtDataXClip{
			Text:      cbText,
			Image:     cbImage,
			Html:      cbHtml,
			TextLen:   int64(len(cbText)),
			ImageLen:  int64(len(cbImage)),
			HtmlLen:   int64(len(cbHtml)),
			ImgSize:   imgSize,
			ImgHeader: imgHeader,
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

	rtkPlatform.SetCopyImageCallback(func(imageHeader rtkCommon.ImgHeader, data []byte) {
		log.Printf("[%s %d] WatchClipboardImg and UpdateImage", rtkMisc.GetFuncName(), rtkMisc.GetLine())

		fileSize := rtkCommon.FileSize{
			SizeHigh: 0,
			SizeLow:  uint32(len(data)),
		}
		updateImageClipboardData(rtkGlobal.NodeInfo.ID, fileSize, imageHeader, data)

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

	rtkPlatform.SetCopyXClipCallback(updateClipboardFromPlatform)
}

func updateClipboardFromPlatform(cbText, cbImage, cbHtml []byte, imageHeader rtkCommon.ImgHeader) {
	updateXClipData(rtkGlobal.NodeInfo.ID, cbText, cbImage, cbHtml, imageHeader)

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

func SetupDstPasteXClipHead(id string, nText, nImage, nHtml int64, imgHeader rtkCommon.ImgHeader) {
	updateXClipHead(id, nText, nImage, nHtml, imgHeader)
}

func SetupDstPasteXClipData(id string, text, image, html []byte) {
	updateXClipData(id, text, image, html, rtkCommon.ImgHeader{})
	rtkPlatform.GoSetupDstPasteXClipData(text, image, html)
}
