package mttor

type Error struct {
	Descriptorion string
}

func (err *Error) Error() string {
	if err == nil {
		return "<nil>"
	}

	return "arcah/mttor: " + err.Descriptorion
}
