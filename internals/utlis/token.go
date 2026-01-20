package utlis

import "github.com/google/uuid"

func TokenGen() string {
	token, err := uuid.NewRandom()

	if err != nil {
		return ""
	}

	return token.String()

}
