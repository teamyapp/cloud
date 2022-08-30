package obs

const LogTypeProp string = "LOG_TYPE"
const CauseProp string = "CAUSE"
const MessageProp string = "MESSAGE"

type Props = map[string]interface{}

func MergeProps(propsA Props, propsB Props) Props {
	props := Props{}
	for key, value := range propsA {
		props[key] = value
	}

	for key, value := range propsB {
		props[key] = value
	}

	return props
}
