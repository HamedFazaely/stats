package provider

import "testing"

func TestGetStats(t *testing.T) {
	url := "http://localhost:9090/v1/status/hamed"
	id := 7
	resp, err := GetOrderStatusFromProvider(url, uint64(id))
	if err != nil {
		t.Fatal(err)
	}
	for _, x := range resp.Data {
		t.Log(x)
	}
}
