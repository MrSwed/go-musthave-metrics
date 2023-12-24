package repository

type Repository interface {
	MemStorage
}

func NewRepository() Repository {
	return NewMemRepository()
}
