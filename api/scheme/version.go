package scheme

type APIVersion struct {
	min int16
	max int16
}

func (a APIVersion) Matches(version int16) bool {
	return a.min <= version && a.max >= version
}

func (a APIVersion) Min() int16 {
	return a.min
}

func (a APIVersion) Max() int16 {
	return a.max
}
