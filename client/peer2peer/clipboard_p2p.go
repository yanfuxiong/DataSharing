package peer2peer

import (
	"bufio"
	"bytes"
	"io"
	"log"
	"net"
	rtkClipboard "rtk-cross-share/client/clipboard"
	rtkCommon "rtk-cross-share/client/common"
	rtkConnection "rtk-cross-share/client/connection"
	rtkPlatform "rtk-cross-share/client/platform"
	rtkMisc "rtk-cross-share/misc"
	"time"
)

func writeImageToSocket(id string) rtkMisc.CrossShareErr {
	startTime := time.Now().UnixMilli()
	rtkConnection.HandleFmtTypeStreamReady(id, rtkCommon.IMAGE_CB) // wait for fmtType stream Ready
	sImage, ok := rtkConnection.GetFmtTypeStream(id, rtkCommon.IMAGE_CB)
	if !ok {
		log.Printf("[%s] Err: Not found  stream by ID:[%s]", rtkMisc.GetFuncInfo(), id)
		return rtkMisc.ERR_BIZ_CB_GET_STREAM_EMPTY
	}
	defer rtkConnection.CloseFmtTypeStream(id, rtkCommon.IMAGE_CB)

	cbData := rtkClipboard.GetLastClipboardData()
	if cbData.FmtType != rtkCommon.IMAGE_CB {
		log.Printf("[%s] GetLastClipboardData Unknown ext data, fmtType: %s", rtkMisc.GetFuncInfo(), cbData.FmtType)
		return rtkMisc.ERR_BIZ_CB_GET_DATA_TYPE_ERR
	}

	extData, ok := cbData.ExtData.(rtkCommon.ExtDataImg)
	if !ok {
		log.Printf("[%s] GetLastClipboardData Unknown ext data, fmtType: %s", rtkMisc.GetFuncInfo(), cbData.FmtType)
		return rtkMisc.ERR_BIZ_CB_INVALID_DATA
	}

	log.Printf("(SRC) Start to copy img, len[%d]...", len(extData.Data))
	// TODO: here must set read deadline to handle receiver exceptions, otherwise it where be block , check how to set deadline
	//sFmtType.SetWriteDeadline(time.Now().Add(30 * time.Second))
	_, err := io.Copy(sImage, bytes.NewReader(extData.Data))
	if err != nil {
		log.Printf("(SRC) Copy imge err:%+v", err)
		return rtkMisc.ERR_BIZ_CB_SRC_COPY_IMAGE
	}
	log.Printf("(SRC) End to copy img, use [%d] ms ...", time.Now().UnixMilli()-startTime)
	bufio.NewWriter(sImage).Flush()

	return rtkMisc.SUCCESS
}

func handleCopyImageFromSocket(id, ipAddr string) rtkMisc.CrossShareErr {
	startTime := time.Now().UnixMilli()

	cbData := rtkClipboard.GetLastClipboardData()
	if cbData.FmtType != rtkCommon.IMAGE_CB {
		log.Printf("[%s] GetLastClipboardData Unknown ext data, fmtType: %s", rtkMisc.GetFuncInfo(), cbData.FmtType)
		return rtkMisc.ERR_BIZ_CB_GET_DATA_TYPE_ERR
	}

	extData, ok := cbData.ExtData.(rtkCommon.ExtDataImg)
	if !ok {
		log.Printf("[%s] Invalid image ext data type", rtkMisc.GetFuncInfo())
		return rtkMisc.ERR_BIZ_CB_INVALID_DATA
	}
	imageSize := int64(extData.Size.SizeHigh)<<32 | int64(extData.Size.SizeLow)

	sImage, ok := rtkConnection.GetFmtTypeStream(id, rtkCommon.IMAGE_CB)
	if !ok {
		log.Printf("[%s] Err: Not found Image stream by ID: %s", rtkMisc.GetFuncInfo(), id)
		return rtkMisc.ERR_BIZ_CB_GET_STREAM_EMPTY
	}
	defer rtkConnection.CloseFmtTypeStream(id, rtkCommon.IMAGE_CB)

	var imgBuffer bytes.Buffer
	imgBuffer.Grow(int(imageSize))

	log.Printf("(DST) IP[%s] Start to Copy image size:[%d]...", ipAddr, imageSize)
	// TODO: here must set read deadline to handle sender exceptions, otherwise it where be block , check how to set deadline
	//sImage.SetReadDeadline(time.Now().Add(30 * time.Second))
	nDstWrite, err := io.Copy(&imgBuffer, sImage)
	if err != nil {
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			log.Printf("[%s] IP:[%s] (DST) Read image  timeout:%+v", rtkMisc.GetFuncInfo(), ipAddr, netErr)
			return rtkMisc.ERR_BIZ_CB_DST_COPY_IMAGE_TIMEOUT
		} else {
			log.Printf("[%s] IP:[%s] (DST) Copy image Error:%+v", rtkMisc.GetFuncInfo(), ipAddr, err)
			return rtkMisc.ERR_BIZ_CB_DST_COPY_IMAGE
		}
	}

	if nDstWrite >= imageSize {
		rtkPlatform.GoDataTransfer(imgBuffer.Bytes())
		rtkPlatform.ReceiveImageCopyDataDone(imageSize, extData.Header) // Only For Android
		imgBuffer.Reset()
		log.Printf("(DST) IP[%s] End to Copy image success, total:[%d] use [%d] ms...", ipAddr, nDstWrite, time.Now().UnixMilli()-startTime)
		return rtkMisc.SUCCESS
	} else {
		log.Printf("(DST) IP[%s] End to Copy image and failed, total:[%d], it less then image size:[%d] ...", ipAddr, nDstWrite, imageSize)
		return rtkMisc.ERR_BIZ_CB_DST_COPY_IMAGE_LOSS
	}
}
