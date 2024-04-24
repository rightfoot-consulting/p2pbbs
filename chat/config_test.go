/*
Copyright © 2024 Stephen Owens steve.owens@rightfoot.consulting
*/
package chat

import (
	"encoding/json"
	"testing"
)

func TestUnmarshalConfig(t *testing.T) {
	testJson := `
		{
			"rendezvous_string":"",
			"bootstrap_peers":[""],
			"listen_addresses":[""],
			"protocol_id": ""
		}
	`

	var config Configuration
	err := json.Unmarshal([]byte(testJson), &config)
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
		return
	}

}
