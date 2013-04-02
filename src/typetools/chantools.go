package typetools

func ChanDebug(c interface{})

func ChanRead(c interface{}, minnum int) (data uintptr, length int, size uint32)
