package filedrop

import (
	"log"
	rtkPlatform "rtk-cross-share/client/platform"
	rtkMisc "rtk-cross-share/misc"
)

func init() {
	rtkPlatform.SetGetFilesCacheSendCountCallback(GetFilesTransferDataSendCacheCount)
	rtkPlatform.SetGoCancelFileTransCallback(CancelFileTransfer)
}

func CancelFileTransfer(id, ip string, timestamp uint64) {
	fileDropDataMutex.Lock()
	defer fileDropDataMutex.Unlock()

	if cacheData, ok := filesDataCacheMap[id]; ok {
		if len(cacheData.filesTransferDataQueue) == 0 {
			log.Printf("[%s] ID:[%s] Not fount cache map data\n\n", rtkMisc.GetFuncInfo(), id)
			return
		}

		if cacheData.filesTransferDataQueue[0].TimeStamp == timestamp {
			if cacheData.cancelFn != nil {
				cacheData.cancelFn()
				cacheData.cancelFn = nil
				cacheData.isCancelByGui = true
				filesDataCacheMap[id] = cacheData
				log.Printf("[%s] ID:[%s],IPAddr:[%s] id:[%s] CancelFileTransfer success by platform GUI!", id, ip, timestamp)
			} else {
				log.Printf("[%s] ID:[%s] Not fount cancelFn from cache map data\n\n", rtkMisc.GetFuncInfo(), id)
			}
		} else {
			if queue, bOk := RemoveItemFromCacheQueue(cacheData.filesTransferDataQueue, timestamp); bOk {
				cacheData.filesTransferDataQueue = queue
				filesDataCacheMap[id] = cacheData

				log.Printf("[%s] ID:[%s],IPAddr:[%s] id:[%s] CancelFileTransfer Remove cache data success by platform GUI!", id, ip, timestamp)
				//rtkConnection.CloseFileDropItemStream(id, timestamp)
				if callbackSendCancelFileTransferMsgToPeer != nil {
					callbackSendCancelFileTransferMsgToPeer(id, timestamp)
				} else {
					log.Println("callbackSendCancelFileTransferMsgToPeer is null!")
				}
			} else {
				log.Printf("[%s] ID:[%s],IPAddr:[%s], timestamp:[%d] invalid! ", rtkMisc.GetFuncInfo(), id, ip, timestamp)
			}
		}
	} else {
		log.Printf("[%s] ID:[%s],IPAddr:[%s] Not fount cache map data!\n\n", rtkMisc.GetFuncInfo(), id, ip)
	}
}

func SetFilesDataToCacheAsSrc(id string) {
	setFilesDataToCache(id, true)
}

func SetFilesDataToCacheAsDst(id string) {
	setFilesDataToCache(id, false)
}

func setFilesDataToCache(id string, isSrc bool) {
	fileDropDataMutex.Lock()
	defer fileDropDataMutex.Unlock()

	filesDataItem, ok := fileDropDataMap[id]
	if !ok {
		log.Printf("[%s], ID[%s] Not fount file data", rtkMisc.GetFuncInfo(), id)
		return
	}

	directType := FilesTransferDirectionType(FilesTransfer_As_Dst)
	if isSrc {
		directType = FilesTransfer_As_Src
	}
	cacheData, exist := filesDataCacheMap[id]
	if !exist {
		filesDataCacheMap[id] = filesDataTransferCache{
			filesTransferDataQueue: []FilesTransferDataItem{FilesTransferDataItem{
				FileDropData:       filesDataItem,
				FileTransDirection: directType,
			}},
			cancelFn:      nil,
			isCancelByGui: false,
		}
	} else {
		cacheData.filesTransferDataQueue = append(cacheData.filesTransferDataQueue, FilesTransferDataItem{
			FileDropData:       filesDataItem,
			FileTransDirection: directType,
		})

		filesDataCacheMap[id] = cacheData
	}
}

