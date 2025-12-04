package domain

import "errors"

var (
	ErrInvalidZipcode  = errors.New("invalid zipcode")
	ErrZipcodeNotFound = errors.New("can not find zipcode")
	ErrWeatherNotFound = errors.New("weather not found")
	ErrInvalidLocation = errors.New("invalid location")
)
