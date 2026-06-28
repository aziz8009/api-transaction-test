package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type validRequest struct {
	Name  string `json:"name" validate:"required"`
	Email string `json:"email" validate:"required,email"`
	Age   int    `json:"age" validate:"gte=18"`
}

func TestValidateStruct_Success(t *testing.T) {
	req := validRequest{
		Name:  "Aziz",
		Email: "aziz@mail.com",
		Age:   25,
	}

	err := ValidateStruct(req)

	require.NoError(t, err)
}

func TestValidateStruct_RequiredField(t *testing.T) {
	req := validRequest{
		Name:  "",
		Email: "aziz@mail.com",
		Age:   25,
	}

	err := ValidateStruct(req)

	require.Error(t, err)
	assert.Equal(t, "validation failed: name is required", err.Error())
}

func TestValidateStruct_EmailValidation(t *testing.T) {
	req := validRequest{
		Name:  "Aziz",
		Email: "invalid-email",
		Age:   25,
	}

	err := ValidateStruct(req)

	require.Error(t, err)
	assert.Equal(t, "validation failed: email is email", err.Error())
}

func TestValidateStruct_GTEValidation(t *testing.T) {
	req := validRequest{
		Name:  "Aziz",
		Email: "aziz@mail.com",
		Age:   17,
	}

	err := ValidateStruct(req)

	require.Error(t, err)
	assert.Equal(t, "validation failed: age is gte", err.Error())
}

func TestValidateStruct_MultipleErrors(t *testing.T) {
	req := validRequest{}

	err := ValidateStruct(req)

	require.Error(t, err)

	assert.Contains(t, err.Error(), "name is required")
	assert.Contains(t, err.Error(), "email is required")
	assert.Contains(t, err.Error(), "age is gte")
}
