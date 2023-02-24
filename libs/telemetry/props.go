package telemetry

const (
	HappenAtProp    string = "HappenAt"
	ServiceNameProp string = "ServiceName"
	CommitProp      string = "Commit"
	SeverityProp    string = "Severity"
	TraceIDProp     string = "TraceId"
	SpanIDProp      string = "SpanId"
	RequestIDProp   string = "RequestId"
	ClientIDProp    string = "ClientId"
	CauseProp       string = "Cause"
	StackTraceProp  string = "StackTrace"
	MessageProp     string = "Message"
	FileNameProp    string = "FileName"
	LineNumberProp  string = "LineNumber"
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
