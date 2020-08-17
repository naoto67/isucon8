package main

import (
	"fmt"
	"sync"
)

var (
	eventSheetMap = sync.Map{}
)

func LockEventSheet(eventID, sheetID int64) (locked bool) {
	key := fmt.Sprintf("%d:%d", eventID, sheetID)
	_, locked = eventSheetMap.LoadOrStore(key, true)
	return
}

func UnlockEventSheet(eventID, sheetID int64) {
	key := fmt.Sprintf("%d:%d", eventID, sheetID)
	eventSheetMap.Delete(key)
}