func GetFilesTransferDataItem(id string) *FilesTransferDataItem {
	fileDropDataMutex.RLock()
	defer fileDropDataMutex.RUnlock()
	if cacheData, ok := filesDataCacheMap[id]; ok {
		nCount := len(cacheData.filesTransferDataQueue)
		if nCount > 0 {
			return &cacheData.filesTransferDataQueue[0]
		} else {
			log.Printf("[%s] ID:[%s] Not fount cache map data\n\n", rtkMisc.GetFuncInfo(), id)
			return nil
		}
	}
	log.Printf("[%s] ID:[%s] Not fount cache map data\n\n", rtkMisc.GetFuncInfo(), id)
	return nil
}

func SetFilesCacheItemComplete(id string, timestamp uint64) {
	fileDropDataMutex.Lock()
	defer fileDropDataMutex.Unlock()
	if cacheData, ok := filesDataCacheMap[id]; ok {
		nItemCount := len(cacheData.filesTransferDataQueue)
		if nItemCount > 0 {
			if cacheData.filesTransferDataQueue, ok = RemoveItemFromCacheQueue(cacheData.filesTransferDataQueue, timestamp); !ok {
				log.Printf("[%s] ID:[%s] Not fount cache map data, Item id:[%d]\n\n", rtkMisc.GetFuncInfo(), id, timestamp)
				return
			}
			if nItemCount == 1 && ok {
				log.Printf("[%s] ID:[%s] compelete a files cache item, id:[%d], all files cache data done! \n\n", rtkMisc.GetFuncInfo(), id, timestamp)
			} else {
				log.Printf("[%s] ID:[%s] compelete a files cache item, id:[%d], still %d records left", rtkMisc.GetFuncInfo(), id, timestamp, nItemCount-1)
			}
			cacheData.cancelFn = nil
			cacheData.isCancelByGui = false
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

func SetCancelFileTransferFunc(id string, fn func()) {
	fileDropDataMutex.Lock()
	defer fileDropDataMutex.Unlock()

	if cacheData, ok := filesDataCacheMap[id]; ok {
		cacheData.cancelFn = fn
		cacheData.isCancelByGui = false
		filesDataCacheMap[id] = cacheData
	} else {
		log.Printf("[%s] ID:[%s] Not fount cache map data\n\n", rtkMisc.GetFuncInfo(), id)
	}
}

func IsCancelFileTransferByGui(id string) bool {
	fileDropDataMutex.RLock()
	defer fileDropDataMutex.RUnlock()
	if cacheData, ok := filesDataCacheMap[id]; ok {
		return cacheData.isCancelByGui
	}
	log.Printf("[%s] ID:[%s] Not fount cache map data\n\n", rtkMisc.GetFuncInfo(), id)
	return false
}

func IsFileTransInProgress(id string, timestamp uint64) bool {
	fileDropDataMutex.RLock()
	defer fileDropDataMutex.RUnlock()
	if cacheData, ok := filesDataCacheMap[id]; ok {
		if len(cacheData.filesTransferDataQueue) > 0 {
			if cacheData.filesTransferDataQueue[0].TimeStamp == timestamp {
				return true
			}
		} else {
			log.Printf("[%s] ID:[%s] Not fount cache map data\n\n", rtkMisc.GetFuncInfo(), id)
		}
	}

	return false
}

func CancelFileTransFromCacheMap(id string, timestamp uint64) {
	fileDropDataMutex.Lock()
	defer fileDropDataMutex.Unlock()
	if cacheData, ok := filesDataCacheMap[id]; ok {
		cacheData.filesTransferDataQueue, ok = RemoveItemFromCacheQueue(cacheData.filesTransferDataQueue, timestamp)
		if ok {
			filesDataCacheMap[id] = cacheData
			//rtkConnection.CloseFileDropItemStream(id, timestamp)
			log.Printf("[%s] ID:[%s] id:[%d] CancelFileTransfer success from cache map data!", rtkMisc.GetFuncInfo(), id, timestamp)
		} else {
			log.Printf("[%s] ID:[%s] Not fount cache map data\n\n", rtkMisc.GetFuncInfo(), id)
		}
	}
}

func RemoveItemFromCacheQueue(slice []FilesTransferDataItem, timestamp uint64) ([]FilesTransferDataItem, bool) {
	tmpSlice := make([]FilesTransferDataItem, 0)
	ok := false
	for _, val := range slice {
		if val.TimeStamp == timestamp {
			ok = true
		} else {
			tmpSlice = append(tmpSlice, val)
		}
	}
	return tmpSlice, ok
}
