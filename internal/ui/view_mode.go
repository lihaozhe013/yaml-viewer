package ui

// ViewMode identifies the inspector density. Spacious is the currently
// implemented layout; Compact is reserved for a future density adjustment.
type ViewMode string

const (
	ViewModeSpacious ViewMode = "spacious"
	ViewModeCompact  ViewMode = "compact"
)
