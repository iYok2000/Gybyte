package valueobject

import "errors"

// ErrInvalidStatus is returned when a Result is constructed or validated with a
// status that is not one of the two allowed wire values (Req 8.3).
var ErrInvalidStatus = errors.New("audit result status must be either 'pass' or 'fail'")

// Result is the outcome of a single Audit_Check. It carries the human-readable
// title of the check, its pass/fail status, and an explanatory message.
type Result struct {
	Title   string
	Status  Status
	Message string
}

// NewResult constructs a Result while enforcing the status invariant: the
// status MUST be either "pass" or "fail" (Req 8.3). An invalid status yields
// ErrInvalidStatus and a zero-value Result.
func NewResult(title string, status Status, message string) (Result, error) {
	r := Result{Title: title, Status: status, Message: message}
	if err := r.Validate(); err != nil {
		return Result{}, err
	}
	return r, nil
}

// Validate re-checks the status invariant before serialization, guarding
// against results assembled without going through NewResult (Req 8.3).
func (r Result) Validate() error {
	if !r.Status.IsValid() {
		return ErrInvalidStatus
	}
	return nil
}
