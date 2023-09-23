
func TestMockToSql(t *testing.T) {
	m := Mock{Filepath: "sample.csv"}
	mockAsSQl, err := mockToSql(m)
	assert.Nil(t, err)
	expectedSQL1 := "SELECT \"something\" AS column1, \"1.0\" AS column2, \"100\" AS column3"
	expectedSQL2 := "SELECT \"something2\" AS column1, \"2.0\" AS column2, \"200\" AS column3"
	expectedSQL3 := "SELECT \"something3\" AS column1, \"3.0\" AS column2, \"300\" AS column3"
	assert.True(t, strings.Contains(mockAsSQl.Sql, expectedSQL1))
	assert.True(t, strings.Contains(mockAsSQl.Sql, expectedSQL2))
	assert.True(t, strings.Contains(mockAsSQl.Sql, expectedSQL3))
}
