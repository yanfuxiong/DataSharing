package filedrop

import (
	"log"
	rtkCommon "rtk-cross-share/client/common"
	rtkPlatform "rtk-cross-share/client/platform"
	rtkMisc "rtk-cross-share/misc"
)

func init() {
	rtkPlatform.SetGetFilesCacheSendCountCallback(GetFilesTransferDataSendCacheCount)
	rtkPlatform.SetGoCancelFileTransCallback(CancelFileTransfer)
}

func CancelFileTransfer(id, ipAddr string, timestamp uint64) {
	fileDropDataMutex.Lock()
	defer fileDropDataMutex.Unlock()

	if cacheData, ok := filesDataCacheMap[id]; ok {
		if len(cacheData.filesTransferDataQueue) == 0 {
			log.Printf("[%s] ID:[%s] Not fount cache map data\n\n", rtkMisc.GetFuncInfo(), id)
			return
		}

		for i, fileDataItem := range cacheData.filesTransferDataQueue {
			if fileDataItem.TimeStamp == timestamp {
				if fileDataItem.isInProgress {
					if fileDataItem.cancelFn != nil {
						if fileDataItem.FileTransDirection == FilesTransfer_As_Src {
							fileDataItem.cancelFn(rtkCommon.FileTransSrcGuiCancel)
						} else {
							fileDataItem.cancelFn(rtkCommon.FileTransDstGuiCancel)
						}
						fileDataItem.cancelFn = nil
						cacheData.filesTransferDataQueue[i] = fileDataItem
						filesDataCacheMap[id] = cacheData
						log.Printf("[%s] ID:[%s],IP:[%s] timestamp:[%d] CancelFileTransfer in progress success by platform GUI!", rtkMisc.GetFuncInfo(), id, ipAddr, timestamp)
						return
					} else {
						log.Printf("[%s] ID:[%s] timestamp:[%d] Not fount cancelFn from cache map data\n\n", rtkMisc.GetFuncInfo(), id, timestamp)
					}
				} else {
					if queue, asSrc, bOk := RemoveItemFromCacheQueue(cacheData.filesTransferDataQueue, timestamp); bOk {
						cacheData.filesTransferDataQueue = queue
						filesDataCacheMap[id] = cacheData
						log.Printf("[%s] ID:[%s],IP:[%s] timestamp:[%d] CancelFileTransfer Remove cache data success by platform GUI!", rtkMisc.GetFuncInfo(), id, ipAddr, timestamp)
						if callbackSendCancelFileTransferMsgToPeer != nil {
							callbackSendCancelFileTransferMsgToPeer(id, ipAddr, timestamp, asSrc)
						} else {
							log.Println("callbackSendCancelFileTransferMsgToPeer is null!")
						}
					}
				}
			}
		}
	} else {
		log.Printf("[%s] ID:[%s],IP:[%s] Not fount cache map data!\n\n", rtkMisc.GetFuncInfo(), id, ipAddr)
	}
}

func SetFilesDataToCacheAsSrc(id string) uint64 {
	return setFilesDataToCache(id, true)
}

func SetFilesDataToCacheAsDst(id string) uint64 {
	return setFilesDataToCache(id, false)
}

func setFilesDataToCache(id string, isSrc bool) uint64 {
	fileDropDataMutex.Lock()
	defer fileDropDataMutex.Unlock()

	filesDataItem, ok := fileDropDataMap[id]
	if !ok {
		log.Printf("[%s], ID[%s] Not fount file data", rtkMisc.GetFuncInfo(), id)
		return 0
	}

	directType := FilesTransfer_As_Dst
	if isSrc {
		directType = FilesTransfer_As_Src
	}
	cacheData, exist := filesDataCacheMap[id]
	if !exist {
		filesDataCacheMap[id] = filesDataTransferCache{
			filesTransferDataQueue: []FilesTransferDataItem{FilesTransferDataItem{
				FileDropData:       filesDataItem,
				FileTransDirection: directType,
				isInProgress:       false,
			}},
		}
	} else {
		cacheData.filesTransferDataQueue = append(cacheData.filesTransferDataQueue, FilesTransferDataItem{
			FileDropData:       filesDataItem,
			FileTransDirection: directType,
			isInProgress:       false,
		})

		filesDataCacheMap[id] = cacheData
	}

	return filesDataItem.TimeStamp
}

