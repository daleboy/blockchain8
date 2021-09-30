package blockchain8

import "log"

func Handle(err error) {

	if err != nil {
		log.Panic(err)
	}
}
