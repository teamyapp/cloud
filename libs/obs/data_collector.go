package obs

type DataCollector struct {
	Logger Logger
}

func NewDataCollector(logger Logger) DataCollector {
	return DataCollector{Logger: logger}
}
