package errific

import (
	"encoding/json"
	"strings"
	"testing"
)

// FuzzJSONSerialization tests JSON marshaling with random inputs
func FuzzJSONSerialization(f *testing.F) {
	Configure()
	
	// Add seed corpus
	f.Add("test error", "ERR_001", "tag1,tag2")
	f.Add("database error", "DB_TIMEOUT", "database,timeout,retryable")
	f.Add("", "NO_CODE", "")
	f.Add("special chars: \"quotes\" \n newlines \t tabs", "SPECIAL", "tag")
	
	f.Fuzz(func(t *testing.T, errMsg, code, tags string) {
		var ErrTest = Err(errMsg)
		
		tagSlice := []string{}
		if tags != "" {
			tagSlice = strings.Split(tags, ",")
		}
		
		err := ErrTest.New().
			WithCode(code).
			WithTags(tagSlice...)
		
		// Should never panic
		data, jsonErr := json.Marshal(err)
		if jsonErr != nil {
			t.Errorf("marshal failed: %v", jsonErr)
			return
		}
		
		// Should be valid JSON
		var decoded map[string]interface{}
		if unmarshalErr := json.Unmarshal(data, &decoded); unmarshalErr != nil {
			t.Errorf("unmarshal failed: %v", unmarshalErr)
			return
		}
		
		// Basic sanity checks
		if _, ok := decoded["error"]; !ok {
			t.Error("decoded JSON should have 'error' field")
		}
	})
}

// FuzzMCPErrorConversion tests MCP error conversion with random inputs
func FuzzMCPErrorConversion(f *testing.F) {
	Configure()
	
	// Add seed corpus
	f.Add("test error", -32603)
	f.Add("parse error", -32700)
	f.Add("tool error", -32000)
	f.Add("", 0)
	
	f.Fuzz(func(t *testing.T, msg string, code int) {
		var ErrTest = Err(msg)
		err := ErrTest.New().WithMCPCode(code)
		
		// Should never panic
		mcpErr := ToMCPError(err)
		
		// Should be JSON serializable
		data, jsonErr := json.Marshal(mcpErr)
		if jsonErr != nil {
			t.Errorf("MCP marshal failed: %v", jsonErr)
			return
		}
		
		// Should be valid JSON
		var decoded map[string]interface{}
		if unmarshalErr := json.Unmarshal(data, &decoded); unmarshalErr != nil {
			t.Errorf("unmarshal failed: %v", unmarshalErr)
			return
		}
		
		// MCP errors should have code and message
		if _, ok := decoded["code"]; !ok {
			t.Error("MCP error should have 'code' field")
		}
		if _, ok := decoded["message"]; !ok {
			t.Error("MCP error should have 'message' field")
		}
	})
}

// FuzzErrorWithContext tests context handling with random inputs
func FuzzErrorWithContext(f *testing.F) {
	Configure()
	
	// Add seed corpus
	f.Add("error", "key1", "value1", "key2", "value2")
	f.Add("test", "", "", "", "")
	f.Add("special", "k\"ey", "val\"ue", "k\ney", "val\nue")
	
	f.Fuzz(func(t *testing.T, errMsg, k1, v1, k2, v2 string) {
		var ErrTest = Err(errMsg)
		
		ctx := Context{}
		if k1 != "" {
			ctx[k1] = v1
		}
		if k2 != "" {
			ctx[k2] = v2
		}
		
		err := ErrTest.New().WithContext(ctx)
		
		// Should never panic
		_ = err.Error()
		
		// Should be JSON serializable
		data, jsonErr := json.Marshal(err)
		if jsonErr != nil {
			t.Errorf("marshal failed: %v", jsonErr)
			return
		}
		
		// Should be valid JSON
		var decoded map[string]interface{}
		if unmarshalErr := json.Unmarshal(data, &decoded); unmarshalErr != nil {
			t.Errorf("unmarshal failed: %v", unmarshalErr)
		}
	})
}
