package intent

import "encoding/json"

// Category represents the type of system operation.
type Category string

const (
	CategoryPackageManagement Category = "package_management"
	CategoryServiceManagement Category = "service_management"
	CategoryFileOperation     Category = "file_operation"
	CategoryNetwork           Category = "network"
	CategorySystemInfo        Category = "system_info"
	CategoryUnknown           Category = "unknown"
)

// ValidCategories is the set of allowed categories.
var ValidCategories = map[Category]bool{
	CategoryPackageManagement: true,
	CategoryServiceManagement: true,
	CategoryFileOperation:     true,
	CategoryNetwork:           true,
	CategorySystemInfo:        true,
	CategoryUnknown:           true,
}

// Action represents the specific operation to perform.
type Action string

const (
	ActionInstall   Action = "install"
	ActionRemove    Action = "remove"
	ActionStart     Action = "start"
	ActionStop      Action = "stop"
	ActionRestart   Action = "restart"
	ActionEnable    Action = "enable"
	ActionDisable   Action = "disable"
	ActionCreate    Action = "create"
	ActionDelete    Action = "delete"
	ActionList      Action = "list"
	ActionShow      Action = "show"
	ActionUpdate    Action = "update"
	ActionConfigure Action = "configure"
)

// ValidActions is the set of allowed actions.
var ValidActions = map[Action]bool{
	ActionInstall:   true,
	ActionRemove:    true,
	ActionStart:     true,
	ActionStop:      true,
	ActionRestart:   true,
	ActionEnable:    true,
	ActionDisable:   true,
	ActionCreate:    true,
	ActionDelete:    true,
	ActionList:      true,
	ActionShow:      true,
	ActionUpdate:    true,
	ActionConfigure: true,
}

// Parameters holds the extracted parameters from the user's intent.
type Parameters struct {
	Target  string          `json:"target"`
	Options FlexibleStrings `json:"options"`
	Scope   string          `json:"scope"`
}

// FlexibleStrings handles any format LLMs might return for array fields:
// ["string"], [{"key":"value"}], "single string", or null.
type FlexibleStrings []string

// UnmarshalJSON handles every format a model might return for an array.
func (fs *FlexibleStrings) UnmarshalJSON(data []byte) error {
	// Try []string first (ideal case)
	var strArr []string
	if err := json.Unmarshal(data, &strArr); err == nil {
		*fs = strArr
		return nil
	}

	// Try []object — extract all string values from each object
	var objArr []map[string]interface{}
	if err := json.Unmarshal(data, &objArr); err == nil {
		var result []string
		for _, obj := range objArr {
			for _, v := range obj {
				if s, ok := v.(string); ok && s != "" {
					result = append(result, s)
				}
			}
		}
		*fs = result
		return nil
	}

	// Try single string
	var single string
	if err := json.Unmarshal(data, &single); err == nil {
		if single != "" {
			*fs = []string{single}
		} else {
			*fs = []string{}
		}
		return nil
	}

	// null or anything else → empty
	*fs = []string{}
	return nil
}

// IntentResult is the structured output of the intent extraction stage.
type IntentResult struct {
	Intent            string          `json:"intent"`
	Category          Category        `json:"category"`
	Action            Action          `json:"action"`
	Parameters        Parameters      `json:"parameters"`
	Confidence        float64         `json:"confidence"`
	SuggestedCommands FlexibleStrings `json:"suggested_commands"`
	RawLLMResponse    string          `json:"-"`
	ExtractionTimeMs  int64           `json:"-"`
}
