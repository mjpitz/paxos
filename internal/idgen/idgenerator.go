package idgen

type IDGenerator interface {
	Next() (id uint64, err error)
}
