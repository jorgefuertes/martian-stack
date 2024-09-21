package server

type (
	Handler func(c Ctx) error
)
