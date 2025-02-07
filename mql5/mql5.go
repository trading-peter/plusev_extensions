package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

func fetch(from time.Time, to time.Time) ([]map[string]any, error) {
	fromStr := from.Format("2006-01-02T15:04:05")
	toStr := to.Format("2006-01-02T15:04:05")

	queryStr := fmt.Sprintf("date_mode=1&from=%s&to=%s&importance=15&currencies=262143", url.QueryEscape(fromStr), url.QueryEscape(toStr))

	client := &http.Client{}
	var data = strings.NewReader(queryStr)
	req, err := http.NewRequest("POST", "https://www.mql5.com/en/economic-calendar/content", data)
	if err != nil {
		return nil, err
	}

	req.Header.Set("accept", "*/*")
	req.Header.Set("accept-language", "en-US,en;q=0.9,de-DE;q=0.8,de;q=0.7,en-DE;q=0.6")
	req.Header.Set("cache-control", "no-cache")
	req.Header.Set("content-type", "application/x-www-form-urlencoded")
	req.Header.Set("user-agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/132.0.0.0 Safari/537.36")
	req.Header.Set("x-requested-with", "XMLHttpRequest")
	resp, err := client.Do(req)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	bodyText, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	rawEvents := []map[string]any{}
	err = json.Unmarshal(bodyText, &rawEvents)
	if err != nil {
		return nil, err
	}

	events := []map[string]any{}

	for _, e := range rawEvents {
		eventName := getValue[string]("EventName", e)
		if eventName == "" {
			continue
		}

		eventType := getValue[float64]("EventType", e, 1)
		currencyCode := getValue[string]("CurrencyCode", e)
		startDate := time.UnixMilli(int64(e["ReleaseDate"].(float64))).UTC().Format(time.RFC3339)
		endDate := time.Time{}.Format(time.RFC3339)

		if eventType == 1 {
			endDate = startDate
		}

		actualValue := getValue[string]("ActualValue", e, "N/A")
		previousValue := getValue[string]("PreviousValue", e, "N/A")
		forecastValue := getValue[string]("ForecastValue", e, "N/A")

		var notes string
		if anyMatches[string](func(v string) bool { return v != "N/A" }, actualValue, previousValue, forecastValue) {
			notes = fmt.Sprintf("Actual: %s | Forecast: %s | Previous: %s", actualValue, forecastValue, previousValue)
		}

		events = append(events, map[string]any{
			"title":     fmt.Sprintf("%s%s", ifThen(currencyCode != "", currencyCode+": ", ""), eventName),
			"startDate": startDate,
			"endDate":   endDate,
			"notes":     notes,
		})
	}

	return events, nil
}

type mapValue interface {
	float64 | string | bool
}

func ifThen[T any](condition bool, trueValue T, falseValue T) T {
	if condition {
		return trueValue
	}
	return falseValue
}

func getValue[T mapValue](key string, data map[string]any, defaultValue ...T) T {
	value, ok := data[key].(T)
	if !ok {
		if len(defaultValue) > 0 {
			return defaultValue[0]
		}
		var defaultValue T
		return defaultValue
	}
	return value
}

func anyMatches[T comparable](predicate func(T) bool, values ...T) bool {
	for _, v := range values {
		if predicate(v) {
			return true
		}
	}
	return false
}
