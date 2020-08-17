package main

import (
	"fmt"
	"time"
)

func fetchEventReservationCount(eventID, eventPrice int64) (map[string]*Sheets, error) {
	res := makeEventSheets(eventPrice)
	rows, err := db.Query("SELECT sheet_id FROM reservations WHERE canceled_at IS NULL AND event_id = ?", eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var sheetID int64
		if err := rows.Scan(&sheetID); err != nil {
			return nil, err
		}
		sheet := getSheetByID(sheetID)
		res[sheet.Rank].Remains = res[sheet.Rank].Remains - 1
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

func getEvents(all bool) ([]*Event, error) {
	var events []*Event
	key := fmt.Sprintf("getEvents:%t", all)
	res, ok := gc.Get(key)
	if ok {
		return res.([]*Event), nil
	}
	eventDict := make(map[int64]*Event)
	chErr := make(chan error)
	go func() {
		rows, err := db.Query("SELECT * FROM events ORDER BY id ASC")
		if err != nil {
			chErr <- err
			return
		}
		defer rows.Close()

		for rows.Next() {
			var event Event
			if err := rows.Scan(&event.ID, &event.Title, &event.PublicFg, &event.ClosedFg, &event.Price); err != nil {
				chErr <- err
				return
			}
			if !all && !event.PublicFg {
				continue
			}
			// 残りを最大にしてEventを作成
			event.Total = 1000
			event.Remains = 1000
			event.Sheets = makeEventSheets(event.Price)
			eventDict[event.ID] = &event
			events = append(events, &event)
		}
		chErr <- nil
	}()

	rows, err := db.Query("SELECT event_id, sheet_id FROM reservations WHERE canceled_at IS NULL")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	if err = <-chErr; err != nil {
		return nil, err
	}
	for rows.Next() {
		var eventSheet struct {
			EventID int64
			SheetID int64
		}
		if err := rows.Scan(&eventSheet.EventID, &eventSheet.SheetID); err != nil {
			return nil, err
		}
		if v, ok := eventDict[eventSheet.EventID]; ok {
			v.Remains = v.Remains - 1
			sheet := getSheetByID(eventSheet.SheetID)
			v.Sheets[sheet.Rank].Remains = v.Sheets[sheet.Rank].Remains - 1
		}
	}
	gc.Set(key, events, 100*time.Millisecond)
	return events, nil
}
