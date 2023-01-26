package obs

const (
	CauseProp      string = "Cause"
	MessageProp    string = "Message"
	HappenAtProp   string = "HappenAt"
	SeverityProp   string = "Severity"
	FileNameProp   string = "FileName"
	LineNumberProp string = "LineNumber"
)

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
