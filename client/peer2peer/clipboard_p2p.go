package peer2peer

import (
	"bufio"
	"bytes"
	"errors"
	"github.com/libp2p/go-yamux/v5"
	"io"
	"log"
	"net"
	rtkClipboard "rtk-cross-share/client/clipboard"
	rtkCommon "rtk-cross-share/client/common"
	rtkConnection "rtk-cross-share/client/connection"
	rtkUtils "rtk-cross-share/client/utils"
	rtkMisc "rtk-cross-share/misc"
	"time"
)

func writeXClipDataToSocket(id, ipAddr string) rtkMisc.CrossShareErr {
	startTime := time.Now().UnixMilli()
	rtkConnection.HandleFmtTypeStreamReady(id, rtkCommon.XCLIP_CB) // wait for fmtType stream Ready
	sXClip, ok := rtkConnection.GetFmtTypeStream(id, rtkCommon.XCLIP_CB)
	if !ok {
		log.Printf("[%s] Err: Not found  stream by ID:[%s]", rtkMisc.GetFuncInfo(), id)
		return rtkMisc.ERR_BIZ_CB_GET_STREAM_EMPTY
	}
	defer rtkConnection.CloseFmtTypeStream(id, rtkCommon.XCLIP_CB)

	cbData := rtkClipboard.GetLastClipboardData()
	if cbData.FmtType != rtkCommon.XCLIP_CB {
		log.Printf("[%s] GetLastClipboardData Unknown ext data, fmtType: %s", rtkMisc.GetFuncInfo(), cbData.FmtType)
		return rtkMisc.ERR_BIZ_CB_GET_DATA_TYPE_ERR
	}

	extData, ok := cbData.ExtData.(rtkCommon.ExtDataXClip)
	if !ok {
		log.Printf("[%s] GetLastClipboardData Unknown ext data, fmtType: %s", rtkMisc.GetFuncInfo(), cbData.FmtType)
		return rtkMisc.ERR_BIZ_CB_INVALID_DATA
	}

	log.Printf("(SRC) IP[%s] Start to copy XClip data, text:[%d] image:[%d] html:[%d] rtf:[%d]...", ipAddr, extData.TextLen, extData.ImageLen, extData.HtmlLen, extData.RtfLen)
	var readers []io.Reader
	if !rtkUtils.GetPeerClientIsSupportXClip(id) {
		readers = make([]io.Reader, 1)
		readers[0] = bytes.NewReader(extData.Image)
	} else {
		readers = make([]io.Reader, 4)
		readers[0] = bytes.NewReader(extData.Text)
		readers[1] = bytes.NewReader(extData.Image)
		readers[2] = bytes.NewReader(extData.Html)
		readers[3] = bytes.NewReader(extData.Rtf)
	}
	nWrite, err := io.Copy(sXClip, io.MultiReader(readers...))
	if err != nil {
		log.Printf("(SRC) IP:[%s] Copy XClip data err:%+v", ipAddr, err)
		if errors.Is(err, yamux.ErrStreamReset) {
			rtkConnection.LastXClipStreamReset(id) //cannot be closed here when old XClip stream reset by DST, Otherwise, there is a possibility of errors in new XClip stream
		}
		return rtkMisc.ERR_BIZ_CB_SRC_COPY
	}

	log.Printf("(SRC) IP[%s] End to copy XClip data, Total size:[%d] use [%d] ms", ipAddr, nWrite, time.Now().UnixMilli()-startTime)
	bufio.NewWriter(sXClip).Flush()
	return rtkMisc.SUCCESS
}

