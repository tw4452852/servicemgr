package main

type Type uint32

const (
	TypeBegin Type = iota

	TypeOpenMic    // 1
	TypeCloseMic   // 2
	TypeMicData    // 3
	TypeScanCode   // 4
	TypeOpenSound  // 5
	TypeCloseSound // 6
	TypeSoundData  // 7

	TypeEnd
)

const (
	ErrorBegin Type = (0xdead << 16) | iota

	ErrorInternal
	ErrorInvalidType
	ErrorConnectionGone
	ErrorSend

	ErrorEnd
)

func (t Type) String() string {
	switch t {
	// types
	case TypeOpenMic:
		return "TypeOpenMic"
	case TypeMicData:
		return "TypeMicData"
	case TypeCloseMic:
		return "TypeCloseMic"
	case TypeScanCode:
		return "TypeScanCode"
	case TypeOpenSound:
		return "TypeOpenSound"
	case TypeCloseSound:
		return "TypeCloseSound"
	case TypeSoundData:
		return "TypeSoundData"

	// errors
	case ErrorInternal:
		return "ErrorInternal"
	case ErrorInvalidType:
		return "ErrorInvalidType"
	case ErrorConnectionGone:
		return "ErrorConnectionGone"
	case ErrorSend:
		return "ErrorSend"

	default:
		return "unknown"
	}
}

func (t Type) IsValid() bool {
	return t.String() != "unknown"
}
