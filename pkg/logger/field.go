package logger

// String creates a Field with a string value
func String(key, value string) Field {
	return Field{Key: key, Value: value}
}

// Int creates a Field with an integer value
func Int(key string, value int) Field {
	return Field{Key: key, Value: value}
}

// Bool creates a Field with a boolean value
func Bool(key string, value bool) Field {
	return Field{Key: key, Value: value}
}

// Any creates a Field with any value type
func Any(key string, value interface{}) Field {
	return Field{Key: key, Value: value}
}
