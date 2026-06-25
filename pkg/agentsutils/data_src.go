package agentsutils

type TestDataSource struct {
}

func (ds *TestDataSource) GetDataFor(string) []string {
	return []string{"abc", "def"}
}

func (ds *TestDataSource) GetStaticData() int64 {
	return int64(42)
}
