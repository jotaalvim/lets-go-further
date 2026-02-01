// Package validator server para guardar erros
package validator

import (
	"regexp"
	"slices"
)

var (
	EmailRX = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
)

type Validator struct {
	Errors map[string]string
}

func New() *Validator {
	return &Validator{make(map[string]string)}
}

func (v *Validator) Valid() bool {
	return len(v.Errors) == 0
}

func (v *Validator) AddError(key string, message string) {
	_, exists := v.Errors[key]

	if !exists {
		v.Errors[key] = message
	}
}

// Check adds the error if the condition is false
func (v *Validator) Check(ok bool, key string, message string) {
	if !ok {
		v.AddError(key, message)
	}
}

// PermittedValue true if a value is in a specefic list
func PermittedValue[T comparable](value T, permitted ...T) bool {
	return slices.Contains(permitted, value)
}

// Matches and returns true if regex match
func Matches(value string, rx *regexp.Regexp) bool {
	return rx.MatchString(value)
}

// UniqueValues  checks if all elements are unique
func UniqueValues[T comparable](values []T) bool {

	uniqueValues := make(map[T]bool)

	for _, value := range values {
		uniqueValues[value] = true
	}
	return len(uniqueValues) == len(values)
}
