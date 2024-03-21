package gpbckpexporter

// Convert bool to float64.
func convertBoolToFloat64(value bool) float64 {
	if value {
		return 1
	}
	return 0
}

// Convert backup status to float64.
func convertStatusFloat64(valueStatus string) float64 {
	if valueStatus == "Failure" {
		return 1
	}
	return 0
}

// Convert empty string to empty label ("none" value).
func convertEmptyLabel(str string) string {
	if str == "" {
		return emptyLabel
	}
	return str
}
