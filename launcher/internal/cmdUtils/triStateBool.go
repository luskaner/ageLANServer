package cmdUtils

type TriStateBool int

const (
	triStateUnset TriStateBool = iota
	triStateTrue
	triStateFalse
)

func (t *TriStateBool) Set(state bool) {
	switch state {
	case true:
		*t = triStateTrue
	case false:
		*t = triStateFalse
	}
}

func (t *TriStateBool) Unset() bool {
	return *t == triStateUnset
}

func (t *TriStateBool) False() bool {
	return *t == triStateFalse
}

func (t *TriStateBool) True() bool {
	return *t == triStateTrue
}
