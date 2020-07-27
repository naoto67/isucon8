package main

func getEventBySheetID(sheetID int64) *Sheet {
	if sheetID <= 50 && sheetID > 0 {
		return &Sheet{
			ID:    sheetID,
			Rank:  "S",
			Num:   sheetID,
			Price: 5000,
		}
	} else if sheetID <= 200 {

		return &Sheet{
			ID:    sheetID,
			Rank:  "A",
			Num:   sheetID - 50,
			Price: 3000,
		}
	} else if sheetID <= 500 {
		return &Sheet{
			ID:    sheetID,
			Rank:  "B",
			Num:   sheetID - 200,
			Price: 1000,
		}
	} else if sheetID <= 1000 {
		return &Sheet{
			ID:    sheetID,
			Rank:  "C",
			Num:   sheetID - 500,
			Price: 0,
		}
	}
	return nil
}

func getSheets() []*Sheet {
	var sheets []*Sheet
	for i := 1; i <= 1000; i++ {
		sheet := getEventBySheetID(int64(i))
		sheets = append(sheets, sheet)
	}
	return sheets
}

func makeEventSheets(eventPrice int64) map[string]*Sheets {
	return map[string]*Sheets{
		"S": &Sheets{Total: 50, Remains: 50, Price: 5000 + eventPrice},
		"A": &Sheets{Total: 150, Remains: 150, Price: 3000 + eventPrice},
		"B": &Sheets{Total: 300, Remains: 300, Price: 1000 + eventPrice},
		"C": &Sheets{Total: 500, Remains: 500, Price: 0 + eventPrice},
	}
}
