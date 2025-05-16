package api

import (
	"vpod/internal/router"
)

func Routes(r *router.Router) {
	r.HandleFunc("PUT /gen", GenFeed)
}
