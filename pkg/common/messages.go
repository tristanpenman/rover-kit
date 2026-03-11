package common

type CommandType string

const (
	CommandForwards  CommandType = "forwards"
	CommandBackwards CommandType = "backwards"
	CommandSpinCW    CommandType = "spin_cw"
	CommandSpinCCW   CommandType = "spin_ccw"
	CommandStop      CommandType = "stop"
	CommandThrottle  CommandType = "throttle"
)

type ForwardsCommand struct {
	Type CommandType `json:"type"`
}

type BackwardsCommand struct {
	Type CommandType `json:"type"`
}

type SpinCWCommand struct {
	Type CommandType `json:"type"`
}

type SpinCCWCommand struct {
	Type CommandType `json:"type"`
}

type StopCommand struct {
	Type CommandType `json:"type"`
}

type ThrottleCommand struct {
	Type  CommandType `json:"type"`
	Value float64     `json:"value"`
}