func handleXClipDataFromSocket(id, ipAddr string) rtkMisc.CrossShareErr {
	startTime := time.Now().UnixMilli()

	cbData := rtkClipboard.GetLastClipboardData()
	if cbData.FmtType != rtkCommon.XCLIP_CB {
		log.Printf("[%s] GetLastClipboardData Unknown ext data, fmtType: %s", rtkMisc.GetFuncInfo(), cbData.FmtType)
		return rtkMisc.ERR_BIZ_CB_GET_DATA_TYPE_ERR
	}

	extData, ok := cbData.ExtData.(rtkCommon.ExtDataXClip)
	if !ok {
		log.Printf("[%s] Invalid XClip ext data type", rtkMisc.GetFuncInfo())
		return rtkMisc.ERR_BIZ_CB_INVALID_DATA
	}

	sXClip, ok := rtkConnection.GetFmtTypeStream(id, rtkCommon.XCLIP_CB)
	if !ok {
		log.Printf("[%s] Err: Not found Image stream by ID: %s", rtkMisc.GetFuncInfo(), id)
		return rtkMisc.ERR_BIZ_CB_GET_STREAM_EMPTY
	}
	defer rtkConnection.CloseFmtTypeStream(id, rtkCommon.XCLIP_CB)

	var xClipBuffer bytes.Buffer
	xClipBuffer.Reset()
	nXClipLen := extData.TextLen + extData.ImageLen + extData.HtmlLen + extData.RtfLen
	xClipBuffer.Grow(int(nXClipLen))

	log.Printf("(DST) IP[%s] Start to Copy XClip data, Total size:[%d]...", ipAddr, nXClipLen)
	nDstWrite, err := io.Copy(&xClipBuffer, sXClip)
	if err != nil {
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			log.Printf("[%s] IP:[%s] (DST) Read XClip data timeout:%+v", rtkMisc.GetFuncInfo(), ipAddr, netErr)
			return rtkMisc.ERR_BIZ_CB_DST_COPY_TIMEOUT
		} else {
			log.Printf("[%s] IP:[%s] (DST) Copy XClip data, Error:%+v", rtkMisc.GetFuncInfo(), ipAddr, err)
			return rtkMisc.ERR_BIZ_CB_DST_COPY
		}
	}

	if nDstWrite >= nXClipLen {
		textData := []byte(nil)
		if extData.TextLen > 0 {
			textData = make([]byte, extData.TextLen)
			io.ReadFull(bytes.NewReader(xClipBuffer.Bytes()[:extData.TextLen]), textData)
		}

		imageData := []byte(nil)
		if extData.ImageLen > 0 {
			imageData = make([]byte, extData.ImageLen)
			io.ReadFull(bytes.NewReader(xClipBuffer.Bytes()[extData.TextLen:(extData.TextLen+extData.ImageLen)]), imageData)
		}

		htmlData := []byte(nil)
		if extData.HtmlLen > 0 {
			htmlData = make([]byte, extData.HtmlLen)
			io.ReadFull(bytes.NewReader(xClipBuffer.Bytes()[(extData.TextLen+extData.ImageLen):(extData.TextLen+extData.ImageLen+extData.HtmlLen)]), htmlData)
		}

		rtfData := []byte(nil)
		if extData.RtfLen > 0 {
			rtfData = make([]byte, extData.RtfLen)
			io.ReadFull(bytes.NewReader(xClipBuffer.Bytes()[(extData.TextLen+extData.ImageLen+extData.HtmlLen):]), rtfData)
		}

		log.Printf("(DST) IP[%s] End to Copy XClip data success, Total size:[%d] use [%d] ms", ipAddr, nDstWrite, time.Now().UnixMilli()-startTime)
		rtkMisc.GoSafe(func() { rtkClipboard.SetupDstPasteXClipData(id, textData, imageData, htmlData, rtfData) })
		xClipBuffer.Reset()
		return rtkMisc.SUCCESS
	} else {
		log.Printf("(DST) IP[%s] End to Copy XClip data and failed, total:[%d], it less then total size:[%d] ...", ipAddr, nDstWrite, nXClipLen)
		return rtkMisc.ERR_BIZ_CB_DST_COPY_LOSS
	}
}
