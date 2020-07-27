package main

func fetchEventReservationCount(eventID, eventPrice int64) (map[string]*Sheets, error) {
	res := makeEventSheets(eventPrice)
	rows, err := db.Query("SELECT sheets.id, rank, price, COUNT(*) as cnt FROM reservations INNER JOIN sheets ON sheets.id = reservations.sheet_id WHERE canceled_at IS NULL AND event_id = ? GROUP BY sheets.rank", eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var rankCount struct {
			SheetID int64  `db:"id"`
			Rank    string `db:"rank"`
			Count   int    `db:"cnt"`
			Price   int    `db:"price"`
		}
		if err := rows.Scan(&rankCount.SheetID, &rankCount.Rank, &rankCount.Price, &rankCount.Count); err != nil {
			return nil, err
		}
		sheets := getSheetsByRank(rankCount.Rank)
		sheets.Price = sheets.Price + eventPrice
		sheets.Remains = sheets.Remains - rankCount.Count
		res[rankCount.Rank] = sheets
	}

	return res, nil
}