func GetFilesTransferDataItem(id string, timestamp uint64) *FilesTransferDataItem {
	fileDropDataMutex.Lock()
	defer fileDropDataMutex.Unlock()
	if cacheData, ok := filesDataCacheMap[id]; ok {
		nCount := len(cacheData.filesTransferDataQueue)
		if nCount == 0 {
			log.Printf("[%s] ID:[%s] Not fount cache map data\n\n", rtkMisc.GetFuncInfo(), id)
			return nil
		}
		for i, itemCacheValue := range cacheData.filesTransferDataQueue {
			if !itemCacheValue.isInProgress && (itemCacheValue.TimeStamp == timestamp || timestamp == 0) {
				itemCacheValue.isInProgress = true
				cacheData.filesTransferDataQueue[i] = itemCacheValue
				filesDataCacheMap[id] = cacheData

				return &FilesTransferDataItem{
					FileDropData: FileDropData{
						SrcFileList:   itemCacheValue.SrcFileList,
						ActionType:    itemCacheValue.ActionType,
						TimeStamp:     itemCacheValue.TimeStamp,
						FolderList:    itemCacheValue.FolderList,
						TotalDescribe: itemCacheValue.TotalDescribe,
						TotalSize:     itemCacheValue.TotalSize,
						DstFilePath:   itemCacheValue.DstFilePath,
						Cmd:           itemCacheValue.Cmd,
					},
					FileTransDirection: itemCacheValue.FileTransDirection,
				}
			}
		}
	}
	log.Printf("[%s] ID:[%s]  timestamp:[%d] Not fount cache map data\n\n", rtkMisc.GetFuncInfo(), id, timestamp)
	return nil
}

func SetFilesCacheItemComplete(id string, timestamp uint64) {
	fileDropDataMutex.Lock()
	defer fileDropDataMutex.Unlock()
	if cacheData, ok := filesDataCacheMap[id]; ok {
		nItemCount := len(cacheData.filesTransferDataQueue)
		if nItemCount > 0 {
			if cacheData.filesTransferDataQueue, _, ok = RemoveItemFromCacheQueue(cacheData.filesTransferDataQueue, timestamp); !ok {
				log.Printf("[%s] ID:[%s] Not fount cache map data, Item id:[%d]\n\n", rtkMisc.GetFuncInfo(), id, timestamp)
				return
			}
			if nItemCount == 1 && ok {
				log.Printf("[%s] ID:[%s] compelete a files cache item, id:[%d], all files cache data done! \n\n", rtkMisc.GetFuncInfo(), id, timestamp)
			} else {
				log.Printf("[%s] ID:[%s] compelete a files cache item, id:[%d], still %d records left", rtkMisc.GetFuncInfo(), id, timestamp, nItemCount-1)
			}

			filesDataCacheMap[id] = cacheData
		} else {
			log.Printf("[%s] ID:[%s] Not fount cache map data\n\n", rtkMisc.GetFuncInfo(), id)
		}
	} else {
		log.Printf("[%s] ID:[%s] Not fount cache map data\n\n", rtkMisc.GetFuncInfo(), id)
	}
}

func GetFilesTransferDataCacheCount(id string) int {
	fileDropDataMutex.RLock()
	defer fileDropDataMutex.RUnlock()
	if cacheData, ok := filesDataCacheMap[id]; ok {
		return len(cacheData.filesTransferDataQueue)
	} else {
		log.Printf("[%s] ID:[%s] Not fount cache map data\n\n", rtkMisc.GetFuncInfo(), id)
	}
	return 0
}

