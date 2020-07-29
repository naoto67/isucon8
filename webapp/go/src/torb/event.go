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

func getEvent(eventID, loginUserID int64) (*Event, error) {
	var event Event
	if err := db.QueryRow("SELECT * FROM events WHERE id = ?", eventID).Scan(&event.ID, &event.Title, &event.PublicFg, &event.ClosedFg, &event.Price); err != nil {
		return nil, err
	}
	event.Remains = 1000
	event.Total = 1000

	rows, err := db.Query("SELECT * FROM reservations WHERE event_id = ? AND canceled_at IS NULL", eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	// 予約の作成
	reservations := make(map[int64]Reservation)
	for rows.Next() {
		var r Reservation
		if err := rows.Scan(&r.ID, &r.EventID, &r.SheetID, &r.UserID, &r.ReservedAt, &r.CanceledAt); err != nil {
			return nil, err
		}

		reservations[r.SheetID] = r
	}
	event.Sheets = makeEventSheets(event.Price)

	sheets := getSheets()
	for _, sheet := range sheets {
		r, ok := reservations[sheet.ID]
		if ok {
			event.Sheets[sheet.Rank].Remains--
			event.Remains--

			sheet.Mine = r.UserID == loginUserID
			sheet.Reserved = true
			sheet.ReservedAtUnix = r.ReservedAt.Unix()
		}
		event.Sheets[sheet.Rank].Detail = append(event.Sheets[sheet.Rank].Detail, sheet)
	}

	return &event, nil
}

func getEvents(all bool) ([]*Event, error) {
	cli, err := FetchMongoDBClient()
	if err != nil {
		return nil, err
	}
	defer cli.Close()
	e, err := cli.FindAllEvents()
	if err != nil {
		return nil, err
	}
	var events []*Event
	for _, v := range e {
		if !all && !v.PublicFg {
			continue
		}
		v, err := getEventWithoutDetail(v)
		if err != nil {
			return nil, err
		}
		events = append(events, v)
	}
	return events, nil
}

func getEventWithoutDetail(e *Event) (*Event, error) {
	res, err := fetchEventReservationCount(e.ID, e.Price)
	if err != nil {
		return nil, err
	}
	e.Total = 1000
	for _, v := range res {
		e.Remains = e.Remains + v.Remains
	}
	e.Sheets = res
	return e, nil
}

func FetchEventDict() (map[int64]*Event, error) {
	cli, err := FetchMongoDBClient()
	if err != nil {
		return nil, err
	}
	defer cli.Close()
	events, err := cli.FindAllEvents()
	if err != nil {
		return nil, err
	}
	dict := make(map[int64]*Event)

	for _, v := range events {
		dict[v.ID] = v
	}
	return dict, nil
}
