package main

func extractLatestMonthData(months []interface{}) map[string]interface{} {
	var latestMonthData map[string]interface{}
	latestMonthKey := -1

	for _, monthEntry := range months {
		monthData, ok := monthEntry.(map[string]interface{})
		if !ok {
			continue
		}

		dateData, ok := monthData["date"].(map[string]interface{})
		if !ok {
			continue
		}

		year, yearOk := dateData["year"].(float64)
		month, monthOk := dateData["month"].(float64)
		if !yearOk || !monthOk {
			continue
		}

		monthKey := int(year)*100 + int(month)
		if monthKey > latestMonthKey {
			latestMonthKey = monthKey
			latestMonthData = monthData
		}
	}

	if latestMonthData != nil {
		return latestMonthData
	}

	lastMonthData, ok := months[len(months)-1].(map[string]interface{})
	if !ok {
		return nil
	}

	return lastMonthData
}