func GetFilesTransferDataSendCacheCount(id string) int {
	fileDropDataMutex.RLock()
	defer fileDropDataMutex.RUnlock()
	nSendCount := int(0)
	if cacheData, ok := filesDataCacheMap[id]; ok {
		for _, value := range cacheData.filesTransferDataQueue {
			if value.FileTransDirection == FilesTransfer_As_Src {
				nSendCount++
			}
		}
	}
	return nSendCount
}

func SetCancelFileTransferFunc(id string, timeStamp uint64, fn func(rtkCommon.CancelBusinessSource)) {
	fileDropDataMutex.Lock()
	defer fileDropDataMutex.Unlock()

	if cacheData, ok := filesDataCacheMap[id]; ok {
		for i, fileDataItem := range cacheData.filesTransferDataQueue {
			if fileDataItem.TimeStamp == timeStamp {
				fileDataItem.cancelFn = fn
				cacheData.filesTransferDataQueue[i] = fileDataItem
				filesDataCacheMap[id] = cacheData
				return
			}
		}
	} else {
		log.Printf("[%s] ID:[%s] Not fount cache map data\n\n", rtkMisc.GetFuncInfo(), id)
	}
}

func SetFilesTransferDataInterrupt(id, srcFileName, dstFileName, dstFullName string, timestamp uint64, offset int64, errCode rtkMisc.CrossShareErr) bool {
	fileDropDataMutex.Lock()
	defer fileDropDataMutex.Unlock()
	if cacheData, ok := filesDataCacheMap[id]; ok {
		for i, fileDataItem := range cacheData.filesTransferDataQueue {
			if fileDataItem.TimeStamp == timestamp {
				cacheData.filesTransferDataQueue[i].InterruptSrcFileName = srcFileName
				cacheData.filesTransferDataQueue[i].InterruptDstFileName = dstFileName
				cacheData.filesTransferDataQueue[i].InterruptDstFullPath = dstFullName
				cacheData.filesTransferDataQueue[i].InterruptFileOffSet = offset
				cacheData.filesTransferDataQueue[i].InterruptLastErrCode = errCode
				filesDataCacheMap[id] = cacheData
				log.Printf("[%s] ID:[%s] timestamp:[%d] Set interrupt info srcfileName:[%s] offset:[%d] success!", rtkMisc.GetFuncInfo(), id, timestamp, srcFileName, offset)
				return true
			}
		}
	}
	return false
}

func CancelNotStartFileTransFromCacheMap(id string, timestamp uint64) bool {
	fileDropDataMutex.Lock()
	defer fileDropDataMutex.Unlock()
	if cacheData, ok := filesDataCacheMap[id]; ok {
		for _, fileDataItem := range cacheData.filesTransferDataQueue {
			if !fileDataItem.isInProgress && fileDataItem.TimeStamp == timestamp {
				if cacheData.filesTransferDataQueue, _, ok = RemoveItemFromCacheQueue(cacheData.filesTransferDataQueue, timestamp); ok {
					filesDataCacheMap[id] = cacheData
					log.Printf("[%s] ID:[%s] timestamp:[%d] CancelFileTransfer success from cache map data!", rtkMisc.GetFuncInfo(), id, timestamp)
					return true
				}
			}
		}
	}
	return false
}

func RemoveItemFromCacheQueue(slice []FilesTransferDataItem, timestamp uint64) ([]FilesTransferDataItem, bool, bool) {
	tmpSlice := make([]FilesTransferDataItem, 0)
	ok := false
	asSrc := false
	for _, val := range slice {
		if val.TimeStamp == timestamp {
			ok = true
			if val.FileTransDirection == FilesTransfer_As_Src {
				asSrc = true
			}
		} else {
			tmpSlice = append(tmpSlice, val)
		}
	}
	return tmpSlice, asSrc, ok
}
