package pipeline

import "fmt"

// Error codes for pipeline failures
const (
	ErrOllamaConn     = "E_OLLAMA_CONN"
	ErrIntentParse     = "E_INTENT_PARSE"
	ErrIntentUnknown   = "E_INTENT_UNKNOWN"
	ErrContextOS       = "E_CONTEXT_OS"
	ErrContextPkg      = "E_CONTEXT_PKG"
	ErrPlanUnsupported = "E_PLAN_UNSUPPORTED"
	ErrValidDeps       = "E_VALID_DEPS"
	ErrValidDisk       = "E_VALID_DISK"
	ErrValidPriv       = "E_VALID_PRIV"
	ErrValidRisk       = "E_VALID_RISK"
	ErrUserAbort       = "E_USER_ABORT"
	ErrExecFail        = "E_EXEC_FAIL"
	ErrExecTimeout     = "E_EXEC_TIMEOUT"
	ErrRecoveryFail    = "E_RECOVERY_FAIL"
)

// PipelineError represents a typed error from any pipeline stage.
type PipelineError struct {
	Code    string
	Stage   string
	Message string
	Cause   error
}

func (e *PipelineError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%s] %s: %s (cause: %v)", e.Code, e.Stage, e.Message, e.Cause)
	}
	return fmt.Sprintf("[%s] %s: %s", e.Code, e.Stage, e.Message)
}

func (e *PipelineError) Unwrap() error {
	return e.Cause
}

// NewPipelineError creates a new PipelineError.
func NewPipelineError(code, stage, message string, cause error) *PipelineError {
	return &PipelineError{
		Code:    code,
		Stage:   stage,
		Message: message,
		Cause:   cause,
	}
}
