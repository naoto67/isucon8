package main

import (
	"errors"
	"strings"
)

func getSheetByID(sheetID int64) *Sheet {
	if sheetID <= 50 && sheetID > 0 {
		return &Sheet{
			ID: sheetID, Rank: "S", Num: sheetID,
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
		sheet := getSheetByID(int64(i))
		sheets = append(sheets, sheet)
	}
	return sheets
}

func getSheetsByRank(rank string) *Sheets {
	switch rank {
	case "S":
		return &Sheets{
			Total:   50,
			Remains: 50,
			Price:   5000,
		}
	case "A":
		return &Sheets{
			Total:   150,
			Remains: 150,
			Price:   3000,
		}
	case "B":
		return &Sheets{
			Total:   300,
			Remains: 300,
			Price:   1000,
		}
	case "C":
		return &Sheets{
			Total:   500,
			Remains: 500,
			Price:   0,
		}
	}
	return nil
}

func getSheetByRankAndNum(rank string, num int64) (*Sheet, error) {
	if rank == "S" && num <= 50 && num > 0 {
		return &Sheet{
			ID:    num,
			Num:   num,
			Price: 5000,
		}, nil
	} else if rank == "A" && num <= 150 && num > 0 {
		return &Sheet{
			ID:    num + 50,
			Num:   num,
			Price: 3000,
		}, nil

	} else if rank == "B" && num <= 300 && num > 0 {
		return &Sheet{
			ID:    num + 200,
			Num:   num,
			Price: 1000,
		}, nil
	} else if rank == "C" && num <= 500 && num > 0 {
		return &Sheet{
			ID:    num + 500,
			Num:   num,
			Price: 0,
		}, nil
	}
	return nil, errors.New("invalid sheet")
}

func makeEventSheets(eventPrice int64) map[string]*Sheets {
	return map[string]*Sheets{
		"S": &Sheets{Total: 50, Remains: 50, Price: 5000 + eventPrice},
		"A": &Sheets{Total: 150, Remains: 150, Price: 3000 + eventPrice},
		"B": &Sheets{Total: 300, Remains: 300, Price: 1000 + eventPrice},
		"C": &Sheets{Total: 500, Remains: 500, Price: 0 + eventPrice},
	}
}

func validateRank(rank string) bool {
	switch rank {
	case "A", "S", "B", "C":
		return true
	case "a", "s", "b", "c":
		return true
	}
	return false
}

func getNotReservedSheets(eventID int64, rank string) (sheets []Sheet, err error) {
	rows, err := db.Query("SELECT s.id FROM sheets s WHERE exists (SELECT s.id FROM reservations r WHERE event_id = ? AND canceled_at IS NULL AND r.sheet_id = s.id) AND rank = ?", eventID, rank)
	if err != nil {
		return
	}
	defer rows.Close()
	res := map[int64]bool{}
	for rows.Next() {
		var id int64
		err = rows.Scan(&id)
		if err != nil {
			return nil, err
		}

		res[id] = true
	}
	count := getSheetCountByRank(rank)

	for i := 1; i <= count; i++ {
		if _, ok := res[int64(i)]; ok {
			continue
		}

		sheet := getSheetByID(int64(i))
		sheets = append(sheets, *sheet)
	}
	return
}

func validateRankSheetNum(rank string, num int64) bool {
	res := false
	Rank := strings.ToUpper(rank)
	switch Rank {
	case "S":
		if 0 < num && num < 51 {
			res = true
		}
	case "A":
		if 0 < num && num < 151 {
			res = true
		}
	case "B":
		if 0 < num && num < 301 {
			res = true
		}
	case "C":
		if 0 < num && num < 501 {
			res = true
		}
	}
	return res
}

func getSheetCountByRank(rank string) int {
	switch rank {
	case "S":
		return 50
	case "A":
		return 150
	case "B":
		return 300
	case "C":
		return 500
	}
	return 0
}
