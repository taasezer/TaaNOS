package intent

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
	Target  string   `json:"target"`
	Options []string `json:"options"`
	Scope   string   `json:"scope"`
}

// IntentResult is the structured output of the intent extraction stage.
type IntentResult struct {
	Intent            string     `json:"intent"`
	Category          Category   `json:"category"`
	Action            Action     `json:"action"`
	Parameters        Parameters `json:"parameters"`
	Confidence        float64    `json:"confidence"`
	SuggestedCommands []string   `json:"suggested_commands"`
	RawLLMResponse    string     `json:"-"`
	ExtractionTimeMs  int64      `json:"-"`
}
