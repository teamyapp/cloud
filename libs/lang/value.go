package lang

import "time"

func ToInternalValue(externalValue any) any {
	switch value := externalValue.(type) {
	case uint:
		return int64(value)
	case *uint:
		if value == nil {
			return nil
		}

		return int64(*value)
	case uint8:
		return int64(value)
	case *uint8:
		if value == nil {
			return nil
		}

		return int64(*value)
	case uint16:
		return int64(value)
	case *uint16:
		if value == nil {
			return nil
		}

		return int64(*value)
	case uint32:
		return int64(value)
	case *uint32:
		if value == nil {
			return nil
		}

		return int64(*value)
	case uint64:
		return int64(value)
	case *uint64:
		if value == nil {
			return nil
		}

		return int64(*value)
	case int:
		return int64(value)
	case *int:
		if value == nil {
			return nil
		}

		return int64(*value)
	case int8:
		return int64(value)
	case *int8:
		if value == nil {
			return nil
		}

		return int64(*value)
	case int16:
		return int64(value)
	case *int16:
		if value == nil {
			return nil
		}

		return int64(*value)
	case int32:
		return int64(value)
	case *int32:
		if value == nil {
			return nil
		}

		return int64(*value)
	case *int64:
		if value == nil {
			return nil
		}

		return *value
	case float32:
		return float64(value)
	case *float32:
		if value == nil {
			return nil
		}

		return float64(*value)
	case *float64:
		if value == nil {
			return nil
		}

		return *value
	case time.Time:
		return value.Unix()
	case *time.Time:
		if value == nil {
			return nil
		}

		return value.Unix()
	case *string:
		if value == nil {
			return nil
		}

		return *value
	case *bool:
		if value == nil {
			return nil
		}

		return *value
	}

	return externalValue
}
