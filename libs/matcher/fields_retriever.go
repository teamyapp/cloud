package matcher

type FieldsRetriever struct {
}

func (f *FieldsRetriever) RetrieveField(object interface{}, attribute string) (value interface{}) {
	return nil
}

func NewFieldsRetriever(customFieldsRetriever map[string]CustomFieldRetriever) *FieldsRetriever {
	return &FieldsRetriever{}
}

type CustomFieldRetriever interface {
	RetrieveField(object interface{}) (value interface{})
}
