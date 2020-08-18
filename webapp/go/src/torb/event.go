package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"sort"
)

var (
	EVENT_COUNT_KEY = "ec"
	EVENT_ID_KEY    = "ei:"
)

func (e Event) toJson() []byte {
	m := map[string]interface{}{}
	m["id"] = e.ID
	m["title"] = e.Title
	m["public"] = e.PublicFg
	m["closed"] = e.ClosedFg
	m["price"] = e.Price
	data, _ := json.Marshal(m)
	return data
}

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
	event, err := FetchEventCache(eventID)
	if err != nil {
		return nil, sql.ErrNoRows
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

	return event, nil
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
	eventDict := make(map[int64]*Event)
	chErr := make(chan error)
	go func() {
		eventsCache, err := FetchEventsCache()
		if err != nil {
			chErr <- err
			return
		}

		for _, event := range eventsCache {
			if !all && !event.PublicFg {
				fmt.Println("getEvents: PublicFg", event.PublicFg)
				fmt.Println("getEvents: event", event)
				continue
			}
			fmt.Println("getEvents: event", event)
			// 残りを最大にしてEventを作成
			event.Total = 1000
			event.Remains = 1000
			event.Sheets = makeEventSheets(event.Price)
			eventDict[event.ID] = event
			events = append(events, event)
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

	sort.Slice(events, func(i, j int) bool { return events[i].ID < events[j].ID })
	return events, nil
}

func InitEventCache() error {
	var events []Event
	err := db.Select(&events, "SELECT * FROM events")
	if err != nil {
		return err
	}
	dict := map[string][]byte{}
	for _, v := range events {
		key := fmt.Sprintf("%s%d", EVENT_ID_KEY, v.ID)
		dict[key] = v.toJson()
	}

	data, err := json.Marshal(events[len(events)-1].ID)
	if err != nil {
		return err
	}
	err = cacheClient.SingleSet(EVENT_COUNT_KEY, data)
	if err != nil {
		return err
	}
	return cacheClient.MultiSet(dict)
}

func RegisterEventCache(event Event) error {
	key := fmt.Sprintf("%s%d", EVENT_ID_KEY, event.ID)
	_, err := cacheClient.Increment(EVENT_COUNT_KEY, 1)
	if err != nil {
		return err
	}

	return cacheClient.SingleSet(key, event.toJson())
}

func UpdateEventCache(event Event) error {
	key := fmt.Sprintf("%s%d", EVENT_ID_KEY, event.ID)
	return cacheClient.SingleSet(key, event.toJson())
}
func FetchEventCache(eventID int64) (*Event, error) {
	key := fmt.Sprintf("%s%d", EVENT_ID_KEY, eventID)
	data, err := cacheClient.SingleGet(key)
	if err != nil {
		return nil, err
	}
	var e Event
	err = json.Unmarshal(data, &e)
	if err != nil {
		return nil, err
	}
	return &e, nil
}

func FetchEventsCache() ([]*Event, error) {
	data, err := cacheClient.SingleGet(EVENT_COUNT_KEY)
	if err != nil {
		return nil, err
	}
	var cnt int
	err = json.Unmarshal(data, &cnt)
	if err != nil {
		return nil, err
	}
	cnt += 3

	keys := []string{}
	for i := 1; i <= cnt; i++ {
		keys = append(keys, fmt.Sprintf("%s%d", EVENT_ID_KEY, i))
	}
	d, err := cacheClient.MultiGet(keys)
	var res []*Event
	for i, _ := range d {
		var e Event
		if d[i] == nil {
			continue
		}
		err = json.Unmarshal(d[i], &e)
		fmt.Println("FetchEventsCache: event", e)
		if err != nil {
			return nil, err
		}

		res = append(res, &e)
	}
	return res, nil
}

func FetchEventsDict() (map[int64]*Event, error) {
	events, err := FetchEventsCache()
	if err != nil {
		return nil, err
	}
	res := map[int64]*Event{}
	for _, e := range events {
		res[e.ID] = e
	}
	return res, nil
}
