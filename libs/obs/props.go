package obs

const (
	LogTypeProp    string = "LogType"
	CauseProp             = "Cause"
	MessageProp           = "Message"
	HappenAtProp          = "HappenAt"
	SeverityProp          = "Severity"
	FileNameProp          = "FileName"
	LineNumberProp        = "LineNumber"
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
